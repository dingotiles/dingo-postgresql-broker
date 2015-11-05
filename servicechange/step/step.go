package step

import (
	"fmt"
	"log"

	"github.com/pivotal-golang/lager"
)

// Step is a step in a workflow to change a cluster (grow, scale, move)
type Step interface {
	Perform(lager.Logger) error
}

func debug(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		log.Fatalf("%s\n\n", err)
	}
}
