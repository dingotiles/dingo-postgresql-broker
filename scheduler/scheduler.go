package scheduler

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/interfaces"
	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/patroni"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/cells"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

type Scheduler struct {
	logger  lager.Logger
	config  config.Scheduler
	cells   cells.Cells
	patroni *patroni.Patroni
}

func NewScheduler(config config.Scheduler, logger lager.Logger) (*Scheduler, error) {
	s := &Scheduler{
		config: config,
		logger: logger,
	}

	patroni, err := patroni.NewPatroni(config.Etcd, s.logger)
	if err != nil {
		return nil, err
	}
	s.patroni = patroni

	clusterLoader, err := state.NewStateEtcd(config.Etcd, s.logger)
	if err != nil {
		return nil, err
	}
	s.cells = cells.NewCells(config.Cells, clusterLoader)

	return s, nil
}

func (s *Scheduler) RunCluster(clusterModel interfaces.ClusterModel, features structs.ClusterFeatures) (err error) {
	err = s.VerifyClusterFeatures(features)
	if err != nil {
		return
	}

	plan, err := s.newPlan(clusterModel, features)
	if err != nil {
		return
	}

	s.logger.Info("scheduler.run-cluster", lager.Data{
		"instance-id": clusterModel.InstanceID(),
		"steps-count": len(plan.steps()),
		"steps":       plan.stepTypes(),
		"features":    features,
	})

	return s.executePlan(clusterModel, plan)
}

func (s *Scheduler) StopCluster(clusterModel interfaces.ClusterModel) error {
	plan, err := s.newPlan(clusterModel, structs.ClusterFeatures{NodeCount: 0})
	if err != nil {
		return err
	}

	s.logger.Info("scheduler.stop-cluster", lager.Data{
		"instance-id": clusterModel.InstanceID(),
		"plan":        plan,
		"steps-count": len(plan.steps()),
		"steps":       plan.stepTypes(),
	})

	return s.executePlan(clusterModel, plan)
}

func (s *Scheduler) executePlan(clusterModel interfaces.ClusterModel, plan plan) error {
	steps := plan.steps()
	clusterModel.BeginScheduling(len(steps))

	for _, step := range steps {
		clusterModel.SchedulingStepStarted(step.StepType())
		err := step.Perform()
		if err != nil {
			clusterModel.SchedulingError(err)
			return err
		}
		clusterModel.SchedulingStepCompleted()
	}
	return nil
}

func (s *Scheduler) VerifyClusterFeatures(features structs.ClusterFeatures) (err error) {
	availableCells, err := s.filterCellsByGUIDs(features.CellGUIDs)
	if err != nil {
		return
	}
	if features.NodeCount > len(availableCells) {
		availableCellGUIDs := make([]string, len(availableCells))
		for i, cell := range availableCells {
			availableCellGUIDs[i] = cell.GUID
		}
		err = fmt.Errorf("Scheduler: Not enough Cell GUIDs (%v) for cluster of %d nodes", availableCellGUIDs, features.NodeCount)
	}
	return
}

// filterCellsByGUIDs returns all cell cells; or the subset filtered by cellGUIDS; or an error
func (s *Scheduler) filterCellsByGUIDs(cellGUIDs []string) (cells.Cells, error) {
	if len(cellGUIDs) > 0 {
		var filteredCells []*cells.Cell
		for _, cellGUID := range cellGUIDs {
			foundCellGUID := false
			for _, cell := range s.cells {
				if cellGUID == cell.GUID {
					filteredCells = append(filteredCells, cell)
					foundCellGUID = true
					continue
				}
			}
			if !foundCellGUID {
				s.logger.Info("scheduler.filter-cells.unknown-cell-guid", lager.Data{"cell-guid": cellGUID})
			}
		}
		if len(filteredCells) == 0 {
			return filteredCells, fmt.Errorf("Scheduler: Cell GUIDs do not match available cells")
		}
		return filteredCells, nil
	} else {
		return s.cells, nil
	}
}
