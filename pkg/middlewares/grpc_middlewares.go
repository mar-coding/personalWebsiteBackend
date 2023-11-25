package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/jhump/protoreflect/desc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log/slog"
	"strings"
)

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
	return grpc_middleware.WithUnaryServerChain(
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

// Deprecated: MethodDescriptors save methods descriptors into context for any methods, please check middleware GrpcJwtMiddleware
func MethodDescriptors(descriptors map[string]*desc.MethodDescriptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md := descriptors[info.FullMethod]
		ctx = metadata.AppendToOutgoingContext(context.WithValue(ctx, "desc", md))
		return handler(ctx, req)
	}
}

// GrpcRecovery recovery panics
func GrpcRecovery(logger logger.Logger) grpc.UnaryServerInterceptor {
	rec := func(p interface{}) (err error) {
		err = status.Errorf(codes.Unknown, "%v", p)
		logger.Error(true, "recovery: panic triggered", "error", err)
		return
	}
	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(rec),
	}
	return grpc_recovery.UnaryServerInterceptor(opts...)
}

// GrpcValidator validate your message fields, for user validator please check https://github.com/envoyproxy/protoc-gen-validate
func GrpcValidator(errHandler errHandler.Handler) grpc.UnaryServerInterceptor {
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

// Deprecated: GrpcJwt middleware for check jwt, please use GrpcJwtMiddleware
func GrpcJwt(permissionDescriptor PermissionDescriptor, serviceInfo *info.ServiceInfo, errHandler errHandler.Handler, jwtPublicSecret, jwtPrivateSecret, captchaSecretKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		permissions, optional, validate, captcha := permissionDescriptor(ctx)

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

// GrpcJwtMiddleware middleware for check jwt token
func GrpcJwtMiddleware(permissionDescriptor PermissionFunc, serviceInfo *info.ServiceInfo, errHandler errHandler.Handler, jwtPublicSecret, jwtPrivateSecret, captchaSecretKey string) grpc.UnaryServerInterceptor {
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

func notOptionalAclContext(ctx context.Context, serviceInfo *info.ServiceInfo, errHandler errHandler.Handler, jwtPublicSecret, jwtPrivateSecret string, validate, optional bool, permissions []int32) (*acl.AclJwt, error) {
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
		span := sentry.StartSpan(ctx, "grpc.server", continueFromGrpcMetadata(md), sentry.TransactionName(info.FullMethod))
		ctx = span.Context()
		defer span.Finish()

		reqBytes, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		hub.Scope().SetRequestBody(reqBytes)

		resp, err := handler(ctx, req)
		if err != nil && o.ReportOn(err) {
			tags := grpc_tags.Extract(ctx)
			for k, v := range tags.Values() {
				hub.Scope().SetTag(k, v.(string))
			}

		}
		span.Status = toSpanStatus(status.Code(err))

		return resp, err
	}
}

// GrpcLocalizer create localizer base on request language for multilingual
func GrpcLocalizer(bundle *i18n.I18n) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		lang := "en"

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			lang = bundle.GetLanguageFromMD(md)
		}

		ctx = context.WithValue(ctx, "localizer", bundle.GetLocalizer(lang))

		return handler(ctx, req)
	}
}

func zapLevel(code codes.Code) zapcore.Level {
	if code == codes.OK {
		return zap.DebugLevel
	}
	return grpc_zap.DefaultCodeToLevel(code)
}

func isPermissionZero(permissions []int32) bool {
	for _, perm := range permissions {
		return 0 == perm
	}
	return false
}
