package unmarshaler

import (
	"encoding/json"
	"errors"
	"gopkg.in/yaml.v3"
	"path/filepath"
)

type Extension string

const (
	JSON Extension = ".json"
	YAML Extension = ".yaml"
	YML  Extension = ".yml"
)

type Unmarshaler interface {
	Unmarshal(payload []byte, config interface{}) error
}

type jsonUnmarshaler struct{}

type yamlUnmarshaler struct{}

func (u yamlUnmarshaler) Unmarshal(payload []byte, config interface{}) error {
	if err := yaml.Unmarshal(payload, config); err != nil {
		return err
	}
	return nil
}

func (u jsonUnmarshaler) Unmarshal(payload []byte, config interface{}) error {
	if err := json.Unmarshal(payload, config); err != nil {
		return err
	}
	return nil
}

// CreateUnmarshaler FactoryPattern function to create the appropriate Unmarshaler based on the file extension
func CreateUnmarshaler(path string) (Unmarshaler, error) {
	ext := filepath.Ext(path)
	switch Extension(ext) {
	case JSON:
		return &jsonUnmarshaler{}, nil
	case YAML, YML:
		return &yamlUnmarshaler{}, nil
	default:
		return nil, errors.New("unsupported file extension")
	}
}
