package step

import (
	"fmt"
	"log"

	"github.com/dingotiles/dingo-postgresql-broker/config"
)

// Step is a step in a workflow to change a cluster (grow, scale, move)
type Step interface {
	StepType() string
	Perform(backends []*config.Backend) error
}

func debug(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		log.Fatalf("%s\n\n", err)
	}
}
