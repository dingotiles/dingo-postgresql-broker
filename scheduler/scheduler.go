package scheduler

import (
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

func (s *Scheduler) Execute(req Request) (err error) {
	req.logRequest()
	if len(req.steps()) == 0 {
		req.logger.Info("request.no-steps")
		return
	}
	req.logger.Info("request.perform", lager.Data{"steps-count": len(req.steps())})
	for _, step := range req.steps() {
		err = step.Perform()
		if err != nil {
			return
		}
	}
	return
}

func (s *Scheduler) initBackends(config []*config.Backend) backend.Backends {

	var backends []*backend.Backend
	for _, cfg := range config {
		backends = append(backends, backend.NewBackend(cfg))
	}

	return backends
}
