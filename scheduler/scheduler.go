package scheduler

import (
	"github.com/dingotiles/dingo-postgresql-broker/config"
	"github.com/pivotal-golang/lager"
)

type Scheduler struct {
	logger lager.Logger
	config config.Scheduler
}

func NewScheduler(config config.Scheduler, logger lager.Logger) *Scheduler {
	return &Scheduler{
		config: config,
		logger: logger,
	}
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
