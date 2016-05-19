package testutil

import (
	"testing"

	"github.com/pivotal-golang/lager"
)

type logAdapter struct {
	t *testing.T
}

func (l logAdapter) Log(level lager.LogLevel, msg []byte) {
	l.t.Logf("Logged message: %s", string(msg))
}

func NewTestLogger(test string, t *testing.T) lager.Logger {
	logger := lager.NewLogger(test)
	logger.RegisterSink(logAdapter{t: t})
	return logger
}
