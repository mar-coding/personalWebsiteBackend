package jwt

import (
	"time"
)

type Option func(*JWT)

func WithIssuer(issuer string) Option {
	return func(jwt *JWT) {
		jwt.issuer = issuer
	}
}

func WithSubject(subject string) Option {
	return func(jwt *JWT) {
		jwt.subject = subject
	}
}

func WithAccessTokenExpireDuration(duration time.Duration) Option {
	return func(j *JWT) {
		j.accessTokenExpireDuration = duration
	}
}

func WithRefreshTokenExpireDuration(duration time.Duration) Option {
	return func(j *JWT) {
		j.refreshTokenExpireDuration = duration
	}
}
