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

// storeObjectsAsBytes defines if nested objects will be stored as byte vectors.
var storeObjectsAsBytes = false

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
	Scheme *Scheme
	// if not to collect modified fields and write to out bytes on Set():
	// - Set(fld, nil): Need to remember which fields should not be read from initial bytes on ToBytes()
	// - need to remember which fields should be read from initial bytes and which were Set() on ToBytes()
	// - what to do if Set() twice for one field?
	// - impossible to write strings, arrays and nested objects because it must be written before the root object started (flatbuffers feature)
	modifiedFields []*modifiedField
	tab            flatbuffers.Table
	isModified     bool
	isReleased     bool
	owner          *Buffer
	builder        *flatbuffers.Builder
}

// func (b *Buffer) Builder() *flatbuffers.Builder {
// 	return b.builder
// }

// Field describes a Scheme field
type Field struct {
	Name        string
	Ft          FieldType
	order       int
	IsMandatory bool
	FieldScheme *Scheme // != nil for FieldTypeObject only
	ownerScheme *Scheme
	IsArray     bool
}

type modifiedField struct {
	value    interface{}
	isAppend bool
	isReleased bool
}

func (m *modifiedField) Release() {
	if buf, ok := m.value.(*Buffer); ok {
		if buf != nil {
			buf.Release()
		}
	}

	if buf, ok := m.value.(*BuffersSlice); ok {
		for _, bb := range buf.Slice {
			if bb != nil {
				bb.Release()
			}
		}

		putBufferSlice(buf)
	}

	m.value = nil
	m.isAppend = false
	m.isReleased = true
}

// ObjectArray used to iterate over array of nested objects
type ObjectArray struct {
	Buffer  *Buffer
	Len     int
	curElem int
	start   flatbuffers.UOffsetT
}

// Next proceeds to a next nested object in the array. If true then .Buffer represents the next element
func (oa *ObjectArray) Next() bool {
	oa.curElem++
	if oa.curElem >= oa.Len {
		return false
	}
	oa.Buffer.tab.Pos = oa.Buffer.tab.Indirect(oa.start + flatbuffers.UOffsetT(oa.curElem)*flatbuffers.SizeUOffsetT)
	return true
}

// Value returns *dynobuffers.Buffer instance as current element
func (oa *ObjectArray) Value() interface{} {
	return oa.Buffer
}

