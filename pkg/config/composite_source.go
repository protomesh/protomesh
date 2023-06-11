package config

import (
	"github.com/protomesh/protomesh"
)

type compositeSource struct {
	s   []protomesh.ConfigSource
	crs map[string]protomesh.Config
}

func NewCompositeSource(s ...protomesh.ConfigSource) protomesh.ConfigSource {
	return &compositeSource{
		s:   s,
		crs: make(map[string]protomesh.Config),
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

func (c *compositeSource) Get(k string) protomesh.Config {

	if cr, ok := c.crs[k]; ok {
		return cr
	}

	var defVal protomesh.Config

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

	return EmptyConfig()
}

func (c *compositeSource) Has(k string) bool {

	cfg := c.Get(k)

	return cfg != nil && cfg.IsSet()

}
