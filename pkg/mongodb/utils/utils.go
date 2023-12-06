package utils

import "go.mongodb.org/mongo-driver/bson/primitive"

// ConvertStringToObjectID convert string to objectID
func ConvertStringToObjectID(id string) primitive.ObjectID {
	objId, _ := primitive.ObjectIDFromHex(id)
	return objId
}

// ConvertStringsToObjectID convert pagination strings to objectIDs
func ConvertStringsToObjectID(ids ...string) []primitive.ObjectID {
	objectIDs := make([]primitive.ObjectID, 0)
	for _, id := range ids {
		objId, _ := primitive.ObjectIDFromHex(id)
		objectIDs = append(objectIDs, objId)
	}
	return objectIDs
}
