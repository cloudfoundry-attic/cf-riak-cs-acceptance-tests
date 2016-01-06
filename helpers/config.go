package helpers

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/services"
)

type RiakCSIntegrationConfig struct {
	services.Config

	RiakCsHost     string `json:"riak_cs_host"`
	RiakCsScheme   string `json:"riak_cs_scheme"`
	ServiceName    string `json:"service_name"`
	PlanName       string `json:"plan_name"`
	BrokerHost     string `json:"broker_host"`
	BrokerProtocol string `json:"broker_protocol"`
}

func (c RiakCSIntegrationConfig) AppURI(appname string) string {
	return c.RiakCsScheme + appname + "." + c.AppsDomain
}

func LoadConfig() (RiakCSIntegrationConfig, error) {
	config := RiakCSIntegrationConfig{}

	path := os.Getenv("CONFIG")
	if path == "" {
		return config, fmt.Errorf("Must set $CONFIG to point to an integration config .json file.")
	}

	err := services.LoadConfig(path, &config)
	if err != nil {
		return config, fmt.Errorf("Loading config: %s", err.Error())
	}

	return config, nil
}

func ValidateConfig(config *RiakCSIntegrationConfig) error {
	err := services.ValidateConfig(&config.Config)
	if err != nil {
		return err
	}

	if config.ServiceName == "" {
		return fmt.Errorf("Field 'service_name' must not be empty")
	}

	if config.PlanName == "" {
		return fmt.Errorf("Field 'plan_name' must not be empty")
	}

	if config.BrokerHost == "" {
		return fmt.Errorf("Field 'broker_host' must not be empty")
	}

	if config.RiakCsHost == "" {
		return fmt.Errorf("Field 'riak_cs_host' must not be empty")
	}

	if config.BrokerProtocol == "" {
		config.BrokerProtocol = "https"
	}

	return nil
}
