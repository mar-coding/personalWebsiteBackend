package acl

import (
	"context"
	"google.golang.org/grpc/metadata"
	"strings"
)

const (
	_defaultAccessTokenContextKey = "authorization"
	_defaultApiKeyContextKey      = "key"
)

// GetAclContext return acl object from context
func GetAclContext(ctx context.Context) (AclContext, error) {
	acl, err := getAclFromContext(ctx)
	if err != nil {
		aclJwt, err := GetAclJwtFromContext(ctx)
		if err != nil {
			return AclContext{}, err
		}
		return AclContext{ctx, aclJwt}, nil
	}
	return AclContext{ctx, acl}, nil
}

// GetBearerTokenFromGrpcContext get jwt token from authorization key in header
func GetBearerTokenFromGrpcContext(ctx context.Context) (string, error) {
	foundedHeaders, err := extractHeaderFromContext(ctx, _defaultAccessTokenContextKey)
	if err != nil {
		return "", err
	}
	if len(foundedHeaders) != 1 {
		return "", ErrNotFoundJwtTokenInHeader
	}

	auth := foundedHeaders[0]

	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return "", ErrMissingBearerPrefixInToken
	}
	token := strings.TrimPrefix(auth, prefix)
	return token, nil
}

// GetApiKeyFromContext from request header by field key
func GetApiKeyFromContext(ctx context.Context, apiKeyHeaderName string) (string, error) {
	if len(apiKeyHeaderName) == 0 {
		apiKeyHeaderName = _defaultApiKeyContextKey
	}
	foundedHeaders, err := extractHeaderFromContext(ctx, apiKeyHeaderName)
	if err != nil {
		return "", err
	}
	if len(foundedHeaders) != 1 {
		return "", ErrNotFoundApiKeyInHeader
	}

	apiKey := foundedHeaders[0]
	return apiKey, nil
}

func extractHeaderFromContext(ctx context.Context, header string) ([]string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrNoHeaderInRequest
	}

	foundedHeaders, ok := md[header]
	if !ok {
		return nil, ErrNoHeaderInRequest
	}

	return foundedHeaders, nil
}

func getAclFromContext(ctx context.Context) (*Acl, error) {
	acl, ok := ctx.Value(_defaultAclContextKey).(*Acl)
	if !ok {
		return nil, ErrNotFoundAclObject
	}
	return acl, nil
}
