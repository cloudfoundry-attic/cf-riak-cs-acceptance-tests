package riak_cs_service

import (
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Riak CS Nodes Register a Route", func() {
	It("Allows users to access the riak-cs service using external url instead of IP of single machine after register the route", func() {
		endpointURL := RiakCSIntegrationConfig.RiakCsScheme + RiakCSIntegrationConfig.RiakCsHost + "/riak-cs/ping"

		runner.NewCmdRunner(runner.Curl("-k", endpointURL), ScaledTimeout(60*time.Second)).WithOutput("OK").Run()
	})
})

var _ = Describe("Riak Broker Registers a Route", func() {
	It("Allows users to access the riak-cs broker using a url", func() {
		endpointURL := "http://" + RiakCSIntegrationConfig.BrokerHost + "/v2/catalog"

		// check for 401 because it means we reached the endpoint, but did not supply credentials.
		// a failure would be a 404
		runner.NewCmdRunner(runner.Curl("-k", "-s", "-w", "%{http_code}", endpointURL, "-o", "/dev/null"), ScaledTimeout(60*time.Second)).WithOutput("401").Run()
	})
})
