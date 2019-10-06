/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"
	"unsafe"

	flatbuffers "github.com/google/flatbuffers/go"
	"gopkg.in/yaml.v2"
)

// FieldType s.e.
type FieldType int

const (
	// FieldTypeUnspecified - wrong type
	FieldTypeUnspecified FieldType = iota
	// FieldTypeNested field is nested Scheme
	FieldTypeNested
	// FieldTypeInt int32
	FieldTypeInt
	// FieldTypeLong int64
	FieldTypeLong
	// FieldTypeFloat float32
	FieldTypeFloat
	// FieldTypeDouble float64
	FieldTypeDouble
	// FieldTypeString variable length
	FieldTypeString
	// FieldTypeBool s.e.
	FieldTypeBool
	// FieldTypeByte byte
	FieldTypeByte
)

var yamlFieldTypesMap = map[string]FieldType{
	"int":    FieldTypeInt,
	"long":   FieldTypeLong,
	"float":  FieldTypeFloat,
	"double": FieldTypeDouble,
	"string": FieldTypeString,
	"bool":   FieldTypeBool,
	"byte":   FieldTypeByte,
	"":       FieldTypeNested,
}

// Buffer is wrapper for FlatBuffers
type Buffer struct {
	scheme         *Scheme
	modifiedFields []*modifiedField
	tab            flatbuffers.Table
}

// Field describes a scheme field
type Field struct {
	QualifiedName string
	Name          string
	Ft            FieldType
	order         int
	IsMandatory   bool
	scheme        *Scheme // != nil for FieldTypeNested only
	isArray       bool
}

type modifiedField struct {
	value interface{}
}

// Array struct used to iterate over arrays
type Array struct {
	curIndex int
	Len      int
	field    *Field
	b        *Buffer
	uOffsetT flatbuffers.UOffsetT
}

// GetNext returns next array element if has one
func (a *Array) GetNext() (interface{}, bool) {
	if a.curIndex >= a.Len {
		return nil, false
	}
	res := a.b.getByField(a.field, a.curIndex)
	a.curIndex++
	return res, true
}

// GetAllIntf s.e.
func (a *Array) GetAllIntf() interface{} {
	res := *(*[]int64)(unsafe.Pointer(&a.b.tab.Bytes[a.uOffsetT]))
	res = res[:int(a.uOffsetT)+a.Len*flatbuffers.SizeInt64]
	// res := make([]int64, a.Len)

	// copy(res, a.b.tab.Bytes[a.uOffsetT:int(a.uOffsetT)+a.Len*flatbuffers.SizeInt64])
	// for i := 0; i < a.Len; i++ {
	// 	res[i] = a.b.getByField(a.field, i)
	// }
	return res
}

// GetAll returns filled array
func (a *Array) GetAll() []interface{} {
	res := make([]interface{}, a.Len)
	// res := *(*[]int64)(unsafe.Pointer(&a.b.tab.Bytes[a.uOffsetT]))
	// res = res[:int(a.uOffsetT)+a.Len*flatbuffers.SizeInt64]
	// res := make([]int64, a.Len)

	// copy(res, a.b.tab.Bytes[a.uOffsetT:int(a.uOffsetT)+a.Len*flatbuffers.SizeInt64])
	// for i := 0; i < a.Len; i++ {
	// 	res[i] = a.b.getByField(a.field, i)
	// }
	return res
}

// Scheme s.e.
type Scheme struct {
	Name      string
	fieldsMap map[string]*Field
	Fields    []*Field
}

// NewBuffer creates new empty Buffer
func NewBuffer(scheme *Scheme) *Buffer {
	b := &Buffer{}
	b.scheme = scheme
	return b
}

// GetInt returns int32 value by name and if the scheme contains the field and the value was set to non-nil
func (b *Buffer) GetInt(name string) (int32, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetInt32(o), true
	}
	return int32(0), false
}

// GetFloat returns float32 value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetFloat(name string) (float32, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetFloat32(o), true
	}
	return float32(0), false
}

// GetString returns string value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetString(name string) (string, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return byteSliceToString(b.tab.ByteVector(o)), true
	}
	return "", false
}

// GetLong returns int64 value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetLong(name string) (int64, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetInt64(o), true
	}
	return int64(0), false
}

// GetDouble returns float64 value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetDouble(name string) (float64, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetFloat64(o), true
	}
	return float64(0), false
}

// GetByte returns byte value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetByte(name string) (byte, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetByte(o), true
	}
	return byte(0), false
}

// GetBool returns bool value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetBool(name string) (bool, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetBool(o), true
	}
	return false, false
}

