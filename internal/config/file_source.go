package config

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"dev.azure.com/pomwm/pom-tech/graviflow"
	"github.com/BurntSushi/toml"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

type fileSource struct {
	filePath string
	config   gjson.Result
}

func NewFileSource(filePath string) graviflow.ConfigSource {
	return &fileSource{
		filePath: filePath,
	}
}

func (f *fileSource) Load() error {

	raw, err := ioutil.ReadFile(f.filePath)
	if err != nil {
		return err
	}

	extSep := strings.LastIndex(f.filePath, ".")

	ext := strings.ToLower(f.filePath[extSep+1:])

	switch ext {

	case "json":
		break

	case "yml", "yaml":

		m := make(map[string]interface{})

		err := yaml.Unmarshal([]byte(raw), &m)
		if err != nil {
			return err
		}

		raw, err = json.Marshal(m)
		if err != nil {
			return err
		}

	case "toml":

		m := make(map[string]interface{})

		_, err := toml.Decode(string(raw), m)
		if err != nil {
			return err
		}

		raw, err = json.Marshal(m)
		if err != nil {
			return err
		}

	default:
		return graviflow.UnkownConfigFormatError

	}

	f.config = gjson.ParseBytes(raw)

	return nil

}

func (f *fileSource) Get(k string) graviflow.Config {

	res := f.config.Get(k)

	if res.Exists() {
		return NewConfig(res.String())
	}

	return emptyConfig()

}
