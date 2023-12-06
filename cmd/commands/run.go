package commands

import (
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/mar-coding/personalWebsiteBackend/APIs"
	"github.com/mar-coding/personalWebsiteBackend/configs"
	"github.com/mar-coding/personalWebsiteBackend/internal/app/blog"
	"github.com/mar-coding/personalWebsiteBackend/pkg/middlewares"
	"github.com/mar-coding/personalWebsiteBackend/pkg/transport"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"strconv"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

const (
	headerSignatureQuery = "X-Signature-Query"
	headerCaptchaQuery   = "X-Captcha-Key"
	headerServiceName    = "X-Service-Name"
)

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

		logging, err := configs.NewLogger(cfg, info)
		if err != nil {
			return err
		}

		grpcAddr := fmt.Sprintf("%s:%s", cfg.Address, strconv.Itoa(cfg.Grpc.Port))
		grpcServer, err := transport.NewGRPCServer(
			grpcAddr,
			cfg.Development,
			middlewares.GrpcRecovery(logging),
			middlewares.GrpcSentryPerformance(logging.GetSentryClient()),
			middlewares.GrpcValidator(errHandler),
			middlewares.GrpcJwtMiddleware(permissionOptions, info, errHandler, os.Getenv("PUBLIC_SECRET"),
				os.Getenv("PRIVATE_SECRET"), ""),
			middlewares.GRPCLogging(logging),
		)

		httpAddr := fmt.Sprintf("%s:%s", cfg.Address, strconv.Itoa(cfg.Rest.Port))
		httpServer := transport.NewHTTPServer(
			cmd.Context(),
			httpAddr,
			grpcAddr,
			cfg.Development,
			APIs.Swagger,
			[]string{headerSignatureQuery, headerCaptchaQuery, headerServiceName, "Content-Type", "Content-Disposition"},
			cfg.Origins,
			defaultHttpMiddlewareFunc,
			runtime.WithIncomingHeaderMatcher(headers),
			runtime.WithErrorHandler(middlewares.ErrorHandler),
		)

		application, err := blog.New(cmd.Context(),
			grpcServer, httpServer,
			cfg,
			info,
			errHandler,
			logging,
			nil,
		)
		if err != nil {
			return err
		}

		application.Run(cmd.Context())
		return nil
	},
}

func headers(key string) (string, bool) {
	switch key {
	case headerCaptchaQuery:
		return key, true
	case headerServiceName:
		return key, true
	case headerSignatureQuery:
		return key, true
	default:
		return key, false
	}
}

func permissionOptions(methodFullName string) ([]int32, bool, bool, bool, error) {
	return []int32{}, false, false, false, nil
}

func defaultHttpMiddlewareFunc(handler http.Handler) http.Handler {
	return handler
}
