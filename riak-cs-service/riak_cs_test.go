package riak_cs_service

import (
	. "github.com/onsi/ginkgo"

	"fmt"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
)

var _ = Describe("Riak CS Service Lifecycle", func() {

	var (
		appName     string
		sinatraPath = "../assets/app_sinatra_service"
	)

	BeforeEach(func() {
		appName = RandomName()

		runner.NewCmdRunner(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), TestContext.LongTimeout()).Run()
	})

	AfterEach(func() {
		runner.NewCmdRunner(Cf("delete", appName, "-f"), TestContext.LongTimeout()).Run()
	})

	It("Allows users to create, bind, write to, read from, unbind, and destroy the service instance", func() {
		serviceName := TestConfig.ServiceName
		planName := TestConfig.PlanName
		ServiceInstanceName := RandomName()

		runner.NewCmdRunner(Cf("create-service", serviceName, planName, ServiceInstanceName), TestContext.LongTimeout()).Run()
		runner.NewCmdRunner(Cf("bind-service", appName, ServiceInstanceName), TestContext.LongTimeout()).Run()
		runner.NewCmdRunner(Cf("start", appName), TestContext.LongTimeout()).Run()

		uri := TestConfig.AppURI(appName) + "/service/blobstore/" + ServiceInstanceName + "/mykey"
		delete_uri := TestConfig.AppURI(appName) + "/service/blobstore/" + ServiceInstanceName

		fmt.Println("Posting to url: ", uri)
		runner.NewCmdRunner(runner.Curl("-k", "-d", "myvalue", uri), TestContext.ShortTimeout()).WithOutput("myvalue").Run()
		fmt.Println("\n")

		fmt.Println("Curling url: ", uri)
		runner.NewCmdRunner(runner.Curl("-k", uri), TestContext.ShortTimeout()).WithOutput("myvalue").Run()
		fmt.Println("\n")

		fmt.Println("Sending delete to: ", delete_uri)
		runner.NewCmdRunner(runner.Curl("-X", "DELETE", "-k", delete_uri), TestContext.ShortTimeout()).WithOutput("successfully_deleted").Run()
		fmt.Println("\n")

		runner.NewCmdRunner(Cf("unbind-service", appName, ServiceInstanceName), TestContext.LongTimeout()).Run()
		runner.NewCmdRunner(Cf("delete-service", "-f", ServiceInstanceName), TestContext.LongTimeout()).Run()
	})
})
