package servicechange_test

import (
	"github.com/cloudfoundry-community/patroni-broker/servicechange"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service instance changes", func() {
	Describe(".Steps", func() {
		It("noop", func() {
			req := servicechange.NewRequest()
			Ω(req.Steps()).To(Equal([]servicechange.Step{}))
		})
	})
})
