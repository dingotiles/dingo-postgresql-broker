package step

import (
	"fmt"
	"log"
)

// Step is a step in a workflow to change a cluster (grow, scale, move)
type Step interface {
	StepType() string
	Perform() error
}

func debug(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		log.Fatalf("%s\n\n", err)
	}
}
