package riak_cs_service

import (
	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"fmt"
	"testing"

	"github.com/cloudfoundry-incubator/cf-riak-cs-acceptance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/services"
)

var TestConfig helpers.RiakCSIntegrationConfig
var TestContext services.Context

func TestServices(t *testing.T) {
	var err error
	TestConfig, err = helpers.LoadConfig()
	if err != nil {
		panic("Loading config: " + err.Error())
	}

	err = helpers.ValidateConfig(&TestConfig)
	if err != nil {
		panic("Validating config: " + err.Error())
	}

	TestContext = services.NewContext(TestConfig.Config, "RiakCSATS")

	BeforeEach(TestContext.Setup)
	AfterEach(TestContext.Teardown)

	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("junit_%d.xml", ginkgoconfig.GinkgoConfig.ParallelNode))
	RunSpecsWithDefaultAndCustomReporters(t, "RiakCS Acceptance Tests", []Reporter{junitReporter})
}
