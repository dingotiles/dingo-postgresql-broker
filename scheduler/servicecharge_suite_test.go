package scheduler_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Testscheduler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Instance Changes Suite")
}
