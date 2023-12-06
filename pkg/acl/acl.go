package acl

import (
	"context"
	"github.com/mar-coding/personalWebsiteBackend/pkg/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/metadata"
)

const _defaultServiceContextKey = "s_id"

// Controller is acl interface that indicate methods to be implemented for auth type
type Controller interface {
	HasAccess(permissionCode int32) bool
	HasAccessInOtherService(permissionCode int32, serviceCode int32) bool
	NotHasAccess(permissionCode int32) bool
	HasAnyPermissionsAccess(permissionCodes ...int32) bool
	HasAllPermissionsAccess(permissionCodes ...int32) bool
	GetPrivateToken(extraData map[string]any) (string, error)
	SetPrivateTokenToOutgoingContext(ctx context.Context, serviceId primitive.ObjectID, extraData map[string]any) (context.Context, error)
	SetAclToContext(ctx context.Context) context.Context
	GetUserID() primitive.ObjectID
	GetSessionID() primitive.ObjectID
	GetRoles() []string
}

type Acl struct {
	userPermissions      map[int32][]int32
	config               *Config
	validatedPermissions bool
	privateJwtToken      string
	userId               primitive.ObjectID
	sessionId            primitive.ObjectID
	roles                []string
}

// New create new acl object
func New(serviceCode int32, userId, sessionId primitive.ObjectID, userPermissions map[int32][]int32, roles []string, options ...Option) (Controller, error) {
	return newAcl(serviceCode, userId, sessionId, userPermissions, roles, applyOption(options...))
}

func newAcl(serviceCode int32, userId, sessionId primitive.ObjectID, userPermissions map[int32][]int32, roles []string, config *Config) (Controller, error) {
	if serviceCode == 0 {
		return nil, ErrServiceCodeNotSet
	}
	config.currentServiceCode = serviceCode
	acl := &Acl{userPermissions: userPermissions, userId: userId, sessionId: sessionId, roles: roles, config: config}

	return acl, nil
}

// HasAccess check has access to service method
func (acl *Acl) HasAccess(permissionCode int32) bool {
	return acl.HasAccessInOtherService(acl.config.currentServiceCode, permissionCode)
}

// HasAnyPermissionsAccess check permissions has access to one permission (or)
func (acl *Acl) HasAnyPermissionsAccess(permissionCodes ...int32) bool {
	for _, code := range permissionCodes {
		if acl.HasAccess(code) {
			return true
		}
	}
	return false
}

// HasAllPermissionsAccess check permissions has access to all permissions (and)
func (acl *Acl) HasAllPermissionsAccess(permissionCodes ...int32) bool {
	for _, code := range permissionCodes {
		if !acl.HasAccess(code) {
			return false
		}
	}
	return true
}

// HasAccessInOtherService check has access to other service method
func (acl *Acl) HasAccessInOtherService(serviceCode int32, permissionCode int32) bool {
	for _, userPermissionCode := range acl.userPermissions[serviceCode] {
		if userPermissionCode == permissionCode {
			return true
		}
	}
	return false
}

// NotHasAccess check not has access to service method
func (acl *Acl) NotHasAccess(permissionCode int32) bool {
	return !acl.HasAccess(permissionCode)
}

// GetPrivateToken get private token from acl
func (acl *Acl) GetPrivateToken(extraData map[string]any) (string, error) {
	if len(acl.config.privateSecretKey) == 0 {
		return "", ErrPrivateSecretKeyIsEmpty
	}
	if len(acl.privateJwtToken) != 0 {
		return acl.privateJwtToken, nil
	}

	j, err := jwt.New(acl.config.publicSecretKey, acl.config.privateSecretKey)
	if err != nil {
		return "", err
	}

	token, err := j.CreatePrivateAccessToken(&jwt.Payload{
		IsPrivate:   true,
		Validated:   acl.validatedPermissions,
		Permissions: acl.userPermissions,
		ExtraData:   extraData,
		SessionId:   acl.sessionId,
		UserId:      acl.userId,
	})
	if err != nil {
		return "", err
	}

	acl.privateJwtToken = token
	return token, nil
}

// SetPrivateTokenToOutgoingContext set private token in to outing context
func (acl *Acl) SetPrivateTokenToOutgoingContext(ctx context.Context, serviceId primitive.ObjectID, extraData map[string]any) (context.Context, error) {
	token, err := acl.GetPrivateToken(extraData)
	if err != nil {
		return nil, err
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md.Set("Authorization", "Bearer "+token)
	md[_defaultServiceContextKey] = append(md[_defaultServiceContextKey], serviceId.Hex())
	outCtx := metadata.NewOutgoingContext(ctx, md)
	return outCtx, nil
}

// SetAclToContext set acl to context with key acl
func (acl *Acl) SetAclToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, _defaultAclContextKey, acl)
}

// GetUserID return user id
func (acl *Acl) GetUserID() primitive.ObjectID {
	return acl.userId
}

// GetSessionID return session id
func (acl *Acl) GetSessionID() primitive.ObjectID {
	return acl.sessionId
}

// GetRoles return user roles
func (acl *Acl) GetRoles() []string {
	return acl.roles
}

func (acl *Acl) isValidPermissions(userPermissions map[int32][]int32) bool {
	for userServiceCode, userPerms := range userPermissions {
		for aclServiceCode, aclPerms := range acl.userPermissions {
			if userServiceCode == aclServiceCode {
				if !acl.isAclWithUserPermissionsMatch(aclPerms, userPerms) {
					return false
				}
			}
		}
	}
	return true
}

func (acl *Acl) isAclWithUserPermissionsMatch(aclPerms, userPerms []int32) bool {
	if len(aclPerms) != len(userPerms) {
		return false
	}
	for i := range aclPerms {
		if aclPerms[i] != userPerms[i] {
			return false
		}
	}
	return true
}
