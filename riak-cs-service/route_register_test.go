package riak_cs_service

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Riak CS Nodes Register a Route", func() {
	It("Allows users to access the riak-cs service using external url instead of IP of single machine after register the route", func() {
		endpointURL := TestConfig.RiakCsScheme + TestConfig.RiakCsHost + "/riak-cs/ping"

		runner.NewCmdRunner(runner.Curl("-k", endpointURL), TestContext.ShortTimeout()).WithOutput("OK").Run()
	})
})

var _ = Describe("Riak Broker Registers a Route", func() {
	It("Allows users to access the riak-cs broker using a url", func() {
		endpointURL := "http://" + TestConfig.BrokerHost + "/v2/catalog"

		// check for 401 because it means we reached the endpoint, but did not supply credentials.
		// a failure would be a 404
		runner.NewCmdRunner(runner.Curl("-k", "-s", "-w", "%{http_code}", endpointURL, "-o", "/dev/null"), TestContext.ShortTimeout()).WithOutput("401").Run()
	})
})
