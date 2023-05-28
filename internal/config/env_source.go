package config

import (
	"os"
	"strings"

	"github.com/upper-institute/graviflow"
)

type envSource struct {
	keyCase graviflow.KeyCase
	configs map[string]graviflow.Config
}

func NewEnvSource(keyCase graviflow.KeyCase) graviflow.ConfigSource {
	return &envSource{
		keyCase: keyCase,
		configs: make(map[string]graviflow.Config),
	}
}

func (e *envSource) Load() error {

	envs := os.Environ()

	for _, env := range envs {

		sep := strings.Index(env, "=")

		key := graviflow.ConvertKeyCase(env[:sep-1], e.keyCase)
		val := env[sep+1:]

		e.configs[key] = NewConfig(val)

	}

	return nil

}

func (e *envSource) Get(k string) graviflow.Config {

	if c, ok := e.configs[k]; ok {
		return c
	}

	return emptyConfig()

}

func (e *envSource) Has(k string) bool {

	_, ok := e.configs[k]

	return ok

}
