package clicmd

import (
	"log"
	"os"

	"github.com/dingotiles/dingo-postgresql-broker/bkrconfig"
)

func loadConfig(configPath string) (cfg *bkrconfig.Config) {
	if os.Getenv("PATRONI_BROKER_CONFIG") != "" {
		configPath = os.Getenv("PATRONI_BROKER_CONFIG")
	}
	cfg, err := bkrconfig.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	return
}
