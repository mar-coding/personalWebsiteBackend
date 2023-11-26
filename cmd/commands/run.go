package commands

import (
	"fmt"
	"github.com/mar-coding/personalWebsiteBackend/configs"
	"github.com/mar-coding/personalWebsiteBackend/pkg/logger"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run Personal WebSite",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := configs.NewConfig(configPath)
		if err != nil {
			return err
		}

		info, err := configs.NewServiceInfo()
		if err != nil {
			return err
		}

		errHandler, err := configs.NewError(info, cfg)
		if err != nil {
			return err
		}

		fmt.Println(cfg.ExtraData.Email)
		fmt.Println(cfg.Address)

		return nil
	},
}

func loggerInitiator(cfg *config.BaseConfig) (logger.Logger, error) {
	logging, err := defaultLogging()
	if cfg.logging != nil {
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
		return err
	}
}

func defaultLogging() (logger.Logger, error) {
	return logger.New(logger.CONSOLE_HANDLER, logger.Options{
		Development:  false,
		Debug:        false,
		EnableCaller: true,
		SkipCaller:   3,
	})
}

func permissionOptions(methodFullName string) ([]int32, bool, bool, bool, error) {
	return []int32{}, false, false, false, nil
}
