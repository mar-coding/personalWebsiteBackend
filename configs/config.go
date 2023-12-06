package configs

import (
	"github.com/mar-coding/personalWebsiteBackend/pkg/configHandler"
	"github.com/mar-coding/personalWebsiteBackend/pkg/errorHandler"
	"github.com/mar-coding/personalWebsiteBackend/pkg/logger"
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

func NewLogger(cfg *configHandler.BaseConfig[ExtraData], info *serviceInfo.ServiceInfo) (logger.Logger, error) {
	logging, err := defaultLogging()
	if cfg.Logging != nil {
		logOpt := logger.Options{
			Development:  cfg.Development,
			Debug:        cfg.Logging.Debug,
			EnableCaller: cfg.Logging.EnableCaller,
			SkipCaller:   3,
		}

		if len(cfg.Logging.SentryDSN) != 0 {
			logOpt.Sentry = &logger.SentryConfig{
				DSN:              cfg.Logging.SentryDSN,
				AttachStacktrace: true,
				ServerName:       info.Name,
				Environment:      logger.DEVELOPMENT,
				EnableTracing:    true,
				Debug:            true,
				TracesSampleRate: 1.0,
			}
			if !cfg.Development {
				logOpt.Sentry.Environment = logger.PRODUCTION
			}
		}

		logging, err = logger.New(logger.HandleType(cfg.Logging.Handler), logOpt)
	}
	if err != nil {
		return nil, err
	}
	return logging, nil
}

func defaultLogging() (logger.Logger, error) {
	return logger.New(logger.CONSOLE_HANDLER, logger.Options{
		Development:  false,
		Debug:        false,
		EnableCaller: true,
		SkipCaller:   3,
	})
}
