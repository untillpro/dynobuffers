/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
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
	// FieldTypeObject field is nested Scheme
	FieldTypeObject
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
	"":       FieldTypeObject,
}

// Buffer is wrapper for FlatBuffers
type Buffer struct {
	Scheme         *Scheme
	modifiedFields []*modifiedField
	tab            flatbuffers.Table
}

// Field describe+s a Scheme field
type Field struct {
	QualifiedName string
	Name          string
	Ft            FieldType
	order         int
	IsMandatory   bool
	Scheme        *Scheme // != nil for FieldTypeObject only
	IsArray       bool
}

type modifiedField struct {
	value interface{}
}

// Array struct used to iterate over arrays
type Array struct {
	curElem  int
	Len      int
	field    *Field
	b        *Buffer
	uOffsetT flatbuffers.UOffsetT
	elemSize int
}

// Next iterates to the next element shich should be obtained via `Value()``
func (a *Array) Next() bool {
	a.curElem++
	return a.curElem < a.Len
}

// Value returns current array element
func (a *Array) Value() interface{} {
	if a.curElem < 0 || a.curElem >= a.Len {
		return nil
	}
	return a.b.getByUOffsetT(a.field, a.uOffsetT+flatbuffers.UOffsetT(a.curElem*a.elemSize))
}

// GetInts returns []int32
// Note: result array is a slice of underlying FlatBuffer byte array. Modifying this array means modifying result of Buffer.ToBytes() unless Buffer.Set() will be called
func (a *Array) GetInts() []int32 {
	target := a.b.tab.Bytes[a.uOffsetT:]
	res := *(*[]int32)(unsafe.Pointer(&target))
	res = res[:a.Len]
	return res
}

// GetFloats returns []float32
// Note: result array is a slice of underlying FlatBuffer byte array. Modifying this array means modifying result of Buffer.ToBytes() unless Buffer.Set() will be called
func (a *Array) GetFloats() []float32 {
	target := a.b.tab.Bytes[a.uOffsetT:]
	res := *(*[]float32)(unsafe.Pointer(&target))
	res = res[:a.Len]
	return res
}

// GetDoubles returns []float64
// Note: result array is a slice of underlying FlatBuffer byte array. Modifying this array means modifying result of Buffer.ToBytes() unless Buffer.Set() will be called
func (a *Array) GetDoubles() []float64 {
	target := a.b.tab.Bytes[a.uOffsetT:]
	res := *(*[]float64)(unsafe.Pointer(&target))
	res = res[:a.Len]
	return res
}

// GetBytes returns []byte
// Note: result array is a slice of underlying FlatBuffer byte array. Modifying this array means modifying result of Buffer.ToBytes() unless Buffer.Set() will be called
func (a *Array) GetBytes() []byte {
	return a.b.tab.Bytes[a.uOffsetT : a.Len+int(a.uOffsetT)]
}

// GetBools returns []bool
// Note: result array is a slice of underlying FlatBuffer byte array. Modifying this array means modifying result of Buffer.ToBytes() unless Buffer.Set() will be called
func (a *Array) GetBools() []bool {
	target := a.b.tab.Bytes[a.uOffsetT:]
	res := *(*[]bool)(unsafe.Pointer(&target))
	res = res[:a.Len]
	return res
}

// GetLongs returns []int64
// Note: result array is a slice of underlying FlatBuffer byte array. Modifying this array means modifying result of Buffer.ToBytes() unless Buffer.Set() will be called
func (a *Array) GetLongs() []int64 {
	target := a.b.tab.Bytes[a.uOffsetT:]
	res := *(*[]int64)(unsafe.Pointer(&target))
	res = res[:a.Len]
	return res
}

// GetStrings returns []string
func (a *Array) GetStrings() []string {
	res := make([]string, a.Len)
	for i := 0; i < a.Len; i++ {
		res[i] = a.b.getByField(a.field, i).(string)
	}
	return res
}

// GetObjects returns array of nested objects
func (a *Array) GetObjects() []*Buffer {
	res := make([]*Buffer, a.Len)
	for i := 0; i < a.Len; i++ {
		res[i] = a.b.getByField(a.field, i).(*Buffer)
	}
	return res
}

