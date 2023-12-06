package blog

import (
	"context"
	"github.com/mar-coding/amqp"
	"github.com/mar-coding/personalWebsiteBackend/APIs/proto-gen/components/microservice/v1"
	"github.com/mar-coding/personalWebsiteBackend/configs"
	"github.com/mar-coding/personalWebsiteBackend/pkg/configHandler"
	"github.com/mar-coding/personalWebsiteBackend/pkg/errorHandler"
	"github.com/mar-coding/personalWebsiteBackend/pkg/logger"
	"github.com/mar-coding/personalWebsiteBackend/pkg/mongodb"
	"github.com/mar-coding/personalWebsiteBackend/pkg/serviceInfo"
	"github.com/mar-coding/personalWebsiteBackend/pkg/transport"
)

type AppBootstrapper interface {
	Run(ctx context.Context)
	Migration(ctx context.Context) error
	Shutdown(ctx context.Context) error
	GetServiceInfo() *microservice.ServiceInfo
	GetServiceConfig() *configHandler.BaseConfig[configs.ExtraData]
	GetMongodbConnector() mongodb.Connector
	GetErrorHandler() errorHandler.Handler
}

type Application struct {
	grpcServer transport.GRPCBootstrapper
	httpServer transport.HTTPBootstrapper

	serviceInfo      *microservice.ServiceInfo
	serviceConfig    *configHandler.BaseConfig[configs.ExtraData]
	mongodbConnector mongodb.Connector
	logger           logger.Logger
	error            errorHandler.Handler
	transaction      *mongodb.Transaction

	serviceInfoClient microservice.ServiceInfoServiceClient
	rabbitMq          amqp.Broker
}

func New(
	ctx context.Context,
	grpcServer transport.GRPCBootstrapper,
	httpServer transport.HTTPBootstrapper,
	serviceConfig *configHandler.BaseConfig[configs.ExtraData],
	serviceInfo *serviceInfo.ServiceInfo,
	errorHandler errorHandler.Handler,
	logger logger.Logger,
	rabbitMq amqp.Broker,
) (AppBootstrapper, error) {
	panic("implement me!")
}

func (a *Application) Run(ctx context.Context) {
	panic("implement me!")
}
