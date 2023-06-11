package temporal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ProtoJSONPayloadConverter converts proto objects to/from JSON.
type ProtoJSONPayloadConverter struct {
	unmarshalOpts protojson.UnmarshalOptions
	marshalOpts   protojson.MarshalOptions
}

var (
	jsonNil, _ = json.Marshal(nil)
)

// NewProtoJSONPayloadConverter creates new instance of `ProtoJSONPayloadConverter`.
func NewProtoJSONPayloadConverter() *ProtoJSONPayloadConverter {
	return &ProtoJSONPayloadConverter{
		unmarshalOpts: protojson.UnmarshalOptions{
			AllowPartial:   true,
			DiscardUnknown: true,
		},
		marshalOpts: protojson.MarshalOptions{
			Multiline:    false,
			AllowPartial: true,
		},
	}
}

func newPayload(data []byte, c converter.PayloadConverter) *commonpb.Payload {
	return &commonpb.Payload{
		Metadata: map[string][]byte{
			converter.MetadataEncoding: []byte(c.Encoding()),
		},
		Data: data,
	}
}

func newProtoPayload(data []byte, c *ProtoJSONPayloadConverter, messageType string) *commonpb.Payload {
	return &commonpb.Payload{
		Metadata: map[string][]byte{
			converter.MetadataEncoding:    []byte(c.Encoding()),
			converter.MetadataMessageType: []byte(messageType),
		},
		Data: data,
	}
}

func pointerTo(val interface{}) reflect.Value {
	valPtr := reflect.New(reflect.TypeOf(val))
	valPtr.Elem().Set(reflect.ValueOf(val))
	return valPtr
}

func newOfSameType(val reflect.Value) reflect.Value {
	valType := val.Type().Elem()     // is value type (i.e. commonpb.WorkflowType)
	newValue := reflect.New(valType) // is of pointer type (i.e. *commonpb.WorkflowType)
	val.Set(newValue)                // set newly created value back to passed value
	return newValue
}

func isInterfaceNil(i interface{}) bool {
	v := reflect.ValueOf(i)
	return i == nil || (v.Kind() == reflect.Ptr && v.IsNil())
}

// ToPayload converts single proto value to payload.
func (c *ProtoJSONPayloadConverter) ToPayload(value interface{}) (*commonpb.Payload, error) {
	if isInterfaceNil(value) {
		return newPayload(jsonNil, c), nil
	}

	builtPointer := false
	for {
		if valueMap, ok := value.(map[string]interface{}); ok {
			buf, err := json.Marshal(valueMap)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", converter.ErrUnableToEncode, err)
			}
			return newProtoPayload(buf, c, "Unknown"), nil

		}
		if valueProto, ok := value.(proto.Message); ok {
			buf, err := c.marshalOpts.Marshal(valueProto)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", converter.ErrUnableToEncode, err)
			}
			return newProtoPayload(buf, c, string(valueProto.ProtoReflect().Descriptor().FullName())), nil
		}
		if builtPointer {
			break
		}
		value = pointerTo(value).Interface()
		builtPointer = true
	}

	return nil, nil
}

// FromPayload converts single proto value from payload.
func (c *ProtoJSONPayloadConverter) FromPayload(payload *commonpb.Payload, valuePtr interface{}) error {
	originalValue := reflect.ValueOf(valuePtr)
	if originalValue.Kind() != reflect.Ptr {
		return fmt.Errorf("type: %T: %w", valuePtr, converter.ErrValuePtrIsNotPointer)
	}

	originalValue = originalValue.Elem()
	if !originalValue.CanSet() {
		return fmt.Errorf("type: %T: %w", valuePtr, converter.ErrUnableToSetValue)
	}

	if bytes.Equal(payload.GetData(), jsonNil) {
		originalValue.Set(reflect.Zero(originalValue.Type()))
		return nil
	}

	if originalValue.Kind() == reflect.Interface {
		return fmt.Errorf("value type: %s: %w", originalValue.Type().String(), converter.ErrValuePtrMustConcreteType)
	}

	value := originalValue
	// If original value is of value type (i.e. commonpb.WorkflowType), create a pointer to it.
	if originalValue.Kind() != reflect.Ptr {
		value = pointerTo(originalValue.Interface())
	}

	protoValue := value.Interface() // protoValue is for sure of pointer type (i.e. *commonpb.WorkflowType).

	protoMessage, isProtoMessage := protoValue.(proto.Message)
	if !isProtoMessage {
		return fmt.Errorf("type: %T: %w", protoValue, converter.ErrTypeNotImplementProtoMessage)
	}

	// If original value is nil, create new instance.
	if originalValue.Kind() == reflect.Ptr && originalValue.IsNil() {
		value = newOfSameType(originalValue)
		protoValue = value.Interface()
		protoMessage = protoValue.(proto.Message)
	}

	err := protojson.Unmarshal(payload.GetData(), protoMessage)

	// If original value wasn't a pointer then set value back to where valuePtr points to.
	if originalValue.Kind() != reflect.Ptr {
		originalValue.Set(value.Elem())
	}

	if err != nil {
		return fmt.Errorf("%w: %v", converter.ErrUnableToDecode, err)
	}

	return nil
}

// ToString converts payload object into human readable string.
func (c *ProtoJSONPayloadConverter) ToString(payload *commonpb.Payload) string {
	return string(payload.GetData())
}

// Encoding returns MetadataEncodingProtoJSON.
func (c *ProtoJSONPayloadConverter) Encoding() string {
	return converter.MetadataEncodingProtoJSON
}
