package jwt

import (
	"errors"
)

var (
	ErrPublicSecretIsEmpty = errors.New("jwt: public secret key is empty")
	ErrParseWithClaims     = errors.New("jwt: failed to parse with claims")
	ErrJwtTokenIsInvalid   = errors.New("jwt: token is invalid")
	ErrTokenExpired        = errors.New("jwt: token expired")
	ErrSignatureInvalid    = errors.New("jwt: token signature is invalid")
	ErrInvalidToken        = errors.New("jwt: token is invalid")
)
