package configs

import (
	"github.com/mar-coding/personalWebsiteBackend/pkg/errorHandler"
	"github.com/mar-coding/personalWebsiteBackend/pkg/serviceInfo"
	"log"
)

func NewConfig(configPath string) (*configHandler.BaseConfig[ExtraData], error) {
	cfg, err := configHandler.New[ExtraData](configPath)
	if err != nil {
		log.Fatal(err)
	}
	return cfg, err
}

func NewServiceInfo() (*serviceInfo.ServiceInfo, error) {
	return serviceInfo.NewFromEmbed(serviceData)
}

func NewError(serviceInfo *serviceInfo.ServiceInfo, serviceConfig *configHandler.BaseConfig[ExtraData]) (errorHandler.Handler, error) {
	return errorHandler.NewError(
		uint32(serviceInfo.Code),
		serviceInfo.Name,
		serviceInfo.Version,
		serviceConfig.Domain,
	)
}