func (b *Buffer) getFieldUOffsetT(name string) flatbuffers.UOffsetT {
	if len(b.tab.Bytes) == 0 {
		return 0
	}
	if f, ok := b.scheme.fieldsMap[name]; ok {
		return b.getFieldUOffsetTByOrder(f.order)
	}
	return 0
}

func (b *Buffer) getFieldUOffsetTByOrder(order int) flatbuffers.UOffsetT {
	if len(b.tab.Bytes) == 0 {
		return 0
	}
	preOffset := flatbuffers.UOffsetT(b.tab.Offset(flatbuffers.VOffsetT((order + 2) * 2)))
	if preOffset == 0 {
		return 0
	}
	return preOffset + b.tab.Pos
}

func (b *Buffer) getByStringField(f *Field) (string, bool) {
	o := b.getFieldUOffsetTByOrder(f.order)
	if o == 0 {
		return "", false
	}
	return byteSliceToString(b.tab.ByteVector(o)), true
}

func (b *Buffer) getByField(f *Field, index int) interface{} {
	uOffsetT := b.getFieldUOffsetTByOrder(f.order)
	if uOffsetT == 0 {
		return nil
	}
	if f.isArray {
		arrayLen := b.tab.VectorLen(uOffsetT - b.tab.Pos)
		uOffsetT = b.tab.Vector(uOffsetT - b.tab.Pos)
		if index < 0 {
			return &Array{0, arrayLen, f, b, uOffsetT}
		}
		if index > arrayLen-1 {
			return nil
		}
		uOffsetT += flatbuffers.UOffsetT(index * getFBFieldSize(f.Ft))
	}
	switch f.Ft {
	case FieldTypeInt:
		return b.tab.GetInt32(uOffsetT)
	case FieldTypeLong:
		return b.tab.GetInt64(uOffsetT)
	case FieldTypeFloat:
		return b.tab.GetFloat32(uOffsetT)
	case FieldTypeDouble:
		return b.tab.GetFloat64(uOffsetT)
	case FieldTypeByte:
		return b.tab.GetByte(uOffsetT)
	case FieldTypeBool:
		return b.tab.GetBool(uOffsetT)
	case FieldTypeNested:
		res := ReadBuffer(b.tab.Bytes, f.scheme)
		res.tab.Pos = b.tab.Indirect(uOffsetT)
		return res
	default:
		return byteSliceToString(b.tab.ByteVector(uOffsetT))
	}
}

func getFBFieldSize(ft FieldType) int {
	switch ft {
	case FieldTypeBool:
		return flatbuffers.SizeBool
	case FieldTypeByte:
		return flatbuffers.SizeByte
	case FieldTypeDouble:
		return flatbuffers.SizeFloat64
	case FieldTypeFloat:
		return flatbuffers.SizeFloat32
	case FieldTypeInt:
		return flatbuffers.SizeInt32
	case FieldTypeLong:
		return flatbuffers.SizeInt64
	default:
		return flatbuffers.SizeUOffsetT
	}
}

// Get returns stored field value by name.
// nil -> field is unset or no such field in the scheme
func (b *Buffer) Get(name string) interface{} {
	f, ok := b.scheme.fieldsMap[name]
	if !ok {
		return nil
	}
	return b.getByField(f, -1)

}

// GetByIndex returns array field element by its index
// no such field, index out of bounds, array field is not set or unset -> nil
func (b *Buffer) GetByIndex(name string, index int) interface{} {
	f, ok := b.scheme.fieldsMap[name]
	if !ok || index < 0 {
		return nil
	}
	return b.getByField(f, index)
}

// ReadBuffer creates Buffer from bytes using provided Scheme
func ReadBuffer(bytes []byte, scheme *Scheme) *Buffer {
	b := NewBuffer(scheme)
	rootUOffsetT := flatbuffers.GetUOffsetT(bytes)
	b.tab.Bytes = bytes
	b.tab.Pos = rootUOffsetT
	return b
}

// Set sets field value by name.
// Byte array is not modified.
// Call ToBytes() to get modified byte array
func (b *Buffer) Set(name string, value interface{}) {
	f, ok := b.scheme.fieldsMap[name]
	if !ok {
		return
	}
	b.set(f, value)
}

func (b *Buffer) set(f *Field, value interface{}) {
	b.prepareModifiedFields()
	b.modifiedFields[f.order] = &modifiedField{value}
}

