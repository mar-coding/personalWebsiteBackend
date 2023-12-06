package helper

import (
	"encoding/json"
	"errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"reflect"
)

// ConvertProtoToModel convert proto structure to model
func ConvertProtoToModel[T any](proto proto.Message) (model T, err error) {
	m, ok := reflect.New(reflect.TypeOf(model).Elem()).Interface().(T)
	if !ok {
		return model, errors.New("model type is invalid")
	}
	b, err := protojson.MarshalOptions{UseEnumNumbers: true}.Marshal(proto)
	if err != nil {
		return m, err
	}
	if err = json.Unmarshal(b, m); err != nil {
		return m, err
	}
	return m, nil
}

// ConvertModelToProto model structure to proto
func ConvertModelToProto[T proto.Message](model any) (protoModel T, err error) {
	m, ok := reflect.New(reflect.TypeOf(protoModel).Elem()).Interface().(T)
	if !ok {
		return protoModel, errors.New("proto type is invalid")
	}
	b, err := json.Marshal(model)
	if err != nil {
		return m, err
	}
	if err = (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(b, m); err != nil {
		return m, err
	}
	return m, nil
}

// ConvertModelsToProto converts a slice of models to a slice of proto messages.
func ConvertModelsToProto[T proto.Message, M any](models []M) ([]T, error) {
	protoModels := make([]T, len(models))

	for i, model := range models {
		protoModel, err := ConvertModelToProto[T](model)
		if err != nil {
			return nil, err
		}
		protoModels[i] = protoModel
	}
	return protoModels, nil
}
