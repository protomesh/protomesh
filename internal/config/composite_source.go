package config

import (
	"github.com/upper-institute/graviflow"
)

type compositeSource struct {
	s   []graviflow.ConfigSource
	crs map[string]graviflow.Config
}

func NewCompositeSource(s ...graviflow.ConfigSource) graviflow.ConfigSource {
	return &compositeSource{
		s:   s,
		crs: make(map[string]graviflow.Config),
	}
}

func (c *compositeSource) Load() error {

	for _, cs := range c.s {

		err := cs.Load()
		if err != nil {
			return err
		}

	}

	return nil

}

func (c *compositeSource) Get(k string) graviflow.Config {

	if cr, ok := c.crs[k]; ok {
		return cr
	}

	var defVal graviflow.Config

	for _, s := range c.s {

		cr := s.Get(k)

		if cr != nil && cr.IsSet() {

			if s.Has(k) {
				c.crs[k] = cr
				return cr
			}

			defVal = cr

		}

	}

	if defVal != nil {
		return defVal
	}

	return emptyConfig()
}

func (c *compositeSource) Has(k string) bool {

	cfg := c.Get(k)

	return cfg != nil && cfg.IsSet()

}
