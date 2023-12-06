package acl

import "errors"

var (
	ErrNoHeaderInRequest          = errors.New("acl: no headers in request")
	ErrNotFoundJwtTokenInHeader   = errors.New("acl: not found jwt token in header")
	ErrMissingBearerPrefixInToken = errors.New("acl: missing Bearer prefix in Authorization header")
	ErrNotFoundApiKeyInHeader     = errors.New("acl: not found api key in header")
	ErrJwtTokenIsInvalid          = errors.New("acl: jwt token is invalid")
	ErrSecretKeyIsEmpty           = errors.New("acl: secret key is empty")
	ErrNotFoundAclJwtObject       = errors.New("acl: failed to extract acl jwt object from context")
	ErrNotFoundAclObject          = errors.New("acl: not found acl object in context")
	ErrServiceCodeNotSet          = errors.New("acl: service code has been not set")
	ErrAclDataIsInvalid           = errors.New("acl: data not valid")
	ErrPrivateSecretKeyIsEmpty    = errors.New("acl: private secret key is empty")
)
