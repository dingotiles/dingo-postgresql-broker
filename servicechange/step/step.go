package step

import "github.com/pivotal-golang/lager"

// Step is a step in a workflow to change a cluster (grow, scale, move)
type Step interface {
	Perform(lager.Logger)
}