// ApplyJSONAndToBytes sets field values described by provided json and returns new FlatBuffer byte array
// resulting buffer has no value (or has nil value) for a mandatory field -> error
// value type and field type are incompatible (e.g. string for numberic field) -> error
// value and field types differs but value fits into field -> no error. Examples:
//   255 fits into float, double, int, long, byte;
//   256 does not fit into byte
//   "str" does not fit into int64 (different types)
// if unexisting field is provided -> no error (e.g. if trying to write data in new scheme into buffer in old scheme)
// arrays and nested objects are rewritten in the Buffer
// primitive fields are appended to the Buffer
func (b *Buffer) ApplyJSONAndToBytes(jsonBytes []byte) ([]byte, error) {
	dest := map[string]interface{}{}
	err := json.Unmarshal(jsonBytes, &dest)
	if err != nil {
		return nil, err
	}
	err = b.applyJSONMap(dest)
	if err != nil {
		return nil, err
	}

	return b.ToBytes()
}

func (b *Buffer) applyJSONMap(data map[string]interface{}) error {
	for fn, fv := range data {
		f, ok := b.scheme.fieldsMap[fn]
		if !ok {
			continue
		}
		if f.Ft == FieldTypeNested {
			if f.isArray {
				datasNested, ok := fv.([]interface{})
				if !ok {
					return fmt.Errorf("array of objects required but %v provided for field %s", fv, f.QualifiedName)
				}
				buffers := make([]*Buffer, len(datasNested))
				for i, dataNestedIntf := range datasNested {
					dataNested, ok := dataNestedIntf.(map[string]interface{})
					if !ok {
						return fmt.Errorf("element value of array field %s must be an object, %v provided", fn, dataNestedIntf)
					}
					buffers[i] = NewBuffer(f.scheme)
					buffers[i].applyJSONMap(dataNested)
				}
				b.Set(f.Name, buffers)
			} else {
				bNested := NewBuffer(f.scheme)
				dataNested, ok := fv.(map[string]interface{})
				if !ok {
					return fmt.Errorf("value of field %s must be an object, %v provided", fn, fv)
				}
				bNested.applyJSONMap(dataNested)
				b.Set(f.Name, bNested)
				// b.Get
			}
		} else {
			b.set(f, fv)
		}
	}
	return nil
}

// ToBytes returns new FlatBuffer byte array with fields modified by Set() and fields which initially had values
// Note: initial byte array and current modifications are kept
func (b *Buffer) ToBytes() ([]byte, error) {
	bl := flatbuffers.NewBuilder(0)
	_, err := b.encodeBuffer(bl)
	if err != nil {
		return nil, err
	}
	return bl.FinishedBytes(), nil
}

func (b *Buffer) prepareModifiedFields() {
	if len(b.modifiedFields) == 0 {
		b.modifiedFields = make([]*modifiedField, len(b.scheme.Fields))
	}
}

