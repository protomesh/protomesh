package config

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/upper-institute/graviflow"
)

type valReader struct {
	val interface{}
}

func NewConfig(val interface{}) graviflow.Config {

	switch typedVal := val.(type) {

	case *time.Time:

		val = typedVal.Format(time.RFC3339)

	case time.Duration:

		val = typedVal.String()

	case int64:

		val = strconv.FormatInt(typedVal, 64)

	case float64:

		val = strconv.FormatFloat(typedVal, 'f', -1, 64)

	case bool:

		if typedVal {
			val = "t"
		} else {
			val = "f"
		}

	case []string:

		raw, _ := json.Marshal(typedVal)

		val = string(raw)

	case string:

		if len(typedVal) == 0 {
			val = nil
		}

	}

	return &valReader{val}
}

func (v *valReader) IsSet() bool {
	return v.val != nil
}

func (v *valReader) StringVal() string {
	return v.val.(string)
}

func (v *valReader) Int64Val() int64 {

	i, _ := strconv.ParseInt(v.StringVal(), 10, 64)

	return i

}

func (v *valReader) Float64Val() float64 {

	i, _ := strconv.ParseFloat(v.StringVal(), 64)

	return i

}

func (v *valReader) StringSliceVal() []string {

	val := []string{}

	json.Unmarshal([]byte(v.StringVal()), &val)

	return val

}

func (v *valReader) BoolVal() bool {

	switch strings.ToLower(v.StringVal()) {
	case "t", "true", "y", "yes":
		return true
	}

	return false

}

func (v *valReader) DurationVal() time.Duration {

	val, _ := time.ParseDuration(v.StringVal())

	return val

}

func (v *valReader) TimeVal() time.Time {

	val, _ := time.Parse(time.RFC3339, v.StringVal())

	return val

}

func (v *valReader) InterfaceVal() interface{} {

	return v.val

}
