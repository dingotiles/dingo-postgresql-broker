package scheduler

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
	"github.com/dingotiles/dingo-postgresql-broker/state"
	"github.com/pivotal-golang/lager"
)

type Scheduler struct {
	logger   lager.Logger
	config   config.Scheduler
	backends backend.Backends
}

func NewScheduler(config config.Scheduler, logger lager.Logger) *Scheduler {
	s := &Scheduler{
		config: config,
		logger: logger,
	}

	s.backends = backend.NewBackends(config.Backends)
	return s
}

func (s *Scheduler) RunCluster(clusterModel *state.ClusterModel, features structs.ClusterFeatures) (err error) {
	err = s.VerifyClusterFeatures(features)
	if err != nil {
		return
	}

	plan, err := s.newPlan(clusterModel, s.config.Etcd, features)
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

func (s *Scheduler) StopCluster(clusterModel *state.ClusterModel) error {
	plan, err := s.newPlan(clusterModel, s.config.Etcd, structs.ClusterFeatures{NodeCount: 0})
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

func (s *Scheduler) executePlan(clusterModel *state.ClusterModel, plan plan) error {
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
			availableCellGUIDs[i] = cell.ID
		}
		err = fmt.Errorf("Scheduler: Not enough Cell GUIDs (%v) for cluster of %d nodes", availableCellGUIDs, features.NodeCount)
	}
	return
}

// filterCellsByGUIDs returns all backend cells; or the subset filtered by cellGUIDS; or an error
func (s *Scheduler) filterCellsByGUIDs(cellGUIDs []string) (backend.Backends, error) {
	if len(cellGUIDs) > 0 {
		var filteredBackends []*backend.Backend
		for _, cellGUID := range cellGUIDs {
			foundCellGUID := false
			for _, backend := range s.backends {
				if cellGUID == backend.ID {
					filteredBackends = append(filteredBackends, backend)
					foundCellGUID = true
					continue
				}
			}
			if !foundCellGUID {
				s.logger.Info("scheduler.filter-backends.unknown-cell-guid", lager.Data{"cell-guid": cellGUID})
			}
		}
		if len(filteredBackends) == 0 {
			return filteredBackends, fmt.Errorf("Scheduler: Cell GUIDs do not match available cells")
		}
		return filteredBackends, nil
	} else {
		return s.backends, nil
	}
}
