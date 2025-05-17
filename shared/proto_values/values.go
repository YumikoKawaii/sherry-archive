package proto_values

import "github.com/gogo/protobuf/types"

func StringValueToPointer(value *types.StringValue) *string {
	if value == nil {
		return nil
	}
	return &value.Value
}

func StringPointerToValue(value *string) *types.StringValue {
	if value == nil {
		return nil
	}
	return &types.StringValue{
		Value: *value,
	}
}

func UInt32ValueToPointer(value *types.UInt32Value) *uint32 {
	if value == nil {
		return nil
	}
	return &value.Value
}

func UInt32PointerToValue(value *uint32) *types.UInt32Value {
	if value == nil {
		return nil
	}
	return &types.UInt32Value{
		Value: *value,
	}
}

func UInt64ValueToPointer(value *types.UInt64Value) *uint64 {
	if value == nil {
		return nil
	}
	return &value.Value
}

func UInt64PointerToValue(value *uint64) *types.UInt64Value {
	if value == nil {
		return nil
	}
	return &types.UInt64Value{
		Value: *value,
	}
}

func StringToPointer(value string) *string {
	return &value
}
