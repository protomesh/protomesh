package config

import (
	"time"

	"github.com/protomesh/go-app"
)

type emptyReader bool

func EmptyConfig() app.Config {
	return emptyReader(false)
}

func (emptyReader) IsSet() bool {
	return false
}

func (emptyReader) StringVal() string {
	return ""
}

func (emptyReader) Int64Val() int64 {
	return 0
}

func (emptyReader) Float64Val() float64 {
	return 0
}

func (emptyReader) StringSliceVal() []string {
	return []string{}
}

func (emptyReader) BoolVal() bool {
	return false
}

func (emptyReader) DurationVal() time.Duration {
	return 0
}

func (emptyReader) TimeVal() time.Time {
	return time.Time{}
}

func (emptyReader) InterfaceVal() interface{} {
	return nil
}
