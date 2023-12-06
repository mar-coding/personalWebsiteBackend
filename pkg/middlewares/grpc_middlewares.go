package middlewares

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/getsentry/sentry-go"
	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcTags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/mar-coding/personalWebsiteBackend/pkg/acl"
	"github.com/mar-coding/personalWebsiteBackend/pkg/encryption"
	"github.com/mar-coding/personalWebsiteBackend/pkg/errorHandler"
	"github.com/mar-coding/personalWebsiteBackend/pkg/helper/languageLocalize"
	"github.com/mar-coding/personalWebsiteBackend/pkg/jwt"
	"github.com/mar-coding/personalWebsiteBackend/pkg/logger"
	"github.com/mar-coding/personalWebsiteBackend/pkg/serviceInfo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log/slog"
	"strings"
	"time"
)

const (
	SESSION_COOKIE_KEY = "session_token"
)

// Deprecated: PermissionDescriptor get permissions and credentialType from descriptor, default validate is true in implementation
type PermissionDescriptor func(ctx context.Context) ([]int32, bool, bool, bool)

// PermissionFunc get permissions and credentialType from descriptor, default validate is true in implementation
type PermissionFunc func(methodFullName string) ([]int32, bool, bool, bool, error)

type validator interface {
	ValidateAll() error
}

type validatorLegacy interface {
	Validate() error
}

// New create chained middleware for grpc
func New(middlewares ...grpc.UnaryServerInterceptor) grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(
		middlewares...,
	)
}

func GRPCLogging(logger logger.Logger) grpc.UnaryServerInterceptor {
	logFunc := logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		logger.Log(ctx, false, slog.Level(lvl), msg, fields...)
	})

	opts := []logging.Option{
		logging.WithLevels(logging.DefaultServerCodeToLevel),
		logging.WithLogOnEvents(logging.FinishCall, logging.StartCall),
	}

	return logging.UnaryServerInterceptor(logFunc, opts...)
}

// GrpcRecovery recovery panics
func GrpcRecovery(logger logger.Logger) grpc.UnaryServerInterceptor {
	rec := func(p interface{}) (err error) {
		err = status.Errorf(codes.Unknown, "%v", p)
		logger.Error(true, "recovery: panic triggered", "error", err)
		return
	}
	opts := []grpcRecovery.Option{
		grpcRecovery.WithRecoveryHandler(rec),
	}
	return grpcRecovery.UnaryServerInterceptor(opts...)
}

// GrpcValidator validate your message fields, for user validator please check https://github.com/envoyproxy/protoc-gen-validate
func GrpcValidator(errHandler errorHandler.Handler) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		switch in := req.(type) {
		case validator:
			if err = in.ValidateAll(); err != nil {
				return nil, errHandler.New(codes.InvalidArgument, nil, err.Error())
			}
		case validatorLegacy:
			if err = in.Validate(); err != nil {
				return nil, errHandler.New(codes.InvalidArgument, nil, err.Error())
			}
		}
		return handler(ctx, req)
	}
}

// GrpcJwtMiddleware middleware for check jwt token
func GrpcJwtMiddleware(permissionDescriptor PermissionFunc, serviceInfo *serviceInfo.ServiceInfo, errHandler errorHandler.Handler, jwtPublicSecret, jwtPrivateSecret, captchaSecretKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		permissions, optional, validate, captcha, err := permissionDescriptor(strings.Replace(info.FullMethod[1:], "/", ".", -1))
		if err != nil {
			return nil, err
		}

		if captcha && len(captchaSecretKey) != 0 {
			if err := googleCaptchaValidation(ctx, captchaSecretKey, errHandler); err != nil {
				return nil, err
			}
		}

		if isPermissionZero(permissions) && !optional {
			return handler(ctx, req)
		}

		if optional {
			_, err = acl.GetBearerTokenFromGrpcContext(ctx)
			if err == nil {
				aclController, err := notOptionalAclContext(ctx, serviceInfo, errHandler, jwtPublicSecret, jwtPrivateSecret, validate, optional, permissions)
				if err != nil {
					return nil, err
				}
				return handler(aclController.SetAclToContext(ctx), req)
			}

			return handler(ctx, req)
		}

		aclController, err := notOptionalAclContext(ctx, serviceInfo, errHandler, jwtPublicSecret, jwtPrivateSecret, validate, optional, permissions)
		if err != nil {
			return nil, err
		}
		return handler(aclController.SetAclToContext(ctx), req)
	}
}

