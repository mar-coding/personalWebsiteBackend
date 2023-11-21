package unmarshaller

import (
	"encoding/json"
	"errors"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
	"path/filepath"
)

type Extension string

const (
	JSON Extension = ".json"
	YAML Extension = ".yaml"
	YML  Extension = ".yml"
	TOML Extension = ".toml"
)

type Unmarshaller interface {
	Unmarshal(payload []byte, config interface{}) error
}

type jsonUnmarshaller struct{}

type yamlUnmarshaller struct{}

type tomlUnmarshaller struct{}

func (u yamlUnmarshaller) Unmarshal(payload []byte, config interface{}) error {
	if err := yaml.Unmarshal(payload, config); err != nil {
		return err
	}
	return nil
}

func (u jsonUnmarshaller) Unmarshal(payload []byte, config interface{}) error {
	if err := json.Unmarshal(payload, config); err != nil {
		return err
	}
	return nil
}

func (u tomlUnmarshaller) Unmarshal(payload []byte, config interface{}) error {
	if err := toml.Unmarshal(payload, config); err != nil {
		return err
	}
	return nil
}

// CreateUnmarshaller FactoryPattern function to create the appropriate Unmarshaller based on the file extension
func CreateUnmarshaller(path string) (Unmarshaller, error) {
	ext := filepath.Ext(path)
	switch Extension(ext) {
	case JSON:
		return &jsonUnmarshaller{}, nil
	case YAML, YML:
		return &yamlUnmarshaller{}, nil
	case TOML:
		return &tomlUnmarshaller{}, nil
	default:
		return nil, errors.New("unsupported file extension")
	}
}
