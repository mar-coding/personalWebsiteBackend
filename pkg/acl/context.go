package acl

import (
	"context"
	"errors"
)

type AclContext struct {
	context.Context
	Controller
}

func NewAclFromContext(ctx context.Context, serviceCode int32, options ...Option) (Controller, error) {
	config := applyOption(options...)

	token, err := GetBearerTokenFromGrpcContext(ctx)
	if err == nil {
		jwtAcl, err := newWithJwt(serviceCode, token, config)
		if err != nil {
			return nil, err
		}
		return jwtAcl, nil
	}

	if errors.Is(err, ErrNoHeaderInRequest) {
		_, err = GetApiKeyFromContext(ctx, config.apiKeyContextName)
		if err != nil {
			return nil, err
		}

		// TODO: return new acl API Key, return newAclFromApiKey(serviceCode, apiKey, config), nil
	}

	return nil, err
}
