package configs

import (
	"fmt"
	"github.com/mar-coding/personalWebsiteBackend/pkg/unmarshaller"
	"os"
)

func NewConfig(configPath string) (*Config, error) {
	b, err := os.ReadFile(configPath)
	if err != nil {
		msg := fmt.Errorf("failed to read config file in %s, got error %v", configPath, err)
		return nil, msg
	}

	unMarshaler, err := unmarshaller.CreateUnmarshaller(configPath)
	config := &Config{}
	err = unMarshaler.Unmarshal(b, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
