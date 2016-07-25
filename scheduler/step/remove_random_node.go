package step

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/dingotiles/dingo-postgresql-broker/broker/interfaces"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/cells"
	"github.com/pivotal-golang/lager"
)

// RemoveRandomNode instructs cluster to delete a node, starting with replicas
type RemoveRandomNode struct {
	clusterModel interfaces.ClusterModel
	leaderID     string
	cells        cells.Cells
	logger       lager.Logger
}

// NewStepRemoveRandomNode creates a StepRemoveRandomNode command
func NewStepRemoveRandomNode(clusterModel interfaces.ClusterModel, leaderID string, cells cells.Cells, logger lager.Logger) Step {
	return RemoveRandomNode{clusterModel: clusterModel, leaderID: leaderID, cells: cells, logger: logger}
}

// StepType prints the type of step
func (step RemoveRandomNode) StepType() string {
	return "RemoveRandomNode"
}

// Perform runs the Step action to modify the Cluster
func (step RemoveRandomNode) Perform() (err error) {
	logger := step.logger

	// 1. Get list of replicas and pick a random one
	nodes := step.clusterModel.Nodes()
	nodeToRemove := randomNode(nodes, step.leaderID)

	cell := step.cells.Get(nodeToRemove.CellGUID)
	if cell == nil {
		err = fmt.Errorf("Internal error: node assigned to a cell that no longer exists (%s)", nodeToRemove.CellGUID)
		logger.Error("remove-random-node.perform", err)
		return
	}

	logger.Info("remove-random-node.perform", lager.Data{
		"instance-id": step.clusterModel.InstanceID(),
		"node-uuid":   nodeToRemove.ID,
		"cell":        cell.GUID,
	})

	err = cell.DeprovisionNode(nodeToRemove, logger)
	if err != nil {
		return nil
	}

	err = step.clusterModel.RemoveNode(nodeToRemove)
	if err != nil {
		logger.Error("remove-random-node.nodes-delete", err)
	}
	return
}

// currently random any node, doesn't have to be a replica
func randomNode(nodes []*structs.Node, leaderID string) *structs.Node {
	n := rand.Intn(len(nodes))
	if nodes[n].ID == leaderID {
		n = int(math.Mod(float64(n+1), float64(len(nodes))))
	}
	return nodes[n]
}
