package scheduler

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/step"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

const (
	defaultNodeSize = 20
)

// p.est represents a user-originating p.est to change a service instance (grow, scale, move)
type plan struct {
	clusterState      *structs.ClusterState
	newFeatures       structs.ClusterFeatures
	availableBackends backend.Backends
	allBackends       backend.Backends
	newNodeSize       int
	logger            lager.Logger
}

// Newp.est creates a p.est to change a service instance
func (s *Scheduler) newPlan(cluster *structs.ClusterState, features structs.ClusterFeatures) (plan, error) {
	backends, err := s.FilterCellsByGUIDs(features.CellGUIDs)
	if err != nil {
		return plan{}, err
	}
	return plan{
		clusterState:      cluster,
		newFeatures:       features,
		availableBackends: backends,
		allBackends:       s.backends,
		logger:            s.logger,
		newNodeSize:       defaultNodeSize,
	}, nil
}

// stepTypes is the ordered sequence of workflow step types to orchestrate a service instance change
func (p plan) stepTypes() []string {
	steps := p.steps()
	stepTypes := make([]string, len(steps))
	for i, step := range steps {
		stepTypes[i] = step.StepType()
	}
	return stepTypes
}

// steps is the ordered sequence of workflow steps to orchestrate a service instance change
func (p plan) steps() (steps []step.Step) {
	for i := 0; i < p.clusterGrowingBy(); i++ {
		steps = append(steps, step.NewStepAddNode(p.clusterState, p.availableBackends, p.logger))
	}

	nodesToBeReplaced := p.nodesToBeReplaced()
	for _ = range nodesToBeReplaced {
		steps = append(steps, step.NewStepAddNode(p.clusterState, p.availableBackends, p.logger))
	}

	for _, replica := range p.replicas(nodesToBeReplaced) {
		steps = append(steps, step.NewStepRemoveNode(replica, p.clusterState, p.allBackends, p.logger))
	}

	if leader := p.leader(nodesToBeReplaced); leader != nil {
		steps = append(steps, step.NewStepRemoveLeader(leader, p.clusterState, p.allBackends, p.logger))
	}

	for i := 0; i < p.clusterShrinkingBy(); i++ {
		steps = append(steps, step.NewStepRemoveRandomNode(p.clusterState, p.allBackends, p.logger))
	}

	return
}

// clusterGrowingBy returns 0 if cluster is staying same # nodes or is reducing in size;
// else returns number of new nodes
func (p plan) clusterGrowingBy() int {
	targetNodeCount := p.newFeatures.NodeCount
	currentNodeCount := p.clusterState.NodeCount()

	if targetNodeCount > currentNodeCount {
		return targetNodeCount - currentNodeCount
	}
	return 0
}

// clusterShrinkingBy returns 0 if cluster is staying same # nodes or is growing in size;
// else returns number of nodes to be removed
func (p plan) clusterShrinkingBy() int {
	targetNodeCount := p.newFeatures.NodeCount
	currentNodeCount := p.clusterState.NodeCount()

	if targetNodeCount < currentNodeCount {
		return currentNodeCount - targetNodeCount
	}
	return 0
}

func (p plan) nodesToBeReplaced() (nodes []*structs.Node) {
	for _, node := range p.clusterState.Nodes {
		validBackend := false
		for _, backend := range p.availableBackends {
			if node.BackendID == backend.ID {
				validBackend = true
				break
			}
		}
		if !validBackend {
			nodes = append(nodes, node)
		}
	}
	return
}

func (p plan) replicas(nodes []*structs.Node) (replicas []*structs.Node) {
	for _, node := range nodes {
		if node.Role != state.LeaderRole {
			replicas = append(replicas, node)
			continue
		}
	}
	return
}

func (p plan) leader(nodes []*structs.Node) *structs.Node {
	for _, node := range nodes {
		if node.Role == state.LeaderRole {
			return node
		}
	}
	return nil
}
