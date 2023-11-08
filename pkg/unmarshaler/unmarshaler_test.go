package unmarshaler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type MyConfig struct {
	FooBar string `yaml:"foo_bar" json:"foo_bar"`
	Bar    string `yaml:"bar" json:"bar"`
}

func TestYAMLUnmarshaler_Unmarshal(t *testing.T) {
	yamlData := []byte(`
foo_bar: test
bar: 123
`)

	config := &MyConfig{}
	unmarshaler := &yamlUnmarshaler{}

	err := unmarshaler.Unmarshal(yamlData, config)
	if err != nil {
		t.Errorf("Failed to unmarshal YAML: %s", err)
	}

	expectedConfig := &MyConfig{
		FooBar: "test",
		Bar:    "123",
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, config)
}

func TestJSONUnmarshaler_Unmarshal(t *testing.T) {
	jsonData := []byte(`
		{ "foo_bar": "test","bar":"123" }
	`)

	expectedConfig := &MyConfig{
		FooBar: "test",
		Bar:    "123",
	}

	config := &MyConfig{}
	unmarshaler := &jsonUnmarshaler{}

	err := unmarshaler.Unmarshal(jsonData, config)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %s", err)
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, config)
}

func TestCreateUnmarshaler(t *testing.T) {
	t.Run("JSON extension", func(t *testing.T) {
		path := "/path/to/file.json"

		unmarshaler, err := CreateUnmarshaler(path)
		assert.NoError(t, err)
		assert.IsType(t, &jsonUnmarshaler{}, unmarshaler)
	})

	t.Run("YAML extension", func(t *testing.T) {
		path := "/path/to/file.yaml"

		unmarshaler, err := CreateUnmarshaler(path)
		assert.NoError(t, err)
		assert.IsType(t, &yamlUnmarshaler{}, unmarshaler)
	})

	t.Run("Unsupported extension", func(t *testing.T) {
		path := "/path/to/file.txt"

		unmarshaler, err := CreateUnmarshaler(path)
		assert.Error(t, err)
		assert.Nil(t, unmarshaler)
	})
}
