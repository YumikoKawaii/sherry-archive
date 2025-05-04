package gatewayopt

import (
	"reflect"
	"strings"

	//nolint: staticcheck
	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// JSONPbMarshaler is a Marshaler that translates GRPC response to HTTP JSON response that conforms standard.
// https://confluence.teko.vn/display/TTD/Standardise+Backend+APIs+response
// Example:
// message ListSegmentsResponse {
//	 repeated Segment segments = 1 [(gogoproto.moretags)="response_field:\"data\""];
//   string message = 2;
//	 int32 code = 3;
// }
// will be translated to JSON message:
// {
//	 "data": {
//	   "segments": [...]
//	 },
//   "message": "a message",
//	 "code": 200
// }
type JSONPbMarshaler struct {
	*runtime.JSONPb
}

func (j JSONPbMarshaler) Marshal(v interface{}) ([]byte, error) {
	protoMsg, ok := v.(proto.Message)
	if !ok {
		return j.JSONPb.Marshal(v)
	}

	type responseData struct {
		Name      string
		JSONTag   string
		OmitEmpty bool
	}

	respDataFields := make(map[string][]*responseData)

	respMap := make(map[string]interface{})

	s := reflect.ValueOf(protoMsg).Elem()
	for i := 0; i < s.NumField(); i++ {
		value := s.Field(i)
		valueField := s.Type().Field(i)
		if strings.HasPrefix(valueField.Name, "XXX_") {
			continue
		}

		// this is not a protobuf field
		if valueField.Tag.Get("protobuf") == "" && valueField.Tag.Get("protobuf_oneof") == "" {
			continue
		}

		jsonTag := findProtobufJSONTag(valueField.Tag.Get("protobuf"))
		if jsonTag == "" {
			jsonTag = valueField.Tag.Get("json")
		}

		var omitEmpty bool
		if strings.HasSuffix(jsonTag, ",omitempty") {
			omitEmpty = true
			jsonTag = strings.TrimSuffix(jsonTag, ",omitempty")
		}

		if _, ok := value.Interface().(proto.Message); ok && value.IsNil() {
			// to avoid err "Marshal called with nil"
			// https://github.com/golang/protobuf/blob/v1.4.3/jsonpb/encode.go#L88
			respMap[jsonTag] = nil
		} else {
			respMap[jsonTag] = value.Interface()
		}

		responseFieldName := valueField.Tag.Get("response_field")
		if responseFieldName != "" {
			// skip if tag name is same as response attribute json tag
			if _, ok := respMap[responseFieldName]; ok {
				continue
			}
			respDataFields[responseFieldName] = append(respDataFields[responseFieldName], &responseData{
				Name:      valueField.Name,
				JSONTag:   jsonTag,
				OmitEmpty: omitEmpty,
			})
		}
	}

	for name, fields := range respDataFields {
		wrappedData := make(map[string]interface{})
		for _, field := range fields {
			wrappedData[field.JSONTag] = respMap[field.JSONTag]
			delete(respMap, field.JSONTag)
		}
		respMap[name] = wrappedData
	}

	return j.JSONPb.Marshal(respMap)
}

// findProtobufJSONTag find json tag defined in protobuf tag
// Eg: input tag `protobuf:"varint,1,opt,name=is_valid,json=isValid,proto3"`, output: isValid
func findProtobufJSONTag(protobufTag string) string {
	prefix := "json="
	tags := strings.Split(protobufTag, ",")

	for _, t := range tags {
		tt := strings.TrimSpace(t)
		if strings.HasPrefix(tt, prefix) {
			return strings.TrimSpace(tt[len(prefix):])
		}
	}
	return ""
}
