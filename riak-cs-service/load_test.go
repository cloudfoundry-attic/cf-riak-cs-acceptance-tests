package riak_cs_service

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "fmt"
    "math"
    "math/rand"
    "sync"
    "time"
    "encoding/json"

    . "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
    . "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
    . "github.com/cloudfoundry-incubator/cf-test-helpers/runner"
    . "github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"

    "github.com/cloudfoundry-incubator/riakcs-acceptance-tests/helpers"
    "github.com/cloudfoundry-incubator/riakcs-acceptance-tests/helpers/s3"
)

var _ = Describe("Riak CS Service Load Tests", func() {

    var (
        shortTimeout time.Duration
        longTimeout  time.Duration
        appName      string
    )

    BeforeEach(func() {
        shortTimeout = ScaledTimeout(10 * time.Second)
        longTimeout = ScaledTimeout(5 * time.Minute)

        appName = RandomName()

        ExecWithTimeout(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), longTimeout)
    })

    AfterEach(func() {
        ExecWithTimeout(Cf("delete", appName, "-f"), longTimeout)
    })

    var batch = func(
    keys []string,
    values map[string]string,
    threads int,
    onEach func(key, value string),
    ) {
        size := len(keys)
        batchSize := int(math.Ceil(float64(size) / float64(threads)))

        startTime := time.Now()

        var wg sync.WaitGroup

        for i := 0; i < threads; i++ {
            wg.Add(1)
            go func(threadID int) {
                defer GinkgoRecover()
                defer wg.Done()
                startIndex := threadID * batchSize
                endIndex := (threadID + 1) * batchSize
                endIndex = int(math.Min(float64(endIndex), float64(size)))
                fmt.Printf("Batch %d [%d:%d]\n", threadID, startIndex, endIndex)
                for _, key := range keys[startIndex:endIndex] {
                    onEach(key, values[key])
                }
                fmt.Printf("Batch %d Complete\n", threadID)
            }(i)
        }

        wg.Wait()

        stopTime := time.Now()
        elapsed := stopTime.Sub(startTime)
        transactionsPerSecond := float64(size) / elapsed.Seconds()
        fmt.Printf("Total Time: %v (%.2f tps)", elapsed, transactionsPerSecond)
    }

    var generatePairs = func(count int) ([]string, map[string]string) {
        keys := make([]string, count)
        for i := 0; i < count; i++ {
            keys[i] = fmt.Sprintf("key-%d", rand.Int63())
        }

        values := make(map[string]string, count)
        for _, key := range keys {
            values[key] = fmt.Sprintf("value-%d", rand.Int63())
        }

        return keys, values
    }

    Context("When serivce is bound to an app", func() {
        var (
            serviceInstanceName string
        )

        BeforeEach(func() {
            serviceName := ServiceName()
            planName := PlanName()
            serviceInstanceName = RandomName()

            ExecWithTimeout(Cf("create-service", serviceName, planName, serviceInstanceName), longTimeout)
            ExecWithTimeout(Cf("bind-service", appName, serviceInstanceName), longTimeout)
            ExecWithTimeout(Cf("start", appName), longTimeout)
        })

        AfterEach(func() {
            ExecWithTimeout(Cf("unbind-service", appName, serviceInstanceName), longTimeout)
            ExecWithTimeout(Cf("delete-service", "-f", serviceInstanceName), longTimeout)
        })

        var newS3Client = func() (client *s3.Client, bucket string) {
            envURI := fmt.Sprintf("%s/env", AppUri(appName))
            cmd := Curl("-k", envURI)
            ExecWithTimeout(cmd, shortTimeout)

            var env helpers.VCAPServices
            json.Unmarshal(cmd.Out.Contents(), &env)

            Expect(env["p-riakcs"]).To(HaveLen(1))
            appEnv := env["p-riakcs"][0]

            config := s3.ClientConfig{
                S3curl: "/Users/pivotal/workspace/s3curl/s3curl.pl",
                Host: appEnv.Credentials.Host(),
                AccessKey: appEnv.Credentials.AccessKeyID,
                SecretKey: appEnv.Credentials.SecretAccessKey,
                Schema: RiakCSIntegrationConfig.RiakCsScheme,
            }

            return &s3.Client{Config: config}, appEnv.Credentials.Bucket()
        }

        FIt("App can write to and then read from riak-cs in parallel under load", func() {
            s3Client, bucket := newS3Client()

            numKeys := 100

            keys, values := generatePairs(numKeys)

            threads := 10

            batch(keys, values, threads, func(key, value string) {
                err := s3Client.Put(bucket, key, []byte(value))
                Expect(err).ToNot(HaveOccurred())
            })

            fmt.Println("\n")

            threads = 10

            batch(keys, values, threads, func(key, value string) {
                bytes, err := s3Client.Get(bucket, key)
                Expect(err).ToNot(HaveOccurred())
                Expect(string(bytes)).To(Equal(value))
            })

            fmt.Println("\n")
        })

    })
})
