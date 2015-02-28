package riak_cs_service

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"fmt"
	"time"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/runner"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
)

var _ = Describe("Riak CS Service Lifecycle", func() {

	var (
		shortTimeout, longTimeout, startTimeout time.Duration
	)

	BeforeEach(func() {
		shortTimeout = ScaledTimeout(10*time.Second)
		longTimeout = ScaledTimeout(60*time.Second)
		startTimeout = ScaledTimeout(5*time.Minute)

		AppName = RandomName()

		ExecWithTimeout(Cf("push", AppName, "-m", "256M", "-p", sinatraPath, "-no-start"), longTimeout)
	})

	AfterEach(func() {
		ExecWithTimeout(Cf("delete", AppName, "-f"), longTimeout)
	})

	It("Allows users to create, bind, write to, read from, unbind, and destroy the service instance", func() {
		ServiceName := ServiceName()
		PlanName := PlanName()
		ServiceInstanceName := RandomName()

		ExecWithTimeout(Cf("create-service", ServiceName, PlanName, ServiceInstanceName), longTimeout)
		ExecWithTimeout(Cf("bind-service", AppName, ServiceInstanceName), longTimeout)
		ExecWithTimeout(Cf("start", AppName), startTimeout)

		uri := AppUri(AppName) + "/service/blobstore/" + ServiceInstanceName + "/mykey"
		delete_uri := AppUri(AppName) + "/service/blobstore/" + ServiceInstanceName

		fmt.Println("Posting to url: ", uri)
		Expect(ExecWithTimeout(Curl("-k", "-d", "myvalue", uri), shortTimeout)).To(Say("myvalue"))
		fmt.Println("\n")

		fmt.Println("Curling url: ", uri)
		Expect(ExecWithTimeout(Curl("-k", uri), shortTimeout)).To(Say("myvalue"))
		fmt.Println("\n")

		fmt.Println("Sending delete to: ", delete_uri)
		Expect(ExecWithTimeout(Curl("-X", "DELETE", "-k", delete_uri), shortTimeout)).To(Say("successfully_deleted"))
		fmt.Println("\n")

		ExecWithTimeout(Cf("unbind-service", AppName, ServiceInstanceName), longTimeout)
		ExecWithTimeout(Cf("delete-service", "-f", ServiceInstanceName), longTimeout)
	})
})
