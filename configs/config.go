package configs

import (
	"github.com/mar-coding/personalWebsiteBackend/pkg/config"
	"github.com/mar-coding/personalWebsiteBackend/pkg/errorHandler"
	"github.com/mar-coding/personalWebsiteBackend/pkg/serviceInfo"
	"log"
)

func NewConfig(configPath string) (*config.Config[ExtraData], error) {
	cfg, err := config.New[ExtraData](configPath)
	if err != nil {
		log.Fatal(err)
	}
	return cfg, err
}

func NewServiceInfo() (*serviceInfo.ServiceInfo, error) {
	return serviceInfo.NewFromEmbed(serviceData)
}

func NewError(serviceInfo *serviceInfo.ServiceInfo, serviceConfig *config.Config[ExtraData]) (errorHandler.Handler, error) {
	return errorHandler.NewError(
		uint32(serviceInfo.Code),
		serviceInfo.Name,
		serviceInfo.Version,
		serviceConfig.Domain,
	)
}