func (b *Buffer) encodeBuffer(bl *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	strUOffsetTs := make([]flatbuffers.UOffsetT, len(b.scheme.Fields))
	nestedBuffers := make([]flatbuffers.UOffsetT, len(b.scheme.Fields))
	arrays := make([]flatbuffers.UOffsetT, len(b.scheme.Fields))
	b.prepareModifiedFields()
	var err error

	for _, f := range b.scheme.Fields {
		if f.isArray {
			arrayUOffsetT := flatbuffers.UOffsetT(0)
			modifiedField := b.modifiedFields[f.order]
			if modifiedField != nil {
				if modifiedField.value != nil {
					if arrayUOffsetT, err = b.encodeArray(bl, f, modifiedField.value); err != nil {
						return 0, err
					}
				}
			} else {
				// copy if initially existed
				if uOffsetT := b.getFieldUOffsetTByOrder(f.order); uOffsetT != 0 {
					arrayLen := b.tab.VectorLen(uOffsetT - b.tab.Pos)
					uOffsetT = b.tab.Vector(uOffsetT - b.tab.Pos)
					arrayUOffsetT = bl.CreateByteVector(b.tab.Bytes[uOffsetT : int(uOffsetT)+arrayLen])
					// if existingArray := b.getByField(f, -1); existingArray != nil {
					// 	if arrayUOffsetT, err = b.encodeArray(bl, f, existingArray); err != nil {
					// 		return 0, err
					// 	}
					// }
				}
			}
			arrays[f.order] = arrayUOffsetT
		} else if f.Ft == FieldTypeNested {
			nestedUOffsetT := flatbuffers.UOffsetT(0)
			modifiedField := b.modifiedFields[f.order]
			if modifiedField != nil {
				if modifiedField.value != nil {
					if nestedBuffer, ok := modifiedField.value.(*Buffer); !ok {
						return 0, fmt.Errorf("nested object required but %v provided for field %s", modifiedField.value, f.QualifiedName)
					} else if nestedUOffsetT, err = nestedBuffer.encodeBuffer(bl); err != nil {
						return 0, err
					}
				}
			} else {
				if b.getFieldUOffsetTByOrder(f.order) != 0 {
					bufToWrite := b.getByField(f, -1) // can not be nil
					if nestedUOffsetT, err = bufToWrite.(*Buffer).encodeBuffer(bl); err != nil {
						return 0, err
					}
				}
			}
			nestedBuffers[f.order] = nestedUOffsetT
		} else if f.Ft == FieldTypeString {
			modifiedStringField := b.modifiedFields[f.order]
			if modifiedStringField != nil {
				if modifiedStringField.value != nil {
					if strToWrite, ok := modifiedStringField.value.(string); ok {
						strUOffsetTs[f.order] = bl.CreateString(strToWrite)
					} else {
						return 0, fmt.Errorf("string required but %v provided for field %s", modifiedStringField.value, f.QualifiedName)
					}
				}
			} else {
				if strToWrite, ok := b.getByStringField(f); ok {
					strUOffsetTs[f.order] = bl.CreateString(strToWrite)
				}
			}
		}
	}

	bl.StartObject(len(b.scheme.fieldsMap))
	for _, f := range b.scheme.Fields {
		isSet := false
		if f.isArray {
			if isSet = arrays[f.order] > 0; isSet {
				bl.PrependUOffsetTSlot(f.order, arrays[f.order], 0)
			}
		} else {
			switch f.Ft {
			case FieldTypeString:
				if isSet = strUOffsetTs[f.order] > 0; isSet {
					bl.PrependUOffsetTSlot(f.order, strUOffsetTs[f.order], 0)
				}
			case FieldTypeNested:
				if isSet = nestedBuffers[f.order] > 0; isSet {
					bl.PrependUOffsetTSlot(f.order, nestedBuffers[f.order], 0)
				}
			default:
				modifiedField := b.modifiedFields[f.order]
				if modifiedField != nil {
					if isSet = modifiedField.value != nil; isSet {
						if !encodeNonStringValue(bl, f, modifiedField.value) {
							return 0, fmt.Errorf("wrong value %v provided for field %s", modifiedField.value, f.QualifiedName)
						}
					}
				} else {
					isSet = copyNonStringValue(bl, b, f)
				}
			}
		}
		if f.IsMandatory && !isSet {
			return 0, fmt.Errorf("Mandatory field %s is not set", f.Name)
		}
	}
	res := bl.EndObject()
	bl.Finish(res)
	return res, nil
}

