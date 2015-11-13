package serviceinstance_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestServiceInstance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Instance Suite")
}
