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
