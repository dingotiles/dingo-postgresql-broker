package scheduler

import (
	"github.com/dingotiles/dingo-postgresql-broker/broker/interfaces"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/cells"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/step"
	"github.com/pivotal-golang/lager"
)

const (
	defaultNodeSize = 20
)

// p.est represents a user-originating p.est to change a service instance (grow, scale, move)
type plan struct {
	clusterModel   interfaces.ClusterModel
	patroni        interfaces.Patroni
	newFeatures    structs.ClusterFeatures
	availableCells cells.Cells
	allCells       cells.Cells
	newNodeSize    int
	logger         lager.Logger
}

// Newp.est creates a p.est to change a service instance
func (s *Scheduler) newPlan(clusterModel interfaces.ClusterModel, features structs.ClusterFeatures) (plan, error) {

	cells, err := s.filterCellsByGUIDs(features.CellGUIDs)
	if err != nil {
		return plan{}, err
	}

	return plan{
		clusterModel:   clusterModel,
		newFeatures:    features,
		newNodeSize:    defaultNodeSize,
		availableCells: cells,
		allCells:       s.cells,
		logger:         s.logger,
		patroni:        s.patroni,
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
	addedNodes := false
	for i := 0; i < p.clusterGrowingBy(); i++ {
		steps = append(steps, step.NewStepAddNode(p.clusterModel, p.patroni, p.availableCells, p.logger))
		addedNodes = true
	}

	nodesToBeReplaced := p.nodesToBeReplaced()
	for _ = range nodesToBeReplaced {
		steps = append(steps, step.NewStepAddNode(p.clusterModel, p.patroni, p.availableCells, p.logger))
		addedNodes = true
	}

	if addedNodes {
		steps = append(steps, step.NewWaitForAllMembers(p.clusterModel, p.patroni, p.logger))
	}

	removedNodes := false
	for _, replica := range p.replicas(nodesToBeReplaced) {
		steps = append(steps, step.NewStepRemoveNode(replica, p.clusterModel, p.allCells, p.logger))
		removedNodes = true
	}

	if leader := p.leader(nodesToBeReplaced); leader != nil {
		steps = append(steps, step.NewStepRemoveLeader(leader, p.clusterModel, p.allCells, p.logger))
		removedNodes = true
	}

	if removedNodes {
		steps = append(steps, step.NewWaitForAllMembers(p.clusterModel, p.patroni, p.logger))
	}

	for i := 0; i < p.clusterShrinkingBy(); i++ {
		steps = append(steps, step.NewStepRemoveRandomNode(p.clusterModel, p.allCells, p.logger))
	}

	if p.newFeatures.NodeCount > 0 {
		steps = append(steps, step.NewWaitForLeader(p.clusterModel, p.patroni, p.logger))
	}

	return
}

// clusterGrowingBy returns 0 if cluster is staying same # nodes or is reducing in size;
// else returns number of new nodes
func (p plan) clusterGrowingBy() int {
	targetNodeCount := p.newFeatures.NodeCount
	currentNodeCount := p.clusterModel.NodeCount()

	if targetNodeCount > currentNodeCount {
		return targetNodeCount - currentNodeCount
	}
	return 0
}

// clusterShrinkingBy returns 0 if cluster is staying same # nodes or is growing in size;
// else returns number of nodes to be removed
func (p plan) clusterShrinkingBy() int {
	targetNodeCount := p.newFeatures.NodeCount
	currentNodeCount := p.clusterModel.NodeCount()

	if targetNodeCount < currentNodeCount {
		return currentNodeCount - targetNodeCount
	}
	return 0
}

func (p plan) nodesToBeReplaced() (nodes []*structs.Node) {
	for _, node := range p.clusterModel.Nodes() {
		validcell := false
		for _, cell := range p.availableCells {
			if node.CellGUID == cell.GUID {
				validcell = true
				break
			}
		}
		if !validcell {
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