// GetAll returns untyped array
func (a *Array) GetAll() interface{} {
	switch a.field.Ft {
	case FieldTypeInt:
		return a.GetInts()
	case FieldTypeLong:
		return a.GetLongs()
	case FieldTypeFloat:
		return a.GetFloats()
	case FieldTypeDouble:
		return a.GetDoubles()
	case FieldTypeByte:
		return a.GetBytes()
	case FieldTypeBool:
		return a.GetBools()
	case FieldTypeString:
		return a.GetStrings()
	default:
		return a.GetObjects()
	}
}

// Scheme describes fields and theirs order in byte array
type Scheme struct {
	Name      string
	FieldsMap map[string]*Field
	Fields    []*Field
}

// NewBuffer creates new empty Buffer
func NewBuffer(Scheme *Scheme) *Buffer {
	b := &Buffer{}
	b.Scheme = Scheme
	return b
}

// GetInt returns int32 value by name and if the Scheme contains the field and the value was set to non-nil
func (b *Buffer) GetInt(name string) (int32, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetInt32(o), true
	}
	return int32(0), false
}

// GetFloat returns float32 value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetFloat(name string) (float32, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetFloat32(o), true
	}
	return float32(0), false
}

// GetString returns string value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetString(name string) (string, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return byteSliceToString(b.tab.ByteVector(o)), true
	}
	return "", false
}

// GetLong returns int64 value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetLong(name string) (int64, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetInt64(o), true
	}
	return int64(0), false
}

// GetDouble returns float64 value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetDouble(name string) (float64, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetFloat64(o), true
	}
	return float64(0), false
}

// GetByte returns byte value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetByte(name string) (byte, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetByte(o), true
	}
	return byte(0), false
}

// GetBool returns bool value by name and if the Scheme contains the field and if the value was set to non-nil
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
	if f, ok := b.Scheme.FieldsMap[name]; ok {
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
	if f.IsArray {
		arrayLen := b.tab.VectorLen(uOffsetT - b.tab.Pos)
		elemSize := getFBFieldSize(f.Ft)
		if isFixedSizeField(f) {
			// arrays with fixed-size elements are stored as byte arrays
			arrayLen = arrayLen / elemSize
		}
		uOffsetT = b.tab.Vector(uOffsetT - b.tab.Pos)
		if index < 0 {
			return &Array{-1, arrayLen, f, b, uOffsetT, elemSize}
		}
		if index > arrayLen-1 {
			return nil
		}
		uOffsetT += flatbuffers.UOffsetT(index * elemSize)
	}
	return b.getByUOffsetT(f, uOffsetT)
}

func (b *Buffer) getByUOffsetT(f *Field, uOffsetT flatbuffers.UOffsetT) interface{} {
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
	case FieldTypeObject:
		res := ReadBuffer(b.tab.Bytes, f.Scheme)
		res.tab.Pos = b.tab.Indirect(uOffsetT)
		if !f.IsArray {
			b.prepareModifiedFields()
			if b.modifiedFields[f.order] == nil {
				b.set(f, res)
			}
		}
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
// field is unset or no such field in the Scheme -> nil
func (b *Buffer) Get(name string) interface{} {
	f, ok := b.Scheme.FieldsMap[name]
	if !ok {
		return nil
	}
	return b.getByField(f, -1)

}

// GetByIndex returns array field element by its index
// no such field, index out of bounds, array field is not set or unset -> nil
func (b *Buffer) GetByIndex(name string, index int) interface{} {
	f, ok := b.Scheme.FieldsMap[name]
	if !ok || index < 0 {
		return nil
	}
	return b.getByField(f, index)
}

// ReadBuffer creates Buffer from bytes using provided Scheme
func ReadBuffer(bytes []byte, Scheme *Scheme) *Buffer {
	b := NewBuffer(Scheme)
	rootUOffsetT := flatbuffers.GetUOffsetT(bytes)
	b.tab.Bytes = bytes
	b.tab.Pos = rootUOffsetT
	return b
}

// Set sets field value by name.
// Underlying byte array is not modified.
// Call ToBytes() to get modified byte array
func (b *Buffer) Set(name string, value interface{}) {
	f, ok := b.Scheme.FieldsMap[name]
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
// Resulting buffer has no value (or has nil value) for a mandatory field -> error
// Value type and field type are incompatible (e.g. string for numberic field) -> error
// Vvalue and field types differs but value fits into field -> no error. Examples:
//   255 fits into float, double, int, long, byte;
//   256 does not fit into byte
//   "str" does not fit into int64 (different types)
// If unexisting field is provided -> no error (e.g. if trying to write data in new Scheme into buffer in old Scheme)
// Json data overrides data stored or mdified previously
// Byte arrays are expected to be base64 strings
// Array element is null -> error (not supported)
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
		f, ok := b.Scheme.FieldsMap[fn]
		if !ok {
			continue
		}
		if f.Ft == FieldTypeObject {
			if f.IsArray {
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
					buffers[i] = NewBuffer(f.Scheme)
					buffers[i].applyJSONMap(dataNested)
				}
				b.Set(f.Name, buffers)
			} else {
				bNested := NewBuffer(f.Scheme)
				dataNested, ok := fv.(map[string]interface{})
				if !ok {
					return fmt.Errorf("value of field %s must be an object, %v provided", fn, fv)
				}
				bNested.applyJSONMap(dataNested)
				b.Set(f.Name, bNested)
			}
		} else {
			if f.IsArray && f.Ft == FieldTypeByte {
				base64Str, ok := fv.(string)
				if !ok {
					return fmt.Errorf("base64 encoded byte array is expected for %s, %v provided", f.QualifiedName, fv)
				}
				fvNew, err := base64.StdEncoding.DecodeString(base64Str)
				if err != nil {
					return err
				}
				fv = fvNew
			}
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
		b.modifiedFields = make([]*modifiedField, len(b.Scheme.Fields))
	}
}

