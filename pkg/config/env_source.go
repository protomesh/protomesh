package config

import (
	"os"
	"strings"

	"github.com/protomesh/go-app"
)

type envSource struct {
	keyCase app.KeyCase
	configs map[string]app.Config
}

func NewEnvSource(keyCase app.KeyCase) app.ConfigSource {
	return &envSource{
		keyCase: keyCase,
		configs: make(map[string]app.Config),
	}
}

func (e *envSource) Load() error {

	envs := os.Environ()

	for _, env := range envs {

		sep := strings.Index(env, "=")

		key := app.ConvertKeyCase(env[:sep-1], e.keyCase)
		val := env[sep+1:]

		e.configs[key] = NewConfig(val)

	}

	return nil

}

func (e *envSource) Get(k string) app.Config {

	if c, ok := e.configs[k]; ok {
		return c
	}

	return EmptyConfig()

}

func (e *envSource) Has(k string) bool {

	_, ok := e.configs[k]

	return ok

}
