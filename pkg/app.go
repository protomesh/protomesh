package internal

import (
	"flag"

	"github.com/protomesh/protomesh"
	"github.com/protomesh/protomesh/pkg/config"
	"github.com/protomesh/protomesh/pkg/logging"
)

const (
	configFile_cfg = "config.file"
	logLevel_cfg   = "log.level"
	logDev_cfg     = "log.dev"
	logJson_cfg    = "log.json"
)

func init() {

	flag.String(protomesh.ConvertKeyCase(configFile_cfg, protomesh.KebabCase), "", "[string]\n\tPath to config file (JSON, TOML or YAML)\n")
	flag.String(protomesh.ConvertKeyCase(logLevel_cfg, protomesh.KebabCase), "", "[string]\n\tLog level: debug, info or error\n")
	flag.Bool(protomesh.ConvertKeyCase(logDev_cfg, protomesh.KebabCase), true, "[boolean]\n\tLog enhanced for development environment (no sampling)\n")
	flag.Bool(protomesh.ConvertKeyCase(logJson_cfg, protomesh.KebabCase), false, "[boolean]\n\tLog in json format\n")

}

type app struct {
	cfg protomesh.ConfigSource

	logBuilder *logging.LoggerBuilder
	log        protomesh.Logger
}

func CreateApp[Dependency any](injector protomesh.DependencyInjector[Dependency], configurator *protomesh.Configurator[Dependency]) protomesh.App {

	cfg := config.NewCompositeSource(
		config.NewFlagSource(configurator.KeyCase, configurator.FlagSet),
		config.NewEnvSource(configurator.KeyCase),
	)

	err := cfg.Load()
	if err != nil {
		panic(err)
	}

	configFile := cfg.Get(configFile_cfg)

	if configFile.IsSet() {

		cfg = config.NewCompositeSource(
			config.NewFlagSource(configurator.KeyCase, configurator.FlagSet),
			config.NewEnvSource(configurator.KeyCase),
			config.NewFileSource(configFile.StringVal()),
		)

		err := cfg.Load()
		if err != nil {
			panic(err)
		}

	}

	logBuilder := &logging.LoggerBuilder{
		LogLevel: cfg.Get(logLevel_cfg),
		LogDev:   cfg.Get(logDev_cfg),
		LogJson:  cfg.Get(logJson_cfg),
	}

	log := logBuilder.Build()

	appInstance := &app{
		cfg:        cfg,
		logBuilder: logBuilder,
		log:        log,
	}

	configurator.Source = cfg

	return appInstance

}

func (a *app) Config() protomesh.ConfigSource {
	return a.cfg
}

func (a *app) Log() protomesh.Logger {
	return a.log
}

func (a *app) Close() {
	a.logBuilder.Sync()
}
