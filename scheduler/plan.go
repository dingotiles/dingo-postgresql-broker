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
	if p.newFeatures.NodeCount == 0 {
		for i := 0; i < p.clusterModel.NodeCount(); i++ {
			steps = append(steps, step.NewStepRemoveRandomNode(p.clusterModel, "", p.allCells, p.logger))
		}
		return
	}

	// we know the leader must survive because the cluster isn't being deleted
	leaderID, _ := p.patroni.ClusterLeader(p.clusterModel.InstanceID())

	addedNodes := false
	for i := 0; i < p.clusterGrowingBy(); i++ {
		steps = append(steps, step.NewStepAddNode(p.clusterModel, p.patroni, p.availableCells, p.logger))
		addedNodes = true
	}

	replicasToBeReplaced, leaderToBeReplaced := p.nodesToBeReplaced(leaderID)
	for _ = range replicasToBeReplaced {
		steps = append(steps, step.NewStepAddNode(p.clusterModel, p.patroni, p.availableCells, p.logger))
		addedNodes = true
	}
	if leaderToBeReplaced != nil {
		steps = append(steps, step.NewStepAddNode(p.clusterModel, p.patroni, p.availableCells, p.logger))
		addedNodes = true
	}

	if addedNodes {
		steps = append(steps, step.NewWaitForAllMembers(p.clusterModel, p.patroni, p.logger))
	}

	removedNodes := false
	for _, replica := range replicasToBeReplaced {
		steps = append(steps, step.NewStepRemoveNode(replica, p.clusterModel, p.allCells, p.logger))
		removedNodes = true
	}

	for i := 0; i < p.clusterShrinkingBy(); i++ {
		steps = append(steps, step.NewStepRemoveRandomNode(p.clusterModel, leaderID, p.allCells, p.logger))
		removedNodes = true
	}

	if removedNodes {
		steps = append(steps, step.NewWaitForAllMembers(p.clusterModel, p.patroni, p.logger))
	}

	if leaderToBeReplaced != nil {
		steps = append(steps, step.NewStepFailoverFrom(p.clusterModel, leaderID, p.patroni, p.logger))
		steps = append(steps, step.NewStepRemoveNode(leaderToBeReplaced, p.clusterModel, p.allCells, p.logger))
	}

	steps = append(steps, step.NewWaitForLeader(p.clusterModel, p.patroni, p.logger))

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

func (p plan) nodesToBeReplaced(leaderID string) (replicas []*structs.Node, leader *structs.Node) {
	for _, node := range p.clusterModel.Nodes() {
		if !p.availableCells.ContainsCell(node.CellGUID) {
			if node.ID == leaderID {
				leader = node
			} else {
				replicas = append(replicas, node)
			}
		}
	}
	return
}
