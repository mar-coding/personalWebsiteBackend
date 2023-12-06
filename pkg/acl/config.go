package acl

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Option func(*Config)

	// GetUserPermissions get user permissions from user service
	GetUserPermissions func(ctx context.Context, userId primitive.ObjectID) map[int32][]int32

	// GetApiKeyInfo is function middleware for get api key information
	GetApiKeyInfo func(ctx context.Context, apiKey string) (userId primitive.ObjectID, permissions map[int32][]int32, err error)
)

type Config struct {
	publicSecretKey        string
	privateSecretKey       string
	validateAcl            bool // validateAcl when parse token validate with user service
	getApiKeyInfoFunc      GetApiKeyInfo
	getUserPermissionsFunc GetUserPermissions
	currentServiceCode     int32
	apiKeyContextName      string
}

// WithPublicSecretKey set public secret key for jwt public token
func WithPublicSecretKey(secretKey string) Option {
	return func(configs *Config) {
		configs.publicSecretKey = secretKey
	}
}

// WithPrivateSecretKey set public secret key for jwt private token
func WithPrivateSecretKey(secretKey string) Option {
	return func(configs *Config) {
		configs.privateSecretKey = secretKey
	}
}

// WithValidateACL validate jwt token or api key from user service
func WithValidateACL(validate bool) Option {
	return func(configs *Config) {
		configs.validateAcl = validate
	}
}

// WithGetAPIKeyInfoFunction set get api key info function middleware
func WithGetAPIKeyInfoFunction(getAPIKeyInfoFunc GetApiKeyInfo) Option {
	return func(configs *Config) {
		configs.getApiKeyInfoFunc = getAPIKeyInfoFunc
	}
}

// WithGetUserPermissionsFunction set get user permissions function middleware
func WithGetUserPermissionsFunction(getUserPermissionsFunc GetUserPermissions) Option {
	return func(configs *Config) {
		configs.getUserPermissionsFunc = getUserPermissionsFunc
	}
}

// WithAPIKeyCustomContextName add custom name for MD header to get api key
func WithAPIKeyCustomContextName(contextName string) Option {
	return func(config *Config) {
		config.apiKeyContextName = contextName
	}
}

func applyOption(options ...Option) *Config {
	config := &Config{}
	for _, opt := range options {
		opt(config)
	}
	return config
}
