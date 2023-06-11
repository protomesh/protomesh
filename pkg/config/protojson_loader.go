package config

import (
	"encoding/json"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/protomesh/protomesh"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

type ProtoJson_SourceFormat string

const (
	ProtoJson_FromJson ProtoJson_SourceFormat = "json"
	ProtoJson_FromToml ProtoJson_SourceFormat = "toml"
	ProtoJson_FromYaml ProtoJson_SourceFormat = "yaml"
)

func ProtoJsonFileExtensionToFormat(filePath string) (ProtoJson_SourceFormat, error) {

	extSep := strings.LastIndex(filePath, ".")

	ext := strings.ToLower(filePath[extSep+1:])

	switch ext {

	case "json":
		return ProtoJson_FromJson, nil

	case "yml", "yaml":
		return ProtoJson_FromYaml, nil

	case "toml":
		return ProtoJson_FromToml, nil

	}

	return "", protomesh.UnkownConfigFormatError

}

func ProtoJsonUnmarshal[M proto.Message](buf []byte, enc ProtoJson_SourceFormat, m M) error {

	switch enc {

	case ProtoJson_FromJson:
		break

	case ProtoJson_FromYaml:

		yamlMsg := make(map[string]interface{})

		err := yaml.Unmarshal([]byte(buf), &yamlMsg)
		if err != nil {
			return err
		}

		buf, err = json.Marshal(yamlMsg)
		if err != nil {
			return err
		}

	case ProtoJson_FromToml:

		tomlMsg := make(map[string]interface{})

		_, err := toml.Decode(string(buf), &tomlMsg)
		if err != nil {
			return err
		}

		buf, err = json.Marshal(tomlMsg)
		if err != nil {
			return err
		}

	default:
		return protomesh.UnkownConfigFormatError

	}

	return protojson.Unmarshal(buf, m)
}
