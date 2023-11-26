package configs

import (
	_ "embed"
)

//go:embed service_info.yml
var serviceData []byte

type ExtraData struct {
	Email   string `yaml:"email" json:"email"`
	Counter int    `yaml:"counter" json:"counter"`
}
