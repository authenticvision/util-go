package configutil

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"log/slog"
	"os"
)

// Read reads a configuration from either a YAML file or the environment using cleanenv.
func Read(config interface{}) {
	configPath := flag.String("c", "", "read configuration from YAML file")
	flag.Usage = cleanenv.FUsage(flag.CommandLine.Output(), config, nil, flag.Usage)
	flag.Parse()
	if *configPath != "" {
		if err := cleanenv.ReadConfig(*configPath, config); err != nil {
			slog.Error("error loading configuration file", "error", err)
			os.Exit(1)
		}
	} else if err := cleanenv.ReadEnv(config); err != nil {
		slog.Error("error loading configuration from environment", "error", err)
		os.Exit(1)
	}
}
