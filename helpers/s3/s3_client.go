package s3

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "io/ioutil"
    "bytes"
    "syscall"
    "strings"
    "strconv"
)

type ClientConfig struct {
    S3curl string
    Host string
    AccessKey string
    SecretKey string
    Schema string
}

type Client struct {
    Config ClientConfig
}

func (c *Client) Put(bucket, object string, bytes []byte) error {
    objectURI := fmt.Sprintf("%s%s/%s/%s", c.Config.Schema, c.Config.Host, bucket, object)

    tempFile, err := ioutil.TempFile("", "raik-cs-get")
    if err != nil {
        return fmt.Errorf("Creating temp file: %s", err.Error())
    }
    tempFilePath := tempFile.Name()
    defer os.RemoveAll(tempFilePath)

    _, err = tempFile.Write(bytes)
    if err != nil {
        return fmt.Errorf("Writing to temp file: %s", err.Error())
    }

    stdout, stderr, exitStatus, err := c.exec(
        "--id=admin",
        fmt.Sprintf("--put=%s", tempFilePath),
        "--",
        "--insecure",
        "--silent",
        "--output", "/dev/stderr",
        "--write-out", "%{http_code}",
        objectURI,
    )
    if err != nil {
        //TODO: print stdout & stderr better
        fmt.Printf("[STDOUT]\n%s\n\n", stdout)
        fmt.Printf("[STDERR]\n%s\n\n", stderr)
        return fmt.Errorf("Putting object (exit %d): %s", exitStatus, err.Error())
    }

    statusCode, err := strconv.Atoi(stdout)
    if err != nil {
        fmt.Printf("[STDOUT]\n%s\n\n", stdout)
        fmt.Printf("[STDERR]\n%s\n\n", stderr)
        return fmt.Errorf("Invalid status code '%s': %s", stdout, err.Error())
    }
    if statusCode < 200 || statusCode >= 300 {
        fmt.Printf("[STDOUT]\n%s\n\n", stdout)
        fmt.Printf("[STDERR]\n%s\n\n", stderr)
        return fmt.Errorf("PUT Request failed (status code %d): %s", statusCode, err.Error())
    }

    return nil
}

func (c *Client) Get(bucket, object string) ([]byte, error) {
    objectURI := fmt.Sprintf("%s%s/%s/%s", c.Config.Schema, c.Config.Host, bucket, object)

    stdout, stderr, exitStatus, err := c.exec(
        "--id=admin",
        "--",
        "--insecure",
        "--silent",
        "--output", "/dev/stderr",
        "--write-out", "%{http_code}",
        objectURI,
    )
    if err != nil {
        //TODO: print stdout & stderr better
        fmt.Printf("[STDOUT]\n%s\n\n", stdout)
        fmt.Printf("[STDERR]\n%s\n\n", stderr)
        return nil, fmt.Errorf("Getting object (exit %d): %s", exitStatus, err.Error())
    }

    statusCode, err := strconv.Atoi(stdout)
    if err != nil {
        fmt.Printf("[STDOUT]\n%s\n\n", stdout)
        fmt.Printf("[STDERR]\n%s\n\n", stderr)
        return nil, fmt.Errorf("Invalid status code '%s': %s", stdout, err.Error())
    }
    if statusCode < 200 || statusCode >= 300 {
        fmt.Printf("[STDOUT]\n%s\n\n", stdout)
        fmt.Printf("[STDERR]\n%s\n\n", stderr)
        return nil, fmt.Errorf("GET Request failed (status code %d): %s", statusCode, err.Error())
    }

    return []byte(stderr), nil
}

func (c *Client) exec(args ...string) (stdout, stderr string, exitStatus int, err error) {
    exitStatus = -1

    configDirPath, err := ioutil.TempDir("", "raik-cs-s3cfg")
    if err != nil {
        return "", "", exitStatus, fmt.Errorf("Creating temp dir: %s", err.Error())
    }
    defer os.RemoveAll(configDirPath)

    configFilePath := fmt.Sprintf("%s/.s3curl", configDirPath)

    err = c.writeConfig(configFilePath)
    if err != nil {
        return "", "", exitStatus, fmt.Errorf("Writing config file: %s", err.Error())
    }

    workingDirPath, err := ioutil.TempDir("", "raik-cs-working-dir")
    if err != nil {
        return "", "", exitStatus, fmt.Errorf("Creating temp dir: %s", err.Error())
    }

    s3curl := c.Config.S3curl

    stdoutBuffer := bytes.NewBufferString("")
    stderrBuffer := bytes.NewBufferString("")

    cmd := exec.Command(s3curl, args...)
    cmd.Dir = workingDirPath
    cmd.Env = []string{
        fmt.Sprintf("HOME=%s", configDirPath),
    } // OR os.Environ()
    cmd.Stdout = stdoutBuffer
    cmd.Stderr = stderrBuffer

    cmdString := strings.Join(append([]string{s3curl}, args...), " ")

    fmt.Printf("\nExecuting Command: %s\n", cmdString)

    err = cmd.Start()
    if err != nil {
        return "", "", exitStatus, fmt.Errorf("Starting command %s: %s", cmdString, err.Error())
    }

    err = cmd.Wait()

    stdout = string(stdoutBuffer.Bytes())
    stderr = string(stderrBuffer.Bytes())

    waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
    if waitStatus.Exited() {
        exitStatus = waitStatus.ExitStatus()
    } else if waitStatus.Signaled() {
        exitStatus = 128 + int(waitStatus.Signal())
    }

    if err != nil {
        return stdout, stderr, exitStatus, fmt.Errorf("Command failed %s: %s", cmdString, err.Error())
    }

    return stdout, stderr, exitStatus, nil
}

func (c *Client) writeConfig(path string) error {
    configString := `%%awsSecretAccessKeys = (
    admin => {
        id => '%s',
        key => '%s'
    }
);`

    configString = fmt.Sprintf(
        configString,
        c.Config.AccessKey,
        c.Config.SecretKey,
    )

    //fmt.Printf("\n[S3CURL CONFIG]\n%s\n\n", configString)

    err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
    if err != nil {
        return fmt.Errorf("Creating dir to write file: %s", err.Error())
    }

    file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        return fmt.Errorf("Creating file %s: %s", path, err.Error())
    }
    defer file.Close()

    _, err = file.Write([]byte(configString))
    if err != nil {
        return fmt.Errorf("Writing content to file %s: %s", path, err.Error())
    }

    return nil
}