func (b *Buffer) encodeBuffer(bl *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	strUOffsetTs := make([]flatbuffers.UOffsetT, len(b.Scheme.Fields))
	nestedBuffers := make([]flatbuffers.UOffsetT, len(b.Scheme.Fields))
	arrays := make([]flatbuffers.UOffsetT, len(b.Scheme.Fields))
	b.prepareModifiedFields()
	var err error

	for _, f := range b.Scheme.Fields {
		if f.IsArray {
			arrayUOffsetT := flatbuffers.UOffsetT(0)
			modifiedField := b.modifiedFields[f.order]
			if modifiedField != nil {
				if modifiedField.value != nil {
					if arrayUOffsetT, err = b.encodeArray(bl, f, modifiedField.value); err != nil {
						return 0, err
					}
				}
			} else {
				// copy from source bytes if not modified and initially existed
				if uOffsetT := b.getFieldUOffsetTByOrder(f.order); uOffsetT != 0 {
					if isFixedSizeField(f) {
						// copy fixed-size array as byte array
						arrayLen := b.tab.VectorLen(uOffsetT - b.tab.Pos)
						uOffsetT = b.tab.Vector(uOffsetT - b.tab.Pos)
						arrayUOffsetT = bl.CreateByteVector(b.tab.Bytes[uOffsetT : int(uOffsetT)+arrayLen])
					} else {
						// re-encode var-size array
						if existingArray := b.getByField(f, -1); existingArray != nil {
							arrayUOffsetT, _ = b.encodeArray(bl, f, existingArray) // no errors should be here
						}
					}
				}
			}
			arrays[f.order] = arrayUOffsetT
		} else if f.Ft == FieldTypeObject {
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
					bufToWrite := b.getByField(f, -1)                         // can not be nil
					nestedUOffsetT, _ = bufToWrite.(*Buffer).encodeBuffer(bl) // no errors should be here
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

	bl.StartObject(len(b.Scheme.FieldsMap))
	for _, f := range b.Scheme.Fields {
		isSet := false
		if f.IsArray {
			if isSet = arrays[f.order] > 0; isSet {
				bl.PrependUOffsetTSlot(f.order, arrays[f.order], 0)
			}
		} else {
			switch f.Ft {
			case FieldTypeString:
				if isSet = strUOffsetTs[f.order] > 0; isSet {
					bl.PrependUOffsetTSlot(f.order, strUOffsetTs[f.order], 0)
				}
			case FieldTypeObject:
				if isSet = nestedBuffers[f.order] > 0; isSet {
					bl.PrependUOffsetTSlot(f.order, nestedBuffers[f.order], 0)
				}
			default:
				modifiedField := b.modifiedFields[f.order]
				if modifiedField != nil {
					if isSet = modifiedField.value != nil; isSet {
						if !encodeFixedSizeValue(bl, f, modifiedField.value) {
							return 0, fmt.Errorf("wrong value %v provided for field %s", modifiedField.value, f.QualifiedName)
						}
					}
				} else {
					isSet = copyFixedSizeValue(bl, b, f)
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

func intfToInt32Arr(f *Field, value interface{}) ([]int32, bool) {
	arr, ok := value.([]int32)
	if !ok {
		intfs, ok := value.([]interface{})
		if !ok {
			return nil, false
		}
		arr = make([]int32, len(intfs))
		for i, intf := range intfs {
			float64Src, ok := intf.(float64)
			if !ok || !isFloat64ValueFitsIntoField(f, float64Src) {
				return nil, false
			}
			arr[i] = int32(float64Src)
		}
	}
	return arr, true
}

func intfToBoolArr(f *Field, value interface{}) ([]bool, bool) {
	arr, ok := value.([]bool)
	if !ok {
		intfs, ok := value.([]interface{})
		if !ok {
			return nil, false
		}
		arr = make([]bool, len(intfs))
		for i, intf := range intfs {
			boolVal, ok := intf.(bool)
			if !ok {
				return nil, false
			}
			arr[i] = boolVal
		}
	}
	return arr, true
}

func intfToInt64Arr(f *Field, value interface{}) ([]int64, bool) {
	arr, ok := value.([]int64)
	if !ok {
		intfs, ok := value.([]interface{})
		if !ok {
			return nil, false
		}
		arr = make([]int64, len(intfs))
		for i, intf := range intfs {
			float64Src, ok := intf.(float64)
			if !ok || !isFloat64ValueFitsIntoField(f, float64Src) {
				return nil, false
			}
			arr[i] = int64(float64Src)
		}
	}
	return arr, true
}

func intfToFloat32Arr(f *Field, value interface{}) ([]float32, bool) {
	arr, ok := value.([]float32)
	if !ok {
		intfs, ok := value.([]interface{})
		if !ok {
			return nil, false
		}
		arr = make([]float32, len(intfs))
		for i, intf := range intfs {
			float64Src, ok := intf.(float64)
			if !ok || !isFloat64ValueFitsIntoField(f, float64Src) {
				return nil, false
			}
			arr[i] = float32(float64Src)
		}
	}
	return arr, true
}

func intfToFloat64Arr(f *Field, value interface{}) ([]float64, bool) {
	arr, ok := value.([]float64)
	if !ok {
		intfs, ok := value.([]interface{})
		if !ok {
			return nil, false
		}
		arr = make([]float64, len(intfs))
		for i, intf := range intfs {
			float64Src, ok := intf.(float64)
			if !ok {
				return nil, false
			}
			arr[i] = float64Src
		}
	}
	return arr, true
}

func (b *Buffer) encodeArray(bl *flatbuffers.Builder, f *Field, value interface{}) (flatbuffers.UOffsetT, error) {
	elemSize := getFBFieldSize(f.Ft)
	switch f.Ft {
	case FieldTypeInt:
		arr, ok := intfToInt32Arr(f, value)
		if !ok {
			return 0, fmt.Errorf("[]int32 required but %v provided for field %s", value, f.QualifiedName)
		}
		if len(arr) == 0 {
			return bl.CreateByteVector([]byte{}), nil
		}
		length := len(arr) * flatbuffers.SizeInt32
		hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(&arr[0])), Len: length, Cap: length}
		target := *(*[]byte)(unsafe.Pointer(&hdr))
		return bl.CreateByteVector(target), nil
	case FieldTypeBool:
		arr, ok := intfToBoolArr(f, value)
		if !ok {
			return 0, fmt.Errorf("[]bool required but %v provided for field %s", value, f.QualifiedName)
		}
		if len(arr) == 0 {
			return bl.CreateByteVector([]byte{}), nil
		}
		length := len(arr) * flatbuffers.SizeBool
		hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(&arr[0])), Len: length, Cap: length}
		target := *(*[]byte)(unsafe.Pointer(&hdr))
		return bl.CreateByteVector(target), nil
	case FieldTypeLong:
		arr, ok := intfToInt64Arr(f, value)
		if !ok {
			return 0, fmt.Errorf("[]int64 required but %v provided for field %s", value, f.QualifiedName)
		}
		if len(arr) == 0 {
			return bl.CreateByteVector([]byte{}), nil
		}
		length := len(arr) * flatbuffers.SizeInt64
		hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(&arr[0])), Len: length, Cap: length}
		target := *(*[]byte)(unsafe.Pointer(&hdr))
		return bl.CreateByteVector(target), nil
	case FieldTypeFloat:
		arr, ok := intfToFloat32Arr(f, value)
		if !ok {
			return 0, fmt.Errorf("[]float32 required but %v provided for field %s", value, f.QualifiedName)
		}
		if len(arr) == 0 {
			return bl.CreateByteVector([]byte{}), nil
		}
		length := len(arr) * flatbuffers.SizeFloat32
		hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(&arr[0])), Len: length, Cap: length}
		target := *(*[]byte)(unsafe.Pointer(&hdr))
		return bl.CreateByteVector(target), nil
	case FieldTypeDouble:
		arr, ok := intfToFloat64Arr(f, value)
		if !ok {
			return 0, fmt.Errorf("[]float32 required but %v provided for field %s", value, f.QualifiedName)
		}

		if len(arr) == 0 {
			return bl.CreateByteVector([]byte{}), nil
		}
		length := len(arr) * flatbuffers.SizeFloat64
		hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(&arr[0])), Len: length, Cap: length}
		target := *(*[]byte)(unsafe.Pointer(&hdr))
		return bl.CreateByteVector(target), nil
	case FieldTypeByte:
		arr, ok := value.([]byte)
		if !ok {
			return 0, fmt.Errorf("[]byte required but %v provided for field %s", value, f.QualifiedName)
		}
		return bl.CreateByteVector(arr), nil
	case FieldTypeString:
		var arr []string
		switch value.(type) {
		case []string:
			// Set("", []string) was called
			arr = value.([]string)
		case []interface{}:
			// came from JSON
			intfs := value.([]interface{})
			arr = make([]string, len(intfs))
			for i, intf := range intfs {
				stringVal, ok := intf.(string)
				if !ok {
					return 0, fmt.Errorf("[]byte required but %v provided for field %s", value, f.QualifiedName)
				}
				arr[i] = stringVal
			}
		case *Array:
			// copied from existing
			arrays := value.(*Array)
			arr = arrays.GetStrings()
		default:
			return 0, fmt.Errorf("%v provided for field %s which can not be converted to []string", value, f.QualifiedName)
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
		nestedUOffsetTs := []flatbuffers.UOffsetT{}
		switch value.(type) {
		case []*Buffer:
			// explicit Set("", []*Buffer) was called
			arr := value.([]*Buffer)
			for i := 0; i < len(arr); i++ {
				if arr[i] == nil {
					return 0, fmt.Errorf("nil element of array field %s is provided. Nils are not supported for array elements", f.QualifiedName)
				}
				nestedUOffsetT, err := arr[i].encodeBuffer(bl)
				if err != nil {
					return 0, err
				}
				nestedUOffsetTs = append(nestedUOffsetTs, nestedUOffsetT)
			}
		case *Array:
			// copied from existing
			arr := value.(*Array)
			for arr.Next() {
				buf := arr.Value().(*Buffer)
				nestedUOffsetT, _ := buf.encodeBuffer(bl) // should be no errors here
				nestedUOffsetTs = append(nestedUOffsetTs, nestedUOffsetT)

			}
		default:
			return 0, fmt.Errorf("%v provided for field %s is not an array of nested objects", value, f.QualifiedName)
		}

		bl.StartVector(elemSize, len(nestedUOffsetTs), elemSize)
		for i := len(nestedUOffsetTs) - 1; i >= 0; i-- {
			bl.PrependUOffsetT(nestedUOffsetTs[i])
		}
		return bl.EndVector(len(nestedUOffsetTs)), nil

	}
}

