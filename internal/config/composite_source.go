package config

import "dev.azure.com/pomwm/pom-tech/graviflow"

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

	for _, s := range c.s {

		cr := s.Get(k)

		if cr.IsSet() {
			c.crs[k] = cr
			return cr
		}

	}

	return emptyConfig()
}
