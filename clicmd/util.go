package clicmd

import (
	"log"
	"os"

	"github.com/dingotiles/patroni-broker/config"
)

func loadConfig(configPath string) (cfg *config.Config) {
	if os.Getenv("PATRONI_BROKER_CONFIG") != "" {
		configPath = os.Getenv("PATRONI_BROKER_CONFIG")
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	return
}
