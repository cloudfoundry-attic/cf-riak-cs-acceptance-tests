package context_setup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestContextSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ContextSetup Suite")
}