func (b *Buffer) encodeArray(bl *flatbuffers.Builder, f *Field, value interface{}) (flatbuffers.UOffsetT, error) {
	elemSize := getFBFieldSize(f.Ft)
	switch f.Ft {
	case FieldTypeInt:
		arr, ok := value.([]int32)
		if !ok {
			intfs, ok := value.([]interface{})
			if !ok {
				return 0, fmt.Errorf("[]int32 required but %v provided for field %s", value, f.QualifiedName)
			}
			arr = make([]int32, len(intfs))
			for i, intf := range intfs {
				switch intf.(type) {
				case float64:
					float64Src := intf.(float64)
					if !isFloat64ValueFitsIntoField(f, float64Src) {
						return 0, fmt.Errorf("[]int32 required but %v provided for field %s", value, f.QualifiedName)
					}
					arr[i] = int32(float64Src)
				case int32:
					arr[i] = intf.(int32)
				default:
					return 0, fmt.Errorf("[]int32 required but %v provided for field %s", value, f.QualifiedName)
				}
			}
		}
		bl.StartVector(elemSize, len(arr), elemSize)
		for i := len(arr) - 1; i >= 0; i-- {
			bl.PrependInt32(arr[i])
		}
		return bl.EndVector(len(arr)), nil
	case FieldTypeBool:
		arr, ok := value.([]bool)
		if !ok {
			intfs, ok := value.([]interface{})
			if !ok {
				return 0, fmt.Errorf("[]bool required but %v provided for field %s", value, f.QualifiedName)
			}
			arr = make([]bool, len(intfs))
			for i, intf := range intfs {
				boolVal, ok := intf.(bool)
				if !ok {
					return 0, fmt.Errorf("[]bool required but %v provided for field %s", value, f.QualifiedName)
				}
				arr[i] = boolVal
			}
		}
		bl.StartVector(elemSize, len(arr), elemSize)
		for i := len(arr) - 1; i >= 0; i-- {
			bl.PrependBool(arr[i])
		}
		return bl.EndVector(len(arr)), nil
	case FieldTypeLong:
		arr, ok := value.([]int64)
		if !ok {
			intfs, ok := value.([]interface{})
			if !ok {
				return 0, fmt.Errorf("[]int64 required but %v provided for field %s", value, f.QualifiedName)
			}
			arr = make([]int64, len(intfs))
			for i, intf := range intfs {
				switch intf.(type) {
				case float64:
					float64Src := intf.(float64)
					if !isFloat64ValueFitsIntoField(f, float64Src) {
						return 0, fmt.Errorf("[]int64 required but %v provided for field %s", value, f.QualifiedName)
					}
					arr[i] = int64(float64Src)
				case int64:
					arr[i] = intf.(int64)
				default:
					return 0, fmt.Errorf("[]int64 required but %v provided for field %s", value, f.QualifiedName)
				}
			}
		}
		bl.StartVector(elemSize, len(arr), elemSize)
		for i := len(arr) - 1; i >= 0; i-- {
			bl.PrependInt64(arr[i])
		}
		return bl.EndVector(len(arr)), nil
	case FieldTypeFloat:
		arr, ok := value.([]float32)
		if !ok {
			intfs, ok := value.([]interface{})
			if !ok {
				return 0, fmt.Errorf("[]float32 required but %v provided for field %s", value, f.QualifiedName)
			}
			arr = make([]float32, len(intfs))
			for i, intf := range intfs {
				switch intf.(type) {
				case float64:
					float64Src := intf.(float64)
					if !isFloat64ValueFitsIntoField(f, float64Src) {
						return 0, fmt.Errorf("[]int64 required but %v provided for field %s", value, f.QualifiedName)
					}
					arr[i] = float32(float64Src)
				case float32:
					arr[i] = intf.(float32)
				default:
					return 0, fmt.Errorf("[]float32 required but %v provided for field %s", value, f.QualifiedName)
				}
			}
		}
		bl.StartVector(elemSize, len(arr), elemSize)
		for i := len(arr) - 1; i >= 0; i-- {
			bl.PrependFloat32(arr[i])
		}
		return bl.EndVector(len(arr)), nil
	case FieldTypeDouble:
		arr, ok := value.([]float64)
		if !ok {
			intfs, ok := value.([]interface{})
			if !ok {
				return 0, fmt.Errorf("[]float32 required but %v provided for field %s", value, f.QualifiedName)
			}
			arr = make([]float64, len(intfs))
			for i, intf := range intfs {
				float64Src, ok := intf.(float64)
				if !ok {
					return 0, fmt.Errorf("[]float64 required but %v provided for field %s", value, f.QualifiedName)
				}
				arr[i] = float64Src
			}
		}
		bl.StartVector(elemSize, len(arr), elemSize)
		for i := len(arr) - 1; i >= 0; i-- {
			bl.PrependFloat64(arr[i])
		}
		return bl.EndVector(len(arr)), nil
	case FieldTypeByte:
		arr, ok := value.([]byte)
		if !ok {
			intfs, ok := value.([]interface{})
			if !ok {
				return 0, fmt.Errorf("[]byte required but %v provided for field %s", value, f.QualifiedName)
			}
			arr = make([]byte, len(intfs))
			for i, intf := range intfs {
				switch intf.(type) {
				case float64:
					float64Src := intf.(float64)
					if !isFloat64ValueFitsIntoField(f, float64Src) {
						return 0, fmt.Errorf("[]byte required but %v provided for field %s", value, f.QualifiedName)
					}
					arr[i] = byte(float64Src)
				case byte:
					arr[i] = intf.(byte)
				default:
					return 0, fmt.Errorf("[]byte required but %v provided for field %s", value, f.QualifiedName)
				}
			}
		}
		bl.StartVector(elemSize, len(arr), elemSize)
		for i := len(arr) - 1; i >= 0; i-- {
			bl.PrependByte(arr[i])
		}
		return bl.EndVector(len(arr)), nil
	case FieldTypeString:
		arr, ok := value.([]string)
		if !ok {
			intfs, ok := value.([]interface{})
			if !ok {
				return 0, fmt.Errorf("[]string required but %v provided for field %s", value, f.QualifiedName)
			}
			arr = make([]string, len(intfs))
			for i, intf := range intfs {
				stringVal, ok := intf.(string)
				if !ok {
					return 0, fmt.Errorf("[]byte required but %v provided for field %s", value, f.QualifiedName)
				}
				arr[i] = stringVal
			}
		}
		stringUOffsetTs := make([]flatbuffers.UOffsetT, len(arr))
		for i := 0; i < len(arr); i++ {
			stringUOffsetTs[i] = bl.CreateString(arr[i])
		}
		bl.StartVector(elemSize, len(arr), elemSize)
		for i := len(arr) - 1; i >= 0; i-- {
			bl.PrependUOffsetT(stringUOffsetTs[i])
		}
		return bl.EndVector(len(arr)), nil
	default:
		arr, ok := value.([]*Buffer)
		if !ok {
			return 0, fmt.Errorf("array of nested objects required but %v provided for field %s", value, f.QualifiedName)
		}
		nestedUOffsetTs := make([]flatbuffers.UOffsetT, len(arr))
		for i := 0; i < len(arr); i++ {
			nestedUOffsetT, err := arr[i].encodeBuffer(bl)
			if err != nil {
				return 0, err
			}
			nestedUOffsetTs[i] = nestedUOffsetT
		}
		bl.StartVector(elemSize, len(arr), elemSize)
		for i := len(arr) - 1; i >= 0; i-- {
			bl.PrependUOffsetT(nestedUOffsetTs[i])
		}
		return bl.EndVector(len(arr)), nil
	}
}

