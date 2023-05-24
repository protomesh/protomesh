package internal

import (
	"flag"

	"dev.azure.com/pomwm/pom-tech/graviflow"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal/config"
	"dev.azure.com/pomwm/pom-tech/graviflow/internal/logging"
)

const (
	configFile_cfg = "config.file"
	logLevel_cfg   = "log.level"
	logDev_cfg     = "log.dev"
	logJson_cfg    = "log.json"
)

func init() {

	flag.String(graviflow.ConvertKeyCase(configFile_cfg, graviflow.KebabCase), "", "[string]\n\tPath to config file (JSON, TOML or YAML)\n")
	flag.String(graviflow.ConvertKeyCase(logLevel_cfg, graviflow.KebabCase), "", "[string]\n\tLog level: debug, info or error\n")
	flag.Bool(graviflow.ConvertKeyCase(logDev_cfg, graviflow.KebabCase), true, "[boolean]\n\tLog enhanced for development environment (no sampling)\n")
	flag.Bool(graviflow.ConvertKeyCase(logJson_cfg, graviflow.KebabCase), true, "[boolean]\n\tLog in json format\n")

}

type app[Dependency any] struct {
	dep Dependency
	cfg graviflow.ConfigSource

	logBuilder *logging.LoggerBuilder
	log        graviflow.Logger
}

func CreateApp[Dependency any](injector graviflow.DependencyInjector[Dependency], configurator *graviflow.Configurator[Dependency]) graviflow.App[Dependency] {

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

	appInstance := &app[Dependency]{
		dep:        injector.Dependency(),
		cfg:        cfg,
		logBuilder: logBuilder,
		log:        log,
	}

	configurator.Source = cfg

	return appInstance

}

func (a *app[Dependency]) Config() graviflow.ConfigSource {
	return a.cfg
}

func (a *app[Dependency]) Log() graviflow.Logger {
	return a.log
}

func (a *app[Dependency]) Dependency() Dependency {
	return a.dep
}

func (a *app[Dependency]) Close() {
	a.logBuilder.Sync()
}
