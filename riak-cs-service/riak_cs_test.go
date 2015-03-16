package riak_cs_service

import (
	. "github.com/onsi/ginkgo"

	"fmt"
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
)

var _ = Describe("Riak CS Service Lifecycle", func() {

	var (
		shortTimeout, longTimeout, startTimeout time.Duration
	)

	BeforeEach(func() {
		shortTimeout = ScaledTimeout(30 * time.Second)
		longTimeout = ScaledTimeout(60 * time.Second)
		startTimeout = ScaledTimeout(5 * time.Minute)

		AppName = RandomName()

		runner.NewCmdRunner(Cf("push", AppName, "-m", "256M", "-p", sinatraPath, "-no-start"), longTimeout).Run()
	})

	AfterEach(func() {
		runner.NewCmdRunner(Cf("delete", AppName, "-f"), longTimeout).Run()
	})

	It("Allows users to create, bind, write to, read from, unbind, and destroy the service instance", func() {
		ServiceName := ServiceName()
		PlanName := PlanName()
		ServiceInstanceName := RandomName()

		runner.NewCmdRunner(Cf("create-service", ServiceName, PlanName, ServiceInstanceName), longTimeout).Run()
		runner.NewCmdRunner(Cf("bind-service", AppName, ServiceInstanceName), longTimeout).Run()
		runner.NewCmdRunner(Cf("start", AppName), startTimeout).Run()

		uri := AppUri(AppName) + "/service/blobstore/" + ServiceInstanceName + "/mykey"
		delete_uri := AppUri(AppName) + "/service/blobstore/" + ServiceInstanceName

		fmt.Println("Posting to url: ", uri)
		runner.NewCmdRunner(runner.Curl("-k", "-d", "myvalue", uri), shortTimeout).WithOutput("myvalue").Run()
		fmt.Println("\n")

		fmt.Println("Curling url: ", uri)
		runner.NewCmdRunner(runner.Curl("-k", uri), shortTimeout).WithOutput("myvalue").Run()
		fmt.Println("\n")

		fmt.Println("Sending delete to: ", delete_uri)
		runner.NewCmdRunner(runner.Curl("-X", "DELETE", "-k", delete_uri), shortTimeout).WithOutput("successfully_deleted").Run()
		fmt.Println("\n")

		runner.NewCmdRunner(Cf("unbind-service", AppName, ServiceInstanceName), longTimeout).Run()
		runner.NewCmdRunner(Cf("delete-service", "-f", ServiceInstanceName), longTimeout).Run()
	})
})