func copyNonStringValue(dest *flatbuffers.Builder, src *Buffer, f *Field) bool {
	offset := src.getFieldUOffsetTByOrder(f.order)
	if offset == 0 {
		return false
	}
	switch f.Ft {
	case FieldTypeInt:
		dest.PrependInt32(src.tab.GetInt32(offset))
	case FieldTypeLong:
		dest.PrependInt64(src.tab.GetInt64(offset))
	case FieldTypeFloat:
		dest.PrependFloat32(src.tab.GetFloat32(offset))
	case FieldTypeDouble:
		dest.PrependFloat64(src.tab.GetFloat64(offset))
	case FieldTypeByte:
		dest.PrependByte(src.tab.GetByte(offset))
	case FieldTypeBool:
		dest.PrependBool(src.tab.GetBool(offset))
	}
	dest.Slot(f.order)
	return true
}

func numberToFloat64(number interface{}) (res float64, ok bool) {
	ok = true
	switch number.(type) {
	case float64:
		res = number.(float64)
	case float32:
		res = float64(number.(float32))
	case int64:
		res = float64(number.(int64))
	case int32:
		res = float64(number.(int32))
	case int:
		res = float64(number.(int))
	case byte:
		res = float64(number.(byte))
	default:
		ok = false
	}
	return res, ok
}

func isFloat64ValueFitsIntoField(f *Field, float64Src float64) bool {
	if float64Src == 0 {
		return true
	}
	if float64Src == float64(int32(float64Src)) {
		if float64Src >= 0 && float64Src <= 255 {
			return f.Ft == FieldTypeInt || f.Ft == FieldTypeLong || f.Ft == FieldTypeDouble || f.Ft == FieldTypeFloat || f.Ft == FieldTypeByte
		}
		return f.Ft == FieldTypeInt || f.Ft == FieldTypeLong || f.Ft == FieldTypeDouble || f.Ft == FieldTypeFloat
	} else if float64Src == float64(int64(float64Src)) {
		return f.Ft == FieldTypeLong || f.Ft == FieldTypeDouble
	} else {
		return f.Ft == FieldTypeDouble || f.Ft == FieldTypeFloat
	}
}

func encodeNonStringValue(bl *flatbuffers.Builder, f *Field, value interface{}) bool {
	switch value.(type) {
	case bool:
		if f.Ft != FieldTypeBool {
			return false
		}
		bl.PrependBool(value.(bool))
	case float64:
		float64Src := value.(float64)
		if !isFloat64ValueFitsIntoField(f, float64Src) {
			return false
		}
		switch f.Ft {
		case FieldTypeInt:
			bl.PrependInt32(int32(float64Src))
		case FieldTypeLong:
			bl.PrependInt64(int64(float64Src))
		case FieldTypeFloat:
			bl.PrependFloat32(float32(float64Src))
		case FieldTypeDouble:
			bl.PrependFloat64(float64Src)
		default:
			bl.PrependByte(byte(float64Src))
		}
	case float32:
		if f.Ft != FieldTypeFloat {
			return false
		}
		bl.PrependFloat32(value.(float32))
	case int64:
		if f.Ft != FieldTypeLong {
			return false
		}
		bl.PrependInt64(value.(int64))
	case int32:
		if f.Ft != FieldTypeInt {
			return false
		}
		bl.PrependInt32(value.(int32))
	case byte:
		if f.Ft != FieldTypeByte {
			return false
		}
		bl.PrependByte(value.(byte))
	case int:
		intVal := value.(int)
		switch f.Ft {
		case FieldTypeInt:
			if math.Abs(float64(intVal)) > math.MaxInt32 {
				return false
			}
			bl.PrependInt32(int32(intVal))
		case FieldTypeLong:
			if math.Abs(float64(intVal)) > math.MaxInt64 {
				return false
			}
			bl.PrependInt64(int64(intVal))
		default:
			if math.Abs(float64(intVal)) > 255 {
				return false
			}
			bl.PrependByte(byte(intVal))
		}
	default:
		return false
	}
	bl.Slot(f.order)
	return true
}

