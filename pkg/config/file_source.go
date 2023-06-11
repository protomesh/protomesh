package config

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/protomesh/protomesh"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

type fileSource struct {
	filePath string
	config   gjson.Result
}

func NewFileSource(filePath string) protomesh.ConfigSource {
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

		_, err := toml.Decode(string(raw), &m)
		if err != nil {
			return err
		}

		raw, err = json.Marshal(m)
		if err != nil {
			return err
		}

	default:
		return protomesh.UnkownConfigFormatError

	}

	f.config = gjson.ParseBytes(raw)

	return nil

}

func (f *fileSource) Get(k string) protomesh.Config {

	res := f.config.Get(k)

	if res.Exists() {
		return NewConfig(res.String())
	}

	return EmptyConfig()

}

func (f *fileSource) Has(k string) bool {
	return f.config.Get(k).Exists()
}
