package config

import (
	"os"
	"strings"

	"github.com/protomesh/protomesh"
)

type envSource struct {
	keyCase protomesh.KeyCase
	configs map[string]protomesh.Config
}

func NewEnvSource(keyCase protomesh.KeyCase) protomesh.ConfigSource {
	return &envSource{
		keyCase: keyCase,
		configs: make(map[string]protomesh.Config),
	}
}

func (e *envSource) Load() error {

	envs := os.Environ()

	for _, env := range envs {

		sep := strings.Index(env, "=")

		key := protomesh.ConvertKeyCase(env[:sep-1], e.keyCase)
		val := env[sep+1:]

		e.configs[key] = NewConfig(val)

	}

	return nil

}

func (e *envSource) Get(k string) protomesh.Config {

	if c, ok := e.configs[k]; ok {
		return c
	}

	return EmptyConfig()

}

func (e *envSource) Has(k string) bool {

	_, ok := e.configs[k]

	return ok

}