func (b *Buffer) getAllValues(start flatbuffers.UOffsetT, arrLen int, f *Field) interface{} {
	bytesSlice := b.tab.Bytes[start:]
	switch f.Ft {
	case FieldTypeInt:
		src := *(*[]int32)(unsafe.Pointer(&bytesSlice))
		src = src[:arrLen]
		res := make([]int32, len(src))
		copy(res, src)
		return res
	case FieldTypeFloat:
		src := *(*[]float32)(unsafe.Pointer(&bytesSlice))
		src = src[:arrLen]
		res := make([]float32, len(src))
		copy(res, src)
		return res
	case FieldTypeDouble:
		src := *(*[]float64)(unsafe.Pointer(&bytesSlice))
		src = src[:arrLen]
		res := make([]float64, len(src))
		copy(res, src)
		return res
	case FieldTypeByte:
		return b.tab.Bytes[start : arrLen+int(start)]
	case FieldTypeBool:
		src := *(*[]bool)(unsafe.Pointer(&bytesSlice))
		src = src[:arrLen]
		res := make([]bool, len(src))
		copy(res, src)
		return res
	case FieldTypeLong:
		src := *(*[]int64)(unsafe.Pointer(&bytesSlice))
		src = src[:arrLen]
		res := make([]int64, len(src))
		copy(res, src)
		return res
	default:
		// string
		res := make([]string, arrLen)
		arrayUOffsetT := b.getFieldUOffsetTByOrder(f.order)
		for i := 0; i < arrLen; i++ {
			elementUOffsetT := b.tab.Vector(arrayUOffsetT-b.tab.Pos) + flatbuffers.UOffsetT(i*flatbuffers.SizeUOffsetT)
			res[i] = byteSliceToString(b.tab.ByteVector(elementUOffsetT))
		}
		return res
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
	if Scheme == nil {
		panic("nil Scheme provided")
	}

	b := getBuffer()

	b.Scheme = Scheme
	b.isReleased = false
	b.owner = nil

	b.isModified = false

	b.Reset(nil)

	return b
}

func (b *Buffer) Release() {
	if !b.isReleased {
		b.ReleaseFields()
		b.isReleased = true
		BufferPool.Put(b)
	}
}

func (b *Buffer) ReleaseFields() {
	if b.modifiedFields != nil && len(b.modifiedFields) > 0 {
		for _, m := range b.modifiedFields {
			if m != nil {
				m.Release()
			}
		}
	}
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
	return b.getByUOffsetT(f, index, uOffsetT)
}

func (b *Buffer) getByUOffsetT(f *Field, index int, uOffsetT flatbuffers.UOffsetT) interface{} {
	if f.IsArray {
		arrayLen := b.tab.VectorLen(uOffsetT - b.tab.Pos)
		elemSize := getFBFieldSize(f.Ft)
		if isFixedSizeField(f) {
			// arrays with fixed-size elements are stored as byte arrays
			arrayLen = arrayLen / elemSize
		}
		if index < 0 {
			if f.Ft == FieldTypeObject {
				arr := &ObjectArray{Buffer: &Buffer{}}
				arr.Len = b.tab.VectorLen(uOffsetT - b.tab.Pos)
				arr.Buffer.tab.Bytes = b.tab.Bytes
				arr.start = b.tab.Vector(uOffsetT - b.tab.Pos)
				arr.curElem = -1
				arr.Buffer.Scheme = f.FieldScheme
				return arr
			}
			uOffsetT = b.tab.Vector(uOffsetT - b.tab.Pos)
			return b.getAllValues(uOffsetT, arrayLen, f)
		}
		if index > arrayLen-1 {
			return nil
		}
		uOffsetT = b.tab.Vector(uOffsetT-b.tab.Pos) + flatbuffers.UOffsetT(index*elemSize)
	}
	return b.getValueByUOffsetT(f, uOffsetT)
}

func (b *Buffer) getValueByUOffsetT(f *Field, uOffsetT flatbuffers.UOffsetT) interface{} {
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
		var res *Buffer
		if storeObjectsAsBytes {
			bytesNested := b.tab.ByteVector(uOffsetT)
			res = ReadBuffer(bytesNested, f.FieldScheme)
		} else {
			res = ReadBuffer(b.tab.Bytes, f.FieldScheme)
			res.tab.Pos = b.tab.Indirect(uOffsetT)
		}
		if !f.IsArray {
			b.prepareModifiedFields()
			if b.modifiedFields[f.order] == nil {
				b.set(f, res)
			}
		}
		res.owner = b
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
// field is scalar -> scalar is returned
// field is an array of scalars -> []T is returned
// field is a nested object -> *dynobuffers.Buffer is returned
// field is an array of nested objects -> *dynobuffers.ObjectArray is returned
// field is not set, set to nil or no such field in the Scheme -> nil
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
	if Scheme == nil {
		panic("nil Scheme provided")
	}

	b := NewBuffer(Scheme)
	//b := &Buffer{}
	//b.Scheme = Scheme
	b.Reset(bytes)
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

func (b *Buffer) setModified() {
	b.isModified = true
	if b.owner != nil {
		b.owner.setModified()
	}
}

func (b *Buffer) set(f *Field, value interface{}) {
	b.prepareModifiedFields()
	m := b.modifiedFields[f.order]

	if m == nil {
		m = &modifiedField{}
	}

	m.value = value
	m.isAppend = false
	m.isReleased = false

	b.modifiedFields[f.order] = m

	if bNested, ok := value.(*Buffer); ok {
		bNested.owner = b
	}

	b.setModified()
}

// Append s.e.
func (b *Buffer) Append(name string, toAppend interface{}) {
	f, ok := b.Scheme.FieldsMap[name]
	if !ok {
		return
	}
	b.append(f, toAppend)
}

func (b *Buffer) append(f *Field, toAppend interface{}) {

	b.prepareModifiedFields()

	m := b.modifiedFields[f.order]

	if m == nil {
		m = &modifiedField{}
	}

	m.value = toAppend
	m.isAppend = true

	b.modifiedFields[f.order] = m

	b.setModified()
}

// ApplyJSONAndToBytes sets field values described by provided json and returns new FlatBuffer byte array
// See `ApplyMap` for details
func (b *Buffer) ApplyJSONAndToBytes(jsonBytes []byte) ([]byte, error) {
	dest := map[string]interface{}{}
	err := json.Unmarshal(jsonBytes, &dest)
	if err != nil {
		return nil, err
	}
	err = b.ApplyMap(dest)
	if err != nil {
		return nil, err
	}

	return b.ToBytesWithBuilder(nil)
}

// ApplyMap sets field values described by provided map[string]interface{}
// Resulting buffer has no value (or has nil value) for a mandatory field -> error
// Value type and field type are incompatible (e.g. string for numberic field) -> error
// Value and field types differs but value fits into field -> no error. Examples:
//   255 fits into float, double, int, long, byte;
//   256 does not fit into byte
//   math.MaxInt64 does not fit into int32
// Unexisting field is provided -> error
// Previousy stored or modified data is rewritten with the provided data
// Byte arrays are expected to be base64 strings
// Array element is nil -> error (not supported)

func (b *Buffer) ApplyMap(data map[string]interface{}) error {
	for fn, fv := range data {
		f, ok := b.Scheme.FieldsMap[fn]
		if !ok {
			return fmt.Errorf("field %s does not exist in the scheme", fn)
		}
		if fv == nil {
			b.set(f, nil)
			continue
		}

		if f.Ft == FieldTypeObject {
			if f.IsArray {
				datasNested, ok := fv.([]interface{})

				if !ok {
					return fmt.Errorf("array of objects required but %#v provided for field %s", fv, f.QualifiedName())
				}

				buffers := getBufferSlice(len(datasNested))

				for i, dataNestedIntf := range datasNested {
					dataNested, ok := dataNestedIntf.(map[string]interface{})

					if !ok {
						return fmt.Errorf("element value of array field %s must be an object, %#v provided", fn, dataNestedIntf)
					}

					buffers.Slice[i] = NewBuffer(f.FieldScheme)
					buffers.Slice[i].owner = b
					buffers.Slice[i].ApplyMap(dataNested)
				}

				b.append(f, buffers)
			} else {
				bNested := NewBuffer(f.FieldScheme)
				bNested.owner = b
				dataNested, ok := fv.(map[string]interface{})
				if !ok {
					return fmt.Errorf("value of field %s must be an object, %#v provided", fn, fv)
				}
				bNested.ApplyMap(dataNested)
				b.set(f, bNested)
			}
		} else {
			if f.IsArray {
				b.append(f, fv)
			} else {
				b.set(f, fv)
			}
		}
	}
	return nil
}

func ApplyMapBuffer(jsonMap []byte) error {
	return nil
}

// ToBytes returns new FlatBuffer byte array with fields modified by Set() and fields which initially had values
// Note: initial byte array and current modifications are kept
func (b *Buffer) ToBytes() ([]byte, error) {

	var builder *flatbuffers.Builder

	if b.builder != nil {
		builder = b.builder
	} else {
		builder = BuilderPool.Get().(*flatbuffers.Builder)
}

	bytes, err := b.ToBytesWithBuilder(builder)

	if b.builder == nil {
		BuilderPool.Put(builder)
	}

	return bytes, err

}

// ToBytesWithBuilder same as ToBytes but uses builder
// builder.Reset() is invoked
func (b *Buffer) ToBytesWithBuilder(builder *flatbuffers.Builder) ([]byte, error) {
	if !b.isModified && len(b.tab.Bytes) > 0 { // mandatory fields should be checked
		return b.tab.Bytes, nil
	}

	if nil != builder {
		builder.Reset()
	} else {
		builder = flatbuffers.NewBuilder(0)
	}

	_, err := b.encodeBuffer(builder)

	if err != nil {
		return nil, err
	}

	buf := builder.FinishedBytes()

	return buf, nil
}

func (b *Buffer) prepareModifiedFields() {
	if len(b.Scheme.Fields) > cap(b.modifiedFields) {
		b.modifiedFields = make([]*modifiedField, len(b.Scheme.Fields))
	} else {
		b.modifiedFields = b.modifiedFields[:len(b.Scheme.Fields)]
	}
}

type offset struct {
	str flatbuffers.UOffsetT
	obj flatbuffers.UOffsetT
	arr flatbuffers.UOffsetT
}

func (b *Buffer) encodeBuffer(bl *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	offsets := getOffsetSlice(len(b.Scheme.Fields))

	b.prepareModifiedFields()

	var err error

	for _, f := range b.Scheme.Fields {
		if f.IsArray {
			arrayUOffsetT := flatbuffers.UOffsetT(0)
			modifiedField := b.modifiedFields[f.order]
			if modifiedField != nil {
				if modifiedField.value != nil {
					val := reflect.ValueOf(modifiedField.value)
					if isSlice(val.Kind()) && val.IsNil() {
						continue
					}
					var toAppendToIntf interface{} = nil
					if modifiedField.isAppend {
						toAppendToIntf = b.getByField(f, -1)
					}
					if arrayUOffsetT, err = b.encodeArray(bl, f, modifiedField.value, toAppendToIntf); err != nil {
						return 0, err
					}
				}
			} else {
				if uOffsetT := b.getFieldUOffsetTByOrder(f.order); uOffsetT != 0 {
					// copy from source bytes if not modified and initially existed
					if isFixedSizeField(f) {
						// copy fixed-size array as byte array
						arrayLen := b.tab.VectorLen(uOffsetT - b.tab.Pos)
						uOffsetT = b.tab.Vector(uOffsetT - b.tab.Pos)
						arrayUOffsetT = bl.CreateByteVector(b.tab.Bytes[uOffsetT : int(uOffsetT)+arrayLen])
					} else {
						// re-encode var-size array
						if existingArray := b.getByUOffsetT(f, -1, uOffsetT); existingArray != nil {
							arrayUOffsetT, _ = b.encodeArray(bl, f, existingArray, nil) // no errors should be here
						}
					}
				}
			}
			offsets.Slice[f.order].arr = arrayUOffsetT
		} else if f.Ft == FieldTypeObject {
			nestedUOffsetT := flatbuffers.UOffsetT(0)
			modifiedField := b.modifiedFields[f.order]
			if modifiedField != nil {
				if modifiedField.value != nil {
					if nestedBuffer, ok := modifiedField.value.(*Buffer); !ok {
						return 0, fmt.Errorf("nested object required but %#v provided for field %s", modifiedField.value, f.QualifiedName())
					} else if storeObjectsAsBytes {
						nestedBytes, err := nestedBuffer.ToBytes()
						if err != nil {
							return 0, fmt.Errorf("failed to encode nested object %s: %s", f.QualifiedName(), err)
						}
						nestedUOffsetT = bl.CreateByteVector(nestedBytes)
					} else if nestedUOffsetT, err = nestedBuffer.encodeBuffer(bl); err != nil {
						return 0, err
					}
				}
			} else {
				if uOffsetT := b.getFieldUOffsetTByOrder(f.order); uOffsetT != 0 {
					bufToWrite := b.getByUOffsetT(f, -1, uOffsetT) // can not be nil
					if storeObjectsAsBytes {
						nestedBytes, _ := bufToWrite.(*Buffer).ToBytes() // no errors should be here
						nestedUOffsetT = bl.CreateByteVector(nestedBytes)
					} else {
						nestedUOffsetT, _ = bufToWrite.(*Buffer).encodeBuffer(bl) // no errors should be here
					}
				}
			}
			offsets.Slice[f.order].obj = nestedUOffsetT
		} else if f.Ft == FieldTypeString {
			modifiedStringField := b.modifiedFields[f.order]
			if modifiedStringField != nil {
				if modifiedStringField.value != nil {
					if strToWrite, ok := modifiedStringField.value.(string); ok {
						offsets.Slice[f.order].str = bl.CreateString(strToWrite)
					} else {
						return 0, fmt.Errorf("string required but %#v provided for field %s", modifiedStringField.value, f.QualifiedName())
					}
				}
			} else {
				if strToWrite, ok := b.getByStringField(f); ok {
					offsets.Slice[f.order].str = bl.CreateString(strToWrite)
				}
			}
		}
	}

	bl.StartObject(len(b.Scheme.Fields))
	for _, f := range b.Scheme.Fields {
		isSet := false
		if f.IsArray {
			if isSet = offsets.Slice[f.order].arr > 0; isSet {
				bl.PrependUOffsetTSlot(f.order, offsets.Slice[f.order].arr, 0)
			}
		} else {
			switch f.Ft {
			case FieldTypeString:
				if isSet = offsets.Slice[f.order].str > 0; isSet {
					bl.PrependUOffsetTSlot(f.order, offsets.Slice[f.order].str, 0)
				}
			case FieldTypeObject:
				if isSet = offsets.Slice[f.order].obj > 0; isSet {
					bl.PrependUOffsetTSlot(f.order, offsets.Slice[f.order].obj, 0)
				}
			default:
				modifiedField := b.modifiedFields[f.order]
				if modifiedField != nil {
					if isSet = modifiedField.value != nil; isSet {
						if !encodeFixedSizeValue(bl, f, modifiedField.value) {
							return 0, fmt.Errorf("wrong value %T(%#v) provided for field %s", modifiedField.value, modifiedField.value, f.QualifiedName())
						}
					}
				} else {
					isSet = copyFixedSizeValue(bl, b, f)
				}
			}
		}
		if f.IsMandatory && !isSet {
			return 0, fmt.Errorf("Mandatory field %s is not set", f.QualifiedName())
		}
	}
	res := bl.EndObject()
	bl.Finish(res)

	putOffsetSlice(offsets)

	return res, nil
}

// HasValue returns if specified field exists in the scheme and its value is set to non-nil
func (b *Buffer) HasValue(name string) bool {
	return b.getFieldUOffsetT(name) != 0
}

// Reset sets current underlying byte array and clears modified fields. Useful for *Buffer instance reuse
// Note: bytes must match the Buffer's scheme
func (b *Buffer) Reset(bytes []byte) {
	b.tab.Bytes = append(b.tab.Bytes[:0], bytes...)
	if len(bytes) == 0 {
		b.tab.Pos = 0
	} else {
		b.tab.Pos = flatbuffers.GetUOffsetT(bytes)
	}

	if b.modifiedFields != nil {
		b.ReleaseFields()

		b.modifiedFields = b.modifiedFields[:0]
	}

	b.isModified = false
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

func (b *Buffer) encodeArray(bl *flatbuffers.Builder, f *Field, value interface{}, toAppendToIntf interface{}) (flatbuffers.UOffsetT, error) {
	elemSize := getFBFieldSize(f.Ft)
	switch f.Ft {
	case FieldTypeInt:
		arr, ok := intfToInt32Arr(f, value)
		if !ok {
			return 0, fmt.Errorf("[]int32 required but %#v provided for field %s", value, f.QualifiedName())
		}
		if toAppendToIntf != nil {
			toAppendTo := toAppendToIntf.([]int32)
			toAppendTo = append(toAppendTo, arr...)
			arr = toAppendTo
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
			return 0, fmt.Errorf("[]bool required but %#v provided for field %s", value, f.QualifiedName())
		}
		if toAppendToIntf != nil {
			toAppendTo := toAppendToIntf.([]bool)
			toAppendTo = append(toAppendTo, arr...)
			arr = toAppendTo
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
			return 0, fmt.Errorf("[]int64 required but %#v provided for field %s", value, f.QualifiedName())
		}
		if toAppendToIntf != nil {
			toAppendTo := toAppendToIntf.([]int64)
			toAppendTo = append(toAppendTo, arr...)
			arr = toAppendTo
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
			return 0, fmt.Errorf("[]float32 required but %#v provided for field %s", value, f.QualifiedName())
		}
		if toAppendToIntf != nil {
			toAppendTo := toAppendToIntf.([]float32)
			toAppendTo = append(toAppendTo, arr...)
			arr = toAppendTo
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
			return 0, fmt.Errorf("[]float32 required but %#v provided for field %s", value, f.QualifiedName())
		}
		if toAppendToIntf != nil {
			toAppendTo := toAppendToIntf.([]float64)
			toAppendTo = append(toAppendTo, arr...)
			arr = toAppendTo
		}
		if len(arr) == 0 {
			return bl.CreateByteVector([]byte{}), nil
		}
		length := len(arr) * flatbuffers.SizeFloat64
		hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(&arr[0])), Len: length, Cap: length}
		target := *(*[]byte)(unsafe.Pointer(&hdr))
		return bl.CreateByteVector(target), nil
	case FieldTypeByte:
		target := []byte{}
		switch arr := value.(type) {
		case []byte:
			target = arr
		case string:
			var err error
			if target, err = base64.StdEncoding.DecodeString(arr); err != nil {
				return 0, fmt.Errorf("the string %s considered as base64-encoded value for field %s: %s", arr, f.QualifiedName(), err)
			}
		default:
			return 0, fmt.Errorf("[]byte or base64-encoded string required but %#v provided for field %s", value, f.QualifiedName())

		}
		if toAppendToIntf != nil {
			toAppendTo := toAppendToIntf.([]byte)
			toAppendTo = append(toAppendTo, target...)
			target = toAppendTo
		}
		return bl.CreateByteVector(target), nil
	case FieldTypeString:
		var target *StringsSlice
		switch arr := value.(type) {
		case []string:
			// Set("", []string) was called
			target = &StringsSlice{Slice: arr}
		case []interface{}:
			// came from JSON
			target = getStringSlice(len(arr)) //make([]string, )
			for i, intf := range arr {
				stringVal, ok := intf.(string)
				if !ok {
					return 0, fmt.Errorf("[]byte required but %#v provided for field %s", value, f.QualifiedName())
				}
				target.Slice[i] = stringVal
			}
		default:
			return 0, fmt.Errorf("%#v provided for field %s which can not be converted to []string", value, f.QualifiedName())
		}
		if toAppendToIntf != nil {
			toAppendTo := toAppendToIntf.([]string)
			toAppendTo = append(toAppendTo, target.Slice...)
			target.Slice = toAppendTo
		}

		stringUOffsetTs := getUOffsetSlice(len(target.Slice))

		for i := 0; i < len(target.Slice); i++ {
			stringUOffsetTs.Slice[i] = bl.CreateString(target.Slice[i])
		}
		bl.StartVector(elemSize, len(target.Slice), elemSize)

		for i := len(target.Slice) - 1; i >= 0; i-- {
			bl.PrependUOffsetT(stringUOffsetTs.Slice[i])
		}

		putUOffsetSlice(stringUOffsetTs)
		of := bl.EndVector(len(target.Slice))

		putStringSlice(target)

		return of, nil
	default:
		nestedUOffsetTs := getUOffsetSlice(0)
		switch arr := value.(type) {
		case *BuffersSlice:
			// explicit Set\Append("", []*Buffer) was called
			for i := 0; i < len(arr.Slice); i++ {
				if arr.Slice[i] == nil {
					return 0, fmt.Errorf("nil element of array field %s is provided. Nils are not supported for array elements", f.QualifiedName())
				}
				if storeObjectsAsBytes {
					nestedBytes, err := arr.Slice[i].ToBytes()
					if err != nil {
						return 0, err
					}
					nestedUOffsetTs.Slice = append(nestedUOffsetTs.Slice, bl.CreateByteVector(nestedBytes))
				} else {
					nestedUOffsetT, err := arr.Slice[i].encodeBuffer(bl)
					if err != nil {
						return 0, err
					}
					nestedUOffsetTs.Slice = append(nestedUOffsetTs.Slice, nestedUOffsetT)
				}
			}
		case []*Buffer:
			// explicit Set\Append("", []*Buffer) was called
			for i := 0; i < len(arr); i++ {
				if arr[i] == nil {
					return 0, fmt.Errorf("nil element of array field %s is provided. Nils are not supported for array elements", f.QualifiedName())
				}
				if storeObjectsAsBytes {
					nestedBytes, err := arr[i].ToBytes()
					if err != nil {
						return 0, err
					}
					nestedUOffsetTs.Slice = append(nestedUOffsetTs.Slice, bl.CreateByteVector(nestedBytes))
				} else {
					nestedUOffsetT, err := arr[i].encodeBuffer(bl)
					if err != nil {
						return 0, err
					}
					nestedUOffsetTs.Slice = append(nestedUOffsetTs.Slice, nestedUOffsetT)
				}
			}
		case *ObjectArray:
			for arr.Next() {
				if storeObjectsAsBytes {
					nestedBytes, _ := arr.Buffer.ToBytes()
					nestedUOffsetTs.Slice = append(nestedUOffsetTs.Slice, bl.CreateByteVector(nestedBytes))
				} else {
					nestedUOffsetT, _ := arr.Buffer.encodeBuffer(bl) // should be no errors here
					nestedUOffsetTs.Slice = append(nestedUOffsetTs.Slice, nestedUOffsetT)
				}
			}

		default:
			return 0, fmt.Errorf("%#v provided for field %s is not an array of nested objects", value, f.QualifiedName())
		}

		if toAppendToIntf != nil {
			toAppendToArr := toAppendToIntf.(*ObjectArray)

			toAppendToUOffsetTs := getUOffsetSlice(toAppendToArr.Len)

			for i := 0; toAppendToArr.Next(); i++ {
				if storeObjectsAsBytes {
					bufBytes, _ := toAppendToArr.Buffer.ToBytes()
					toAppendToUOffsetTs.Slice[i] = bl.CreateByteVector(bufBytes)
				} else {
					toAppendToUOffsetTs.Slice[i], _ = toAppendToArr.Buffer.encodeBuffer(bl)
				}
			}

			toAppendToUOffsetTs.Slice = append(toAppendToUOffsetTs.Slice, nestedUOffsetTs.Slice...)

			putUOffsetSlice(nestedUOffsetTs)

			nestedUOffsetTs = toAppendToUOffsetTs
		}

		bl.StartVector(elemSize, len(nestedUOffsetTs.Slice), elemSize)
		for i := len(nestedUOffsetTs.Slice) - 1; i >= 0; i-- {
			bl.PrependUOffsetT(nestedUOffsetTs.Slice[i])
		}

		o := bl.EndVector(len(nestedUOffsetTs.Slice))
		putUOffsetSlice(nestedUOffsetTs)

		return o, nil
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
		res := f.Ft == FieldTypeInt || f.Ft == FieldTypeLong || f.Ft == FieldTypeDouble || f.Ft == FieldTypeFloat
		if float64Src >= 0 && float64Src <= 255 {
			return res || f.Ft == FieldTypeByte
		}
		return res
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
func (b *Buffer) ToJSON() []byte {
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
			buf.WriteString("\"" + f.Name + "\":")
			if f.Ft == FieldTypeObject {
				if f.IsArray {
					buf.WriteString("[")
					if arr, ok := value.(*ObjectArray); ok {
						for arr.Next() {
							buf.Write(arr.Buffer.ToJSON())
							buf.WriteString(",")
						}
					} else if bs, ok := value.(*BuffersSlice); ok {
						for _, buffer := range bs.Slice {
							if buffer != nil {
								buf.Write(buffer.ToJSON())
								buf.WriteString(",")
							}
						}
					} else {
						buffers, _ := value.([]*Buffer)
						for _, buffer := range buffers {
							if buffer != nil {
							buf.Write(buffer.ToJSON())
							buf.WriteString(",")
						}

					}
					}
					buf.Truncate(buf.Len() - 1)
					buf.WriteString("]")
				} else {
					buf.Write(value.(*Buffer).ToJSON())
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
	return []byte(strings.Replace(buf.String(), "\n", "", -1))
}

// GetBytes returns underlying byte buffer
func (b *Buffer) GetBytes() []byte {
	return b.tab.Bytes
}

// GetNames returns list of field names which values are non-nil in current buffer
// Set() fields are not considered
// fields of nested objects are not considered
func (b *Buffer) GetNames() []string {
	if len(b.tab.Bytes) == 0 {
		return nil
	}

	res := make([]string, 0, len(b.Scheme.Fields))
	vTable := flatbuffers.UOffsetT(flatbuffers.SOffsetT(b.tab.Pos) - b.tab.GetSOffsetT(b.tab.Pos))
	vOffsetT := b.tab.GetVOffsetT(vTable)

	for order := 0; true; order++ {
		vTableOffset := flatbuffers.VOffsetT((order + 2) * 2)
		if vTableOffset >= vOffsetT {
			break
		}
		if b.tab.GetVOffsetT(vTable+flatbuffers.UOffsetT(vTableOffset)) > 0 {
			res = append(res, b.Scheme.Fields[order].Name)
		}
	}
	return res
}

// ToJSONMap returns map[string]interface{} representation of the buffer compatible to json
// numeric field types are kept (not float64 as json does)
func (b *Buffer) ToJSONMap() map[string]interface{} {
	res := map[string]interface{}{}
	for _, f := range b.Scheme.Fields {
		var storedVal interface{}
		if len(b.modifiedFields) == 0 {
			storedVal = b.getByField(f, -1)
		} else {
			modifiedField := b.modifiedFields[f.order]
			if modifiedField != nil {
				storedVal = modifiedField.value
			} else {
				storedVal = b.getByField(f, -1)
			}
		}
		if storedVal != nil {
			if f.Ft == FieldTypeObject {
				if f.IsArray {
					targetArr := []interface{}{}
					if arr, ok := storedVal.(*ObjectArray); ok {
						for arr.Next() {
							targetArr = append(targetArr, arr.Buffer.ToJSONMap())
						}
					} else {
						buffers, _ := storedVal.(*BuffersSlice)
						for _, buffer := range buffers.Slice {
							targetArr = append(targetArr, buffer.ToJSONMap())
						}
					}
					res[f.Name] = targetArr
				} else {
					res[f.Name] = storedVal.(*Buffer).ToJSONMap()
				}
			} else {
				res[f.Name] = storedVal
			}
		}
	}
	return res
}

// NewScheme creates new empty Scheme
func NewScheme() *Scheme {
	return &Scheme{"", map[string]*Field{}, []*Field{}}
}

// AddField adds field
func (s *Scheme) AddField(name string, ft FieldType, isMandatory bool) {
	s.AddFieldC(name, ft, nil, isMandatory, false)
}

// AddArray adds array field
func (s *Scheme) AddArray(name string, elementType FieldType, isMandatory bool) {
	s.AddFieldC(name, elementType, nil, isMandatory, true)
}

// AddNested adds nested object field
func (s *Scheme) AddNested(name string, nested *Scheme, isMandatory bool) {
	s.AddFieldC(name, FieldTypeObject, nested, isMandatory, false)
}

// AddNestedArray adds array of nested objects field
func (s *Scheme) AddNestedArray(name string, nested *Scheme, isMandatory bool) {
	s.AddFieldC(name, FieldTypeObject, nested, isMandatory, true)
}

// AddFieldC adds new finely-tuned field
func (s *Scheme) AddFieldC(name string, ft FieldType, nested *Scheme, isMandatory bool, IsArray bool) {
	newField := &Field{name, ft, len(s.FieldsMap), isMandatory, nested, s, IsArray}
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
					valTemp, err := f.FieldScheme.MarshalYAML()
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
	newS, err := MapSliceToScheme(mapSlice)
	if err != nil {
		return err
	}
	s.Fields = newS.Fields
	s.FieldsMap = newS.FieldsMap
	return nil
}

// GetNestedScheme returns Scheme of nested object if the field has FieldTypeObject type, nil otherwise
func (s *Scheme) GetNestedScheme(nestedObjectField string) *Scheme {
	if f, ok := s.FieldsMap[nestedObjectField]; ok {
		return f.FieldScheme
	}
	return nil
}

// QualifiedName returns ownerScheme.fieldName
func (f *Field) QualifiedName() string {
	if len(f.ownerScheme.Name) > 0 {
		return f.ownerScheme.Name + "." + f.Name
	}
	return f.Name
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
	return MapSliceToScheme(mapSlice)
}

// MapSliceToScheme s.e.
func MapSliceToScheme(mapSlice yaml.MapSlice) (*Scheme, error) {
	res := NewScheme()
	for _, mapItem := range mapSlice {
		if nestedMapSlice, ok := mapItem.Value.(yaml.MapSlice); ok {
			fieldName, isMandatory, IsArray := fieldPropsFromYaml(mapItem.Key.(string))
			nestedScheme, err := MapSliceToScheme(nestedMapSlice)
			nestedScheme.Name = fieldName
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

func fieldPropsFromYaml(yamlStr string) (fieldName string, isMandatory bool, isArray bool) {
	isMandatory = unicode.IsUpper(rune(yamlStr[0]))

	isArray = strings.HasSuffix(yamlStr, "..")
	if isArray {
		yamlStr = yamlStr[:len(yamlStr)-2]
	}
	fieldName = yamlStr
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

func isSlice(kind reflect.Kind) bool {
	return kind == reflect.Array || kind == reflect.Slice
}
