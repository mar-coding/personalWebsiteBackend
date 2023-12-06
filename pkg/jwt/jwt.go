package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type SecretType uint

const (
	PUBLIC  = iota // PUBLIC check jwt token with public secret key
	PRIVATE        // PRIVATE check jwt token with private secret key
)

var (
	_defaultAccessTokenDuration  = 24 * time.Hour
	_defaultRefreshTokenDuration = 24 * time.Hour * 30
)

type JWT struct {
	publicSecretKey, privateSecretKey []byte
	subject                           string
	issuer                            string
	accessTokenExpireDuration         time.Duration
	refreshTokenExpireDuration        time.Duration
}

type Payload struct {
	jwt.RegisteredClaims
	UserId        primitive.ObjectID `bson:"u_id" json:"u_id"`
	SessionId     primitive.ObjectID `bson:"sess_id" json:"sess_id"`
	Permissions   map[int32][]int32  `bson:"perms,omitempty" json:"perms,omitempty"`
	Roles         []string           `bson:"roles,omitempty" json:"roles,omitempty"`
	IsPrivate     bool               `bson:"is_p,omitempty" json:"is_p,omitempty"`
	Validated     bool               `bson:"valid,omitempty" json:"valid,omitempty"`
	ExtraData     map[string]any     `bson:"e_data,omitempty" json:"e_data,omitempty"`
	ApplicationId string             `bson:"application_id,omitempty" json:"applicationId,omitempty"`
}

// New create new jwt object with public and private secret key
func New(publicSecretKey, privateSecretKey string, opts ...Option) (*JWT, error) {
	jwtToken := new(JWT)

	jwtToken.accessTokenExpireDuration = _defaultAccessTokenDuration
	jwtToken.refreshTokenExpireDuration = _defaultRefreshTokenDuration

	if len(publicSecretKey) == 0 {
		return nil, ErrPublicSecretIsEmpty
	}

	jwtToken.publicSecretKey = []byte(publicSecretKey)
	jwtToken.privateSecretKey = []byte(privateSecretKey)

	for _, opt := range opts {
		opt(jwtToken)
	}

	return jwtToken, nil
}

// CreateAccessToken generate new access token with claim payload
func (j *JWT) CreateAccessToken(payload *Payload) (string, error) {
	expireAt := new(jwt.NumericDate)
	expireAt.Time = time.Now().Add(j.accessTokenExpireDuration)
	payload.ExpiresAt = expireAt

	return j.createToken(PUBLIC, payload)
}

// CreateRefreshToken generate new refresh token with claim payload
func (j *JWT) CreateRefreshToken(payload *Payload) (string, error) {
	expireAt := new(jwt.NumericDate)
	expireAt.Time = time.Now().Add(j.refreshTokenExpireDuration)
	payload.ExpiresAt = expireAt

	return j.createToken(PUBLIC, payload)
}

// CreatePrivateAccessToken generate new private access token with claim payload
func (j *JWT) CreatePrivateAccessToken(payload *Payload) (string, error) {
	expireAt := new(jwt.NumericDate)
	expireAt.Time = time.Now().Add(j.accessTokenExpireDuration)
	payload.ExpiresAt = expireAt

	return j.createToken(PRIVATE, payload)
}

// ParseJwtToken parse jwt token with public or private secret key
func (j *JWT) ParseJwtToken(jwtToken string, secretType SecretType) (*Payload, error) {
	secret := make([]byte, 0)

	switch secretType {
	case PUBLIC:
		secret = j.publicSecretKey
	case PRIVATE:
		secret = j.privateSecretKey
	}

	token, err := jwt.ParseWithClaims(jwtToken, &Payload{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if errors.Is(err, jwt.ErrTokenExpired) {
		return nil, ErrTokenExpired
	}

	if errors.Is(err, jwt.ErrSignatureInvalid) {
		return nil, ErrSignatureInvalid
	}

	if err != nil {
		return nil, ErrParseWithClaims
	}

	if claims, ok := token.Claims.(*Payload); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrJwtTokenIsInvalid
}

// IsValidJwtToken check token is valid
func IsValidJwtToken(jwtToken string, secretKey string) (bool, error) {
	token, err := new(jwt.Parser).Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if errors.Is(err, jwt.ErrTokenExpired) {
		return false, ErrTokenExpired
	}

	if errors.Is(err, jwt.ErrSignatureInvalid) {
		return false, ErrSignatureInvalid
	}

	if err != nil {
		return false, ErrInvalidToken
	}

	if !token.Valid {
		return false, ErrInvalidToken
	}
	return true, nil
}

// ParseUnverifiedJwtToken parse token without verify by secret key
func ParseUnverifiedJwtToken(jwtToken string) (*Payload, error) {
	parser := new(jwt.Parser)

	token, _, err := parser.ParseUnverified(jwtToken, &Payload{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Payload); ok {
		return claims, nil
	}
	return nil, errors.New("jwt: token is invalid")
}

func (j *JWT) createToken(secretType SecretType, payload *Payload) (string, error) {
	if len(j.issuer) != 0 {
		issueAt := new(jwt.NumericDate)
		issueAt.Time = time.Now()

		payload.Issuer = j.issuer
		payload.IssuedAt = issueAt
	}

	if len(j.subject) != 0 {
		payload.Subject = j.subject
	}

	secret := make([]byte, 0)
	switch secretType {
	case PUBLIC:
		secret = j.publicSecretKey
	case PRIVATE:
		secret = j.privateSecretKey
	}

	if secretType == PUBLIC {
		secret = j.publicSecretKey
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return claims.SignedString(secret)
}

// IsTokenExpired check token expired with time.Now
func (p *Payload) IsTokenExpired() bool {
	return p.ExpiresAt.Time.Before(time.Now())
}

// GetUserID return user id
func (p *Payload) GetUserID() primitive.ObjectID {
	return p.UserId
}

// GetSessionID return user session id
func (p *Payload) GetSessionID() primitive.ObjectID {
	return p.SessionId
}

// GetServicePermissions return slice of service permissions by service code
func (p *Payload) GetServicePermissions(serviceCode int32) []int32 {
	return p.Permissions[serviceCode]
}

// GetRoles return list of user roles
func (p *Payload) GetRoles() []string {
	return p.Roles
}
