package scheduler

import (
	"fmt"

	"github.com/dingotiles/dingo-postgresql-broker/broker/structs"
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/dingotiles/dingo-postgresql-broker/scheduler/backend"
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

func (s *Scheduler) RunCluster(cluster structs.ClusterState, features structs.ClusterFeatures) (structs.ClusterState, error) {
	plan, err := s.newPlan(&cluster, features)
	if err != nil {
		return cluster, err
	}

	s.logger.Info("scheduler.run-cluster", lager.Data{
		"steps-count": len(plan.steps()),
		"features":    features,
	})
	for _, step := range plan.steps() {
		err := step.Perform()
		if err != nil {
			return cluster, err
		}
	}
	return cluster, nil
}

func (s *Scheduler) StopCluster(cluster structs.ClusterState) (structs.ClusterState, error) {
	plan, err := s.newPlan(&cluster, structs.ClusterFeatures{NodeCount: 0})
	if err != nil {
		return cluster, err
	}

	s.logger.Info("scheduler.stop-cluster", lager.Data{
		"plan":        plan,
		"steps-count": len(plan.steps()),
	})
	for _, step := range plan.steps() {
		err := step.Perform()
		if err != nil {
			return cluster, err
		}
	}
	return cluster, nil
}

func (s *Scheduler) initBackends(config []*config.Backend) backend.Backends {

	var backends []*backend.Backend
	for _, cfg := range config {
		backends = append(backends, backend.NewBackend(cfg))
	}

	return backends
}

func (s *Scheduler) filterBackendsByCellGUIDs(cellGUIDs []string) (backend.Backends, error) {
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
			return filteredBackends, fmt.Errorf("Cell GUIDs do not match available cells")
		}
		return filteredBackends, nil
	} else {
		return s.backends, nil
	}
}
