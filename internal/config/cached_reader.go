package config

import (
	"time"

	"github.com/upper-institute/graviflow"
)

type cachedReader struct {
	isSet bool

	stringVal      string
	float64Val     float64
	int64Val       int64
	stringSliceVal []string
	boolVal        bool
	durationVal    time.Duration
	timeVal        time.Time
	interfaceVal   interface{}
}

func CacheConfig(cr graviflow.Config) graviflow.Config {
	return &cachedReader{
		isSet:          cr.IsSet(),
		stringVal:      cr.StringVal(),
		float64Val:     cr.Float64Val(),
		int64Val:       cr.Int64Val(),
		stringSliceVal: cr.StringSliceVal(),
		boolVal:        cr.BoolVal(),
		durationVal:    cr.DurationVal(),
		timeVal:        cr.TimeVal(),
		interfaceVal:   cr.InterfaceVal(),
	}
}

func (c *cachedReader) IsSet() bool {
	return c.isSet
}

func (c *cachedReader) StringVal() string {
	return c.stringVal
}

func (c *cachedReader) Float64Val() float64 {
	return c.float64Val
}

func (c *cachedReader) Int64Val() int64 {
	return c.int64Val
}

func (c *cachedReader) StringSliceVal() []string {
	return c.stringSliceVal
}

func (c *cachedReader) BoolVal() bool {
	return c.boolVal
}

func (c *cachedReader) DurationVal() time.Duration {
	return c.durationVal
}

func (c *cachedReader) TimeVal() time.Time {
	return c.timeVal
}

func (c *cachedReader) InterfaceVal() interface{} {
	return c.interfaceVal
}