func copyFixedSizeValue(dest *flatbuffers.Builder, src *Buffer, f *Field) bool {
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

func encodeFixedSizeValue(bl *flatbuffers.Builder, f *Field, value interface{}) bool {
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

// ToJSON returns JSON key->value string
func (b *Buffer) ToJSON() string {
	buf := bytes.NewBufferString("")
	e := json.NewEncoder(buf)
	buf.WriteString("{")
	for _, f := range b.Scheme.Fields {
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
			if f.Ft == FieldTypeObject {
				if f.IsArray {
					buf.WriteString("[")
					buffers, ok := value.([]*Buffer)
					if !ok {
						arr := value.(*Array)
						buffers = arr.GetObjects()
					}
					for _, buffer := range buffers {
						buf.WriteString(buffer.ToJSON())
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

// NewScheme creates new empty Scheme
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
	s.addField(name, FieldTypeObject, nested, isMandatory, false)
}

// AddNestedArray adds array of nested objects field
// First letter is capital -> field is mandatory. First letter will be uncapitalized.
func (s *Scheme) AddNestedArray(name string, nested *Scheme, isMandatory bool) {
	s.addField(name, FieldTypeObject, nested, isMandatory, true)
}

func (s *Scheme) addField(name string, ft FieldType, nested *Scheme, isMandatory bool, IsArray bool) {
	newField := &Field{s.getQualifiedFieldName(name), name, ft, len(s.FieldsMap), isMandatory, nested, IsArray}
	s.FieldsMap[name] = newField
	s.Fields = append(s.Fields, newField)
}

// MarshalYAML marshals Scheme to yaml. Needs to conform to yaml.Marshaler interface
func (s *Scheme) MarshalYAML() (interface{}, error) {
	res := yaml.MapSlice{}
	for _, f := range s.Fields {
		for ftStr, curFt := range yamlFieldTypesMap {
			if curFt == f.Ft {
				fieldName := f.Name
				if f.IsMandatory {
					fnBytes := []byte(fieldName)
					fnBytes[0] = []byte(strings.ToUpper(fieldName))[0]
					fieldName = string(fnBytes)
				}
				if f.IsArray {
					fieldName = fieldName + ".."
				}
				var val interface{}
				if f.Ft == FieldTypeObject {
					valTemp, err := f.Scheme.MarshalYAML()
					if err != nil {
						return nil, err
					}
					val = valTemp
				} else {
					val = ftStr
				}
				item := yaml.MapItem{Key: fieldName, Value: val}
				res = append(res, item)
			}
		}
	}
	return res, nil
}

// UnmarshalYAML unmarshals Scheme from yaml. Needs to conform to yaml.Unmarshaler interface
func (s *Scheme) UnmarshalYAML(unmarshal func(interface{}) error) error {
	mapSlice := yaml.MapSlice{}
	if err := unmarshal(&mapSlice); err != nil {
		return err
	}
	newS, err := mapSliceToScheme(mapSlice)
	if err != nil {
		return err
	}
	s.Fields = newS.Fields
	s.FieldsMap = newS.FieldsMap
	return nil
}

func (s *Scheme) getQualifiedFieldName(fieldName string) string {
	if len(s.Name) > 0 {
		return s.Name + "." + fieldName
	}
	return fieldName
}

// GetNestedScheme returns Scheme of nested object if the field has FieldTypeObject type, nil otherwise
func (s *Scheme) GetNestedScheme(nestedObjectField string) *Scheme {
	if f, ok := s.FieldsMap[nestedObjectField]; ok {
		return f.Scheme
	}
	return nil
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
	res := NewScheme()
	for _, mapItem := range mapSlice {
		if nestedMapSlice, ok := mapItem.Value.(yaml.MapSlice); ok {
			fieldName, isMandatory, IsArray := fieldPropsFromYaml(mapItem.Key.(string))
			nestedScheme, err := mapSliceToScheme(nestedMapSlice)
			if err != nil {
				return nil, err
			}
			if IsArray {
				res.AddNestedArray(fieldName, nestedScheme, isMandatory)
			} else {
				res.AddNested(fieldName, nestedScheme, isMandatory)
			}
		} else if typeStr, ok := mapItem.Value.(string); ok {
			fieldName, isMandatory, IsArray := fieldPropsFromYaml(mapItem.Key.(string))
			if ft, ok := yamlFieldTypesMap[typeStr]; ok {
				if IsArray {
					res.AddArray(fieldName, ft, isMandatory)
				} else {
					res.AddField(fieldName, ft, isMandatory)
				}
			} else {
				return nil, errors.New("unknown field type: " + typeStr)
			}
		}
	}
	return res, nil
}

func fieldPropsFromYaml(name string) (fieldName string, isMandatory bool, IsArray bool) {
	isMandatory = unicode.IsUpper(rune(name[0]))

	IsArray = strings.HasSuffix(name, "..")
	if IsArray {
		name = name[:len(name)-2]
	}
	fieldName = name
	if isMandatory {
		fnBytes := []byte(fieldName)
		fnBytes[0] = []byte(strings.ToLower(string(fnBytes[0])))[0]
		fieldName = string(fnBytes)
	}
	return
}

// byteSliceToString converts a []byte to string without a heap allocation.
func byteSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func isFixedSizeField(f *Field) bool {
	return f.Ft != FieldTypeObject && f.Ft != FieldTypeString
}