// GrpcSessionMiddleware middleware for check session expire time
func GrpcSessionMiddleware(permissionDescriptor PermissionFunc, errHandler errorHandler.Handler, sessionPrivateKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		permissions, _, _, _, err := permissionDescriptor(strings.Replace(info.FullMethod[1:], "/", ".", -1))
		if err != nil {
			return nil, err
		}

		if isPermissionZero(permissions) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, err
		}

		sessionToken := md.Get(SESSION_COOKIE_KEY)
		if len(sessionToken) == 0 {
			return nil, errHandler.New(codes.Unauthenticated, nil, "session_token cookie not found in request")
		}

		sessionByte, err := base64.StdEncoding.DecodeString(sessionToken[0])
		if err != nil {
			return nil, errHandler.New(codes.Unauthenticated, nil, "session_token is invalid")
		}

		aesEncryption, err := encryption.NewAES[int64](sessionPrivateKey)
		expTimeInt, err := aesEncryption.Decrypt(sessionByte)
		if err != nil {
			return nil, errHandler.New(codes.Unauthenticated, nil, "session_token is invalid")
		}

		if time.UnixMicro(expTimeInt).Before(time.Now()) {
			return nil, errHandler.New(codes.Unauthenticated, nil, "token has been expired")
		}

		return handler(ctx, req)
	}
}

func notOptionalAclContext(ctx context.Context, serviceInfo *serviceInfo.ServiceInfo, errHandler errorHandler.Handler, jwtPublicSecret, jwtPrivateSecret string, validate, optional bool, permissions []int32) (*acl.JwtController, error) {
	jwtToken, err := acl.GetBearerTokenFromGrpcContext(ctx)
	if err != nil {
		if errors.Is(err, acl.ErrNotFoundJwtTokenInHeader) {
			return nil, errHandler.New(codes.Unauthenticated, nil, "jwt token not found in header")
		}

		if errors.Is(err, acl.ErrMissingBearerPrefixInToken) {
			return nil, errHandler.New(codes.Unauthenticated, nil, "missing prefix Bearer in your token")
		}
	}

	aclController, err := acl.NewWithJwt(serviceInfo.Code, jwtToken,
		acl.WithValidateACL(validate),
		acl.WithPublicSecretKey(jwtPublicSecret),
		acl.WithPrivateSecretKey(jwtPrivateSecret),
	)

	if err != nil {
		if errors.Is(err, acl.ErrSecretKeyIsEmpty) {
			return nil, errHandler.New(codes.Internal, nil, "public secret key is empty")
		}
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errHandler.New(codes.Unauthenticated, nil, "jwt token expired")
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, errHandler.New(codes.Unauthenticated, nil, "jwt token signature is invalid")
		}
		return nil, errHandler.New(codes.Unauthenticated, nil, "jwt token is invalid")
	}

	if !optional {
		if !aclController.HasAnyPermissionsAccess(permissions...) {
			return nil, errHandler.New(codes.PermissionDenied, nil, "you don't have permission to access method")
		}
	}

	return aclController, nil
}

// GrpcSentryPerformance track request performance in sentry performance
func GrpcSentryPerformance(client *sentry.Client, opts ...Option) grpc.UnaryServerInterceptor {
	o := newConfig(opts)
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		hub := sentry.NewHub(client, sentry.NewScope())
		ctx = sentry.SetHubOnContext(ctx, hub)

		md, _ := metadata.FromIncomingContext(ctx) // nil check in continueFromGrpcMetadata
		span := sentry.StartSpan(ctx, "grpc.server", continueFromGrpcMetadata(md), sentry.WithTransactionName(info.FullMethod))
		ctx = span.Context()
		defer span.Finish()

		reqBytes, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		hub.Scope().SetRequestBody(reqBytes)

		resp, err := handler(ctx, req)
		if err != nil && o.ReportOn(err) {
			tags := grpcTags.Extract(ctx)
			for k, v := range tags.Values() {
				hub.Scope().SetTag(k, v.(string))
			}

		}
		span.Status = toSpanStatus(status.Code(err))

		return resp, err
	}
}

// GrpcLocalize create localize base on request language for multilingual
func GrpcLocalize(bundle *languageLocalize.I18n) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		lang := "en"

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			lang = bundle.GetLanguageFromMD(md)
		}

		ctx = context.WithValue(ctx, "localize", bundle.GetLocalize(lang))

		return handler(ctx, req)
	}
}

func zapLevel(code codes.Code) zapcore.Level {
	if code == codes.OK {
		return zap.DebugLevel
	}
	return grpcZap.DefaultCodeToLevel(code)
}

func isPermissionZero(permissions []int32) bool {
	for _, perm := range permissions {
		return 0 == perm
	}
	return false
}
