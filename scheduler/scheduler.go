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

	s.backends = s.initBackends(config.Backends)
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
	steps := plan.steps()
	clusterModel.NewClusterPlan(len(steps))

	for _, step := range steps {
		err = step.Perform()
		if err != nil {
			return
		}
		clusterModel.PlanStepComplete()
	}
	return
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
	for _, step := range plan.steps() {
		err := step.Perform()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) initBackends(config []*config.Backend) backend.Backends {

	var backends []*backend.Backend
	for _, cfg := range config {
		backends = append(backends, backend.NewBackend(cfg))
	}

	return backends
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
