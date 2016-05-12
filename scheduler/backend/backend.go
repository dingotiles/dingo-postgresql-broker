package backend

import "github.com/dingotiles/dingo-postgresql-broker/config"

type Backend struct {
	Config *config.Backend
}

type Backends []*Backend

func NewBackend(config *config.Backend) *Backend {
	return &Backend{
		Config: config,
	}
}