// ToJSON returns JSON flat key->value string
func (b *Buffer) ToJSON() string {
	b.prepareModifiedFields()
	buf := bytes.NewBufferString("")
	e := json.NewEncoder(buf)
	buf.WriteString("{")
	for _, f := range b.scheme.Fields {
		var value interface{}
		if len(b.modifiedFields) == 0 {
			value = b.getByField(f, -1)
		} else {
			modifiedField := b.modifiedFields[f.order]
			if modifiedField != nil {
				value = modifiedField.value
			} else {
				value = b.getByField(f, -1)
			}
		}
		if value != nil {
			if arr, ok := value.(*Array); ok {
				value = arr.GetAll()
			}
			buf.WriteString("\"" + f.Name + "\":")
			if f.Ft == FieldTypeNested {
				if f.isArray {
					buf.WriteString("[")
					buffers := value.([]interface{})
					for _, bufferIntf := range buffers {
						buf.WriteString(bufferIntf.(*Buffer).ToJSON())
						buf.WriteString(",")
					}
					buf.Truncate(buf.Len() - 1)
					buf.WriteString("]")
				} else {
					buf.WriteString(value.(*Buffer).ToJSON())
				}
			} else {
				e.Encode(value)
			}
			buf.WriteString(",")
		}
	}
	if buf.Len() > 1 {
		buf.Truncate(buf.Len() - 1)
	}
	buf.WriteString("}")
	return strings.Replace(buf.String(), "\n", "", -1)
}

// NewScheme creates new empty scheme
func NewScheme() *Scheme {
	return &Scheme{"", map[string]*Field{}, []*Field{}}
}

// AddField adds field
// First letter is capital -> field is mandatory. First letter will be uncapitalized.
func (s *Scheme) AddField(name string, ft FieldType, isMandatory bool) {
	s.addField(name, ft, nil, isMandatory, false)
}

// AddArray adds array field
// First letter is capital -> field is mandatory. First letter will be uncapitalized.
func (s *Scheme) AddArray(name string, elementType FieldType, isMandatory bool) {
	s.addField(name, elementType, nil, isMandatory, true)
}

// AddNested adds nested object field
// First letter is capital -> field is mandatory. First letter will be uncapitalized.
func (s *Scheme) AddNested(name string, nested *Scheme, isMandatory bool) {
	s.addField(name, FieldTypeNested, nested, isMandatory, false)
}

// AddNestedArray adds array of nested objects field
// First letter is capital -> field is mandatory. First letter will be uncapitalized.
func (s *Scheme) AddNestedArray(name string, nested *Scheme, isMandatory bool) {
	s.addField(name, FieldTypeNested, nested, isMandatory, true)
}

func (s *Scheme) addField(name string, ft FieldType, nested *Scheme, isMandatory bool, isArray bool) {
	newField := &Field{s.getQualifiedFieldName(name), name, ft, len(s.fieldsMap), isMandatory, nested, isArray}
	s.fieldsMap[name] = newField
	s.Fields = append(s.Fields, newField)
}

// HasField returns if the Scheme contains the specified field
func (s *Scheme) HasField(name string) bool {
	_, ok := s.fieldsMap[name]
	return ok
}

// MarshalText used to conform to yaml.TextMarshaler interface
func (s *Scheme) MarshalText() (text []byte, err error) {
	return []byte(s.ToYaml()), nil
}

// func (s *Scheme) getAssignationFromJSONError(f *Field, fv interface{}) error {
// 	switch f.Ft {
// 	case FieldTypeBool:
// 		_, ok := fv.(bool)
// 		return fmt.Errorf("bool required but %v provided for field %s", fv, f.QualifiedName)
// 	case FieldTypeString:
// 		_, ok := fv.(string)
// 		return fmt.Errorf("string required but %v provided for field %s", fv, f.QualifiedName)
// 	default:
// 		float64Src, ok := fv.(float64)
// 		if !ok {
// 			return fmt.Errorf("number must be float64 only field %s", f.QualifiedName)
// 		}
// 		if float64Src == 0 {
// 			return nil
// 		}
// 		if float64Src == float64(int32(float64Src)) {
// 			if float64Src >= 0 && float64Src <= 255 {
// 				if f.Ft == FieldTypeInt || f.Ft == FieldTypeLong || f.Ft == FieldTypeDouble || f.Ft == FieldTypeFloat || f.Ft == FieldTypeByte {
// 					return nil
// 				}
// 			} else {
// 				if f.Ft == FieldTypeInt || f.Ft == FieldTypeLong || f.Ft == FieldTypeDouble || f.Ft == FieldTypeFloat {
// 					return nil
// 				}
// 			}
// 		} else if float64Src == float64(int64(float64Src)) {
// 			if f.Ft == FieldTypeLong || f.Ft == FieldTypeDouble {
// 				return nil
// 			}
// 		} else {
// 			if f.Ft == FieldTypeDouble || f.Ft == FieldTypeFloat {
// 				return nil
// 			}
// 		}
// 		return fmt.Errorf("%d does not fit into field %s", float64Src, f.QualifiedName)
// 	}
// }

