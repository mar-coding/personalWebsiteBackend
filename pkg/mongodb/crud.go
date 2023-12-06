package mongodb

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"strings"
)

// UpdateFields update specific fields in a document
func UpdateFields(ctx context.Context, collection *mongo.Collection, searchFilter bson.M, document any, fields ...string) error {
	fieldsForUpdate, err := getBsonFieldsData(document, fields...)
	if err != nil {
		return err
	}

	_, err = collection.UpdateOne(ctx, searchFilter, bson.D{{"$set",
		fieldsForUpdate,
	}})

	return err
}

// PushItemsToArrayField arrayFieldName like "access_types" or when is nested array must be like "user.$.roles"
func PushItemsToArrayField(ctx context.Context, collection *mongo.Collection, search bson.M, arrayFieldName string, items ...interface{}) error {
	var err error
	if len(items) == 1 {
		_, err = collection.UpdateOne(ctx, search, bson.M{"$push": bson.M{arrayFieldName: items[0]}})
	} else if len(items) > 1 {
		_, err = collection.UpdateOne(ctx, search, bson.M{"$push": bson.M{arrayFieldName: bson.M{"$each": items}}})
	}
	return err
}

func getBsonFieldsData(document any, updateFields ...string) (bson.D, error) {
	if len(updateFields) == 0 {
		return nil, errors.New("mongodb: update fields is empty")
	}

	documentType := reflect.TypeOf(document).Elem()
	updateData := bson.D{}

	for _, updateFieldName := range updateFields {
		bsonFieldName, err := getBsonField(documentType, updateFieldName)
		if err != nil {
			return nil, err
		}

		updateValue := reflect.ValueOf(document).Elem().FieldByName(updateFieldName)
		updateData = append(updateData, primitive.E{Key: bsonFieldName, Value: updateValue.Interface()})
	}
	return updateData, nil
}

func getBsonField(documentType reflect.Type, fieldName string) (string, error) {
	bsonFieldName := fieldName
	updateField, hasField := documentType.FieldByName(fieldName)
	if !hasField {
		return "", errors.New(fmt.Sprintf("mongodb: failed to get field %s from document", fieldName))
	}
	bsonTag := strings.Split(updateField.Tag.Get("bson"), ";")
	if len(bsonTag) > 0 {
		bsonTagData := strings.Split(bsonTag[0], ",")
		bsonFieldName = bsonTagData[0]
	}
	return bsonFieldName, nil
}
