package serviceInfo

import "go.mongodb.org/mongo-driver/bson/primitive"

type ServiceInfo struct {
	Id            primitive.ObjectID `bson:"_id" json:"id"`
	Name          string             `bson:"name" json:"name" yaml:"name"`
	Description   string             `bson:"description" json:"description" yaml:"description"`
	Code          int32              `bson:"code" json:"code" yaml:"code"`
	Version       string             `bson:"version" json:"version" yaml:"version"`
	Permissions   []*Permission      `bson:"permissions" json:"permissions" yaml:"permissions"`
	ApplicationId string             `bson:"application_id,omitempty" json:"applicationId" yaml:"application_id"`
}

type Permission struct {
	Name        string `bson:"name" json:"name" yaml:"name"`
	Description string `bson:"description" json:"description" yaml:"description"`
	Code        int32  `bson:"code" json:"code" yaml:"code"`
}