// UnmarshalText is used to conform to yaml.TextMarshaler inteface
func (s *Scheme) UnmarshalText(text []byte) error {
	newS, err := YamlToScheme(byteSliceToString(text))
	if err != nil {
		return err
	}
	s.fieldsMap = newS.fieldsMap
	s.Fields = newS.Fields
	return nil
}

// ToYaml returns scheme in yaml format
func (s *Scheme) ToYaml() string {
	res := s.toYaml(0)
	return res
}

func (s *Scheme) toYaml(level int) string {
	buf := bytes.NewBufferString("")
	for _, f := range s.Fields {
		for ftStr, curFt := range yamlFieldTypesMap {
			if curFt == f.Ft {
				fieldName := f.Name
				if f.IsMandatory {
					fieldName = strings.Title(fieldName)
				}
				if f.isArray {
					fieldName = fieldName + ".."
				}
				buf.WriteString(strings.Repeat("  ", level) + fieldName + ": " + ftStr + "\n")
				if f.Ft == FieldTypeNested {
					yamlNested := f.scheme.toYaml(level + 1)
					buf.WriteString(yamlNested)
				}
				break
			}
		}
	}
	return buf.String()
}

func (s *Scheme) getQualifiedFieldName(fieldName string) string {
	if len(s.Name) > 0 {
		return s.Name + "." + fieldName
	}
	return fieldName
}

// YamlToScheme creates Scheme by provided yaml `fieldName: yamlFieldType`
// Field types:
//   - `int` -> `int32`
//   - `long` -> `int64`
//   - `float` -> `float32`
//   - `double` -> `float64`
//   - `bool` -> `bool`
//   - `string` -> `string`
//   - `byte` -> `byte`
// Field name starts with the capital letter -> field is mandatory
// Field name ends with `..` -> field is an array
// See [dynobuffers_test.go](dynobuffers_test.go) for examples
func YamlToScheme(yamlStr string) (*Scheme, error) {
	mapSlice := yaml.MapSlice{}
	err := yaml.Unmarshal([]byte(yamlStr), &mapSlice)
	if err != nil {
		return nil, err
	}
	return mapSliceToScheme(mapSlice)
}

func mapSliceToScheme(mapSlice yaml.MapSlice) (*Scheme, error) {
	scheme := NewScheme()
	for _, mapItem := range mapSlice {
		if nestedMapSlice, ok := mapItem.Value.(yaml.MapSlice); ok {
			fieldName, isMandatory, isArray := fieldPropsFromYaml(mapItem.Key.(string))
			nestedScheme, err := mapSliceToScheme(nestedMapSlice)
			if err != nil {
				return nil, err
			}
			if isArray {
				scheme.AddNestedArray(fieldName, nestedScheme, isMandatory)
			} else {
				scheme.AddNested(fieldName, nestedScheme, isMandatory)
			}
		} else if typeStr, ok := mapItem.Value.(string); ok {
			fieldName, isMandatory, isArray := fieldPropsFromYaml(mapItem.Key.(string))
			if ft, ok := yamlFieldTypesMap[typeStr]; ok {
				if isArray {
					scheme.AddArray(fieldName, ft, isMandatory)
				} else {
					scheme.AddField(fieldName, ft, isMandatory)
				}
			} else {
				return nil, errors.New("unknown field type: " + typeStr)
			}
		}
	}
	return scheme, nil
}

func fieldPropsFromYaml(name string) (fieldName string, isMandatory bool, isArray bool) {
	isMandatory = unicode.IsUpper(rune(name[0]))
	if isMandatory {
		name = strings.ToLower(name)
	}
	isArray = strings.HasSuffix(name, "..")
	if isArray {
		name = name[:len(name)-2]
	}
	fieldName = name
	return
}

// byteSliceToString converts a []byte to string without a heap allocation.
func byteSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
