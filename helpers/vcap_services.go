package helpers

import (
    "regexp"
    "fmt"
)

// VCAPServices is a map of service names to service details.
// It is provided to CF Apps via the VCAP_SERVICES environment variable.
type VCAPServices map[string][]Service

type Service struct {
    Name string             `json:"name"`
    Label string            `json:"label"`
    Tags []string           `json:"tags"`
    Plan string             `json:"plan"`
    Credentials Credentials `json:"credentials"`
}

type Credentials struct {
    URI             string `json:"uri"`
    AccessKeyID     string `json:"access_key_id"`
    SecretAccessKey string `json:"secret_access_key"`
}

func (creds *Credentials) Bucket() string {
    exp := regexp.MustCompile("https://.*:.*@.*/(.*)")
    matches := exp.FindStringSubmatch(creds.URI)
    if matches == nil {
        panic(fmt.Sprintf("Invalid URI pattern: '%s'", creds.URI))
    }
    return matches[1]
}

func (creds *Credentials) Host() string {
    exp := regexp.MustCompile("https://.*:.*@(.*)/.*")
    matches := exp.FindStringSubmatch(creds.URI)
    if matches == nil {
        panic(fmt.Sprintf("Invalid URI pattern: '%s'", creds.URI))
    }
    return matches[1]
}
