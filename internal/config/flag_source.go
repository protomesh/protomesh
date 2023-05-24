package config

import (
	"flag"

	"dev.azure.com/pomwm/pom-tech/graviflow"
)

type FlagSet interface {
	VisitAll(fn func(*flag.Flag))
}

type flagSource struct {
	keyCase graviflow.KeyCase
	flagSet FlagSet
	configs map[string]graviflow.Config
}

func NewFlagSource(keyCase graviflow.KeyCase, flagSet FlagSet) graviflow.ConfigSource {
	return &flagSource{
		keyCase: keyCase,
		flagSet: flagSet,
		configs: make(map[string]graviflow.Config),
	}
}

func (f *flagSource) Load() error {

	f.flagSet.VisitAll(func(fg *flag.Flag) {

		key := graviflow.ConvertKeyCase(fg.Name, f.keyCase)
		val := fg.Value.String()

		if len(val) == 0 {
			return
		}

		f.configs[key] = NewConfig(val)

	})

	return nil
}

func (f *flagSource) Get(k string) graviflow.Config {

	if c, ok := f.configs[k]; ok {
		return c
	}

	return nil

}
