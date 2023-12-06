package acl

import (
	"context"
	"github.com/mar-coding/personalWebsiteBackend/pkg/jwt"
)

type JwtController struct {
	Controller
	JwtPayload *jwt.Payload
	token      string
}

const _defaultAclContextKey = "acl"

// NewWithJwt create acl object for control jwt token
func NewWithJwt(serviceCode int32, jwtToken string, options ...Option) (*JwtController, error) {
	return newWithJwt(serviceCode, jwtToken, applyOption(options...))
}

func newWithJwt(serviceCode int32, jwtToken string, config *Config) (*JwtController, error) {
	aclJwt := &JwtController{
		token: jwtToken,
	}

	payload, err := jwt.ParseUnverifiedJwtToken(jwtToken)
	if err != nil {
		return nil, ErrJwtTokenIsInvalid
	}

	aclJwt.JwtPayload = payload

	secretKey := config.publicSecretKey
	if payload.IsPrivate {
		secretKey = config.privateSecretKey
	}

	if len(secretKey) == 0 {
		return nil, ErrSecretKeyIsEmpty
	}

	if valid, err := jwt.IsValidJwtToken(jwtToken, secretKey); !valid {
		return nil, err
	}

	if acl, err := newAcl(serviceCode, payload.UserId, payload.SessionId, payload.Permissions, payload.Roles, config); err != nil {
		return nil, err
	} else {
		aclJwt.Controller = acl
	}

	return aclJwt, nil
}

// GetAclJwtFromContext get acl jwt from context
func GetAclJwtFromContext(ctx context.Context) (*JwtController, error) {
	if jwtToken, ok := ctx.Value(_defaultAclContextKey).(*JwtController); ok {
		return jwtToken, nil
	}
	return nil, ErrNotFoundAclJwtObject
}

// SetAclJwtToContext set acl jwt into context
func (a *JwtController) SetAclJwtToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, _defaultAclContextKey, a)
}

// GetExtraData get extra data from jwtData
func (a *JwtController) GetExtraData() map[string]interface{} {
	return a.JwtPayload.ExtraData
}

// GetApplicationId return applicationId of jwt payload
func (a *JwtController) GetApplicationId() string {
	return a.JwtPayload.ApplicationId
}
