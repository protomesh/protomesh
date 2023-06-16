package config

import "github.com/protomesh/go-app"

type compositeSource struct {
	s   []app.ConfigSource
	crs map[string]app.Config
}

func NewCompositeSource(s ...app.ConfigSource) app.ConfigSource {
	return &compositeSource{
		s:   s,
		crs: make(map[string]app.Config),
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

func (c *compositeSource) Get(k string) app.Config {

	if cr, ok := c.crs[k]; ok {
		return cr
	}

	var defVal app.Config

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
