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
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode"
	"unsafe"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/untillpro/gojay"
	"gopkg.in/yaml.v2"
)

// FieldType s.e.
type FieldType int

const (
	// FieldTypeUnspecified - wrong type
	FieldTypeUnspecified FieldType = iota
	// FieldTypeObject field is nested Scheme
	FieldTypeObject
	// FieldTypeInt32 int32
	FieldTypeInt32
	// FieldTypeInt64 int64
	FieldTypeInt64
	// FieldTypeFloat32 float32
	FieldTypeFloat32
	// FieldTypeFloat64 float64
	FieldTypeFloat64
	// FieldTypeString variable length
	FieldTypeString
	// FieldTypeBool bool
	FieldTypeBool
	// FieldTypeByte byte
	FieldTypeByte
)

// storeObjectsAsBytes defines if nested objects will be stored as byte vectors.
// true:  BenchmarkSimpleDynobuffersArrayOfObjectsSet-4   	 1000000	      1020 ns/op	     272 B/op	      14 allocs/op
// false: BenchmarkSimpleDynobuffersArrayOfObjectsSet-4   	 1652797	       732 ns/op	     184 B/op	       8 allocs/op
var storeObjectsAsBytes = false

var yamlFieldTypesMap = map[string]FieldType{
	"int32":   FieldTypeInt32,
	"int64":   FieldTypeInt64,
	"float32": FieldTypeFloat32,
	"float64": FieldTypeFloat64,
	"string":  FieldTypeString,
	"bool":    FieldTypeBool,
	"byte":    FieldTypeByte,
	"":        FieldTypeObject,
}

var fieldTypesNamesMap = map[FieldType]string{}

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
	toRelease      []interface{}
}

// Field describes a Scheme field
type Field struct {
	Name        string
	Ft          FieldType
	Order       int
	IsMandatory bool
	FieldScheme *Scheme // != nil for FieldTypeObject only
	ownerScheme *Scheme
	IsArray     bool
}

type modifiedField struct {
	value      interface{}
	isAppend   bool
	isReleased bool // Allows to reuse the `modifiedField` instance instead of `modifiedFields[i] = nil` on release modified fields
}

func (m *modifiedField) Release() {
	if m.isReleased {
		return
	}
	switch typed := m.value.(type) {
	case interface{ Release() }:
		typed.Release()
	case []*Buffer:
		for _, b := range typed {
			if b != nil {
				b.Release()
			}
		}
	}
	m.value = nil
	m.isAppend = false
	m.isReleased = true
}

// ObjectArray used to iterate over array of nested objects
// ObjectArray.Buffer should be used for reading only
type ObjectArray struct {
	Buffer     *Buffer
	Len        int
	curElem    int
	start      flatbuffers.UOffsetT
	isReleased bool
}

// Next proceeds to a next nested object in the array. If true then .Buffer represents the next element
func (oa *ObjectArray) Next() bool {
	oa.curElem++
	if oa.curElem >= oa.Len {
		return false
	}
	oa.Buffer.tab.Pos = oa.Buffer.tab.Indirect(oa.start + flatbuffers.UOffsetT(oa.Len-1-oa.curElem)*flatbuffers.SizeUOffsetT)
	return true
}

// Value returns *dynobuffers.Buffer instance as current element
func (oa *ObjectArray) Value() interface{} {
	return oa.Buffer
}

// Release returns used ObjectArray instance to the pool. Releases also ObjectArray.Buffer
// Note: ObjectArray instance itself, ObjectArray.Buffer, result of ObjectArray.Buffer.ToBytes() must  not be used after Release()
func (oa *ObjectArray) Release() {
	if !oa.isReleased {
		oa.Buffer.Release()
		oa.Buffer = nil
		oa.isReleased = true
		putObjectArray(oa)
	}
}

func (b *Buffer) getAllValues(start flatbuffers.UOffsetT, f *Field) interface{} {
	switch f.Ft {
	case FieldTypeInt32:
		intf := getImplIInt32Array(b, start)
		res := make([]int32, intf.Len())
		for i := 0; i < intf.Len(); i++ {
			res[i] = intf.At(i)
		}
		return res
	case FieldTypeFloat32:
		intf := getImplIFloat32Array(b, start)
		res := make([]float32, intf.Len())
		for i := 0; i < intf.Len(); i++ {
			res[i] = intf.At(i)
		}
		return res
	case FieldTypeFloat64:
		intf := getImplIFloat64Array(b, start)
		res := make([]float64, intf.Len())
		for i := 0; i < intf.Len(); i++ {
			res[i] = intf.At(i)
		}
		return res
	case FieldTypeByte:
		intf := getImplIByteArray(b, start)
		return intf.Bytes()
	case FieldTypeBool:
		intf := getImplIBoolArray(b, start)
		res := make([]bool, intf.Len())
		for i := 0; i < intf.Len(); i++ {
			res[i] = intf.At(i)
		}
		return res
	case FieldTypeInt64:
		intf := getImplIInt64Array(b, start)
		res := make([]int64, intf.Len())
		for i := 0; i < intf.Len(); i++ {
			res[i] = intf.At(i)
		}
		return res
	default: // string
		intf := getImplIStringArray(b, start)
		res := make([]string, intf.Len())
		for i := 0; i < intf.Len(); i++ {
			res[i] = intf.At(i)
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
	b.isModified = false
	b.toRelease = b.toRelease[:0]
	b.owner = nil
	b.Reset(nil)

	return b
}

// Release returns used Buffer into pool
// Note: Buffer instance itself, result of ToBytes() must not be used after Release()
func (b *Buffer) Release() {
	if b.isReleased {
		return
	}
	b.releaseFields()
	for _, releaseableIntf := range b.toRelease {
		if releaseable, ok := releaseableIntf.(interface{ Release() }); ok {
			releaseable.Release()
		}
	}
	b.isReleased = true
	putBuffer(b)
}

func (b *Buffer) releaseFields() {
	for _, m := range b.modifiedFields {
		if m != nil {
			m.Release()
		}
	}
}

// GetInt32 returns int32 value by name and if the Scheme contains the field and the value was set to non-nil
func (b *Buffer) GetInt32(name string) (int32, bool) {
	if o := b.getFieldUOffsetT(name); o != 0 {
		return b.tab.GetInt32(o), true
	}
	return 0, false
}

// GetFloat32 returns float32 value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetFloat32(name string) (float32, bool) {
	if o := b.getFieldUOffsetT(name); o != 0 {
		return b.tab.GetFloat32(o), true
	}
	return 0, false
}

// GetString returns string value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetString(name string) (string, bool) {
	if o := b.getFieldUOffsetT(name); o != 0 {
		return byteSliceToString(b.tab.ByteVector(o)), true
	}
	return "", false
}

// GetInt64 returns int64 value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetInt64(name string) (int64, bool) {
	if o := b.getFieldUOffsetT(name); o != 0 {
		return b.tab.GetInt64(o), true
	}
	return 0, false
}

// GetFloat64 returns float64 value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetFloat64(name string) (float64, bool) {
	if o := b.getFieldUOffsetT(name); o != 0 {
		return b.tab.GetFloat64(o), true
	}
	return 0, false
}

// GetByte returns byte value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetByte(name string) (byte, bool) {
	if o := b.getFieldUOffsetT(name); o != 0 {
		return b.tab.GetByte(o), true
	}
	return 0, false
}

// GetBool returns bool value by name and if the Scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetBool(name string) (bool, bool) {
	if o := b.getFieldUOffsetT(name); o != 0 {
		return b.tab.GetBool(o), true
	}
	return false, false
}

func (b *Buffer) getFieldUOffsetT(name string) flatbuffers.UOffsetT {
	if len(b.tab.Bytes) > 0 {
		if f, ok := b.Scheme.FieldsMap[name]; ok {
			return b.getFieldUOffsetTByOrder(f.Order)
		}
	}
	return 0
}

func (b *Buffer) getFieldUOffsetTByOrder(order int) flatbuffers.UOffsetT {
	if len(b.tab.Bytes) > 0 {
		if preOffset := flatbuffers.UOffsetT(b.tab.Offset(flatbuffers.VOffsetT((order + 2) * 2))); preOffset > 0 {
			return preOffset + b.tab.Pos
		}
	}
	return 0
}

func (b *Buffer) getByField(f *Field) interface{} {
	if uOffsetT := b.getFieldUOffsetTByOrder(f.Order); uOffsetT != 0 {
		return b.getByUOffsetT(f, uOffsetT)
	}
	return nil
}

func (b *Buffer) getByUOffsetT(f *Field, uOffsetT flatbuffers.UOffsetT) interface{} {
	if f.IsArray {
		if f.Ft == FieldTypeObject {
			arr := getObjectArray()
			arr.Buffer = NewBuffer(f.FieldScheme)
			arr.Len = b.tab.VectorLen(uOffsetT - b.tab.Pos)
			arr.curElem = -1
			arr.start = b.tab.Vector(uOffsetT - b.tab.Pos)
			arr.Buffer.tab.Bytes = b.tab.Bytes
			arr.Buffer.isModified = true // to force build correct bytes array on arr.Buffer.ToBytes(). Otherwise the entire b.tab.Bytes will be returned (if unmodified) instead of arr.Buffer
			b.toRelease = append(b.toRelease, arr)
			return arr
		}
		return b.getAllValues(uOffsetT, f)
	}
	switch f.Ft {
	case FieldTypeInt32:
		return b.tab.GetInt32(uOffsetT)
	case FieldTypeInt64:
		return b.tab.GetInt64(uOffsetT)
	case FieldTypeFloat32:
		return b.tab.GetFloat32(uOffsetT)
	case FieldTypeFloat64:
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
		setted := false
		if !f.IsArray {
			b.prepareModifiedFields()
			if b.modifiedFields[f.Order] == nil || b.modifiedFields[f.Order].isReleased {
				setted = true
				b.set(f, res)
			}
		}
		res.setModified() // to force new correct bytes generation on GetBytes(). Otherwise the entire b.tab.Bytes will be returned - it is not res, it _contains_ res
		res.owner = b
		if !setted {
			// in modified fields _. will be released by modifiedFields.Release(). Otherwise will be released on b.Release()
			b.toRelease = append(b.toRelease, res)
		}

		return res
	default:
		return byteSliceToString(b.tab.ByteVector(uOffsetT))
	}
}

// GetByField is an analogue of Get() but accepts a known Field
func (b *Buffer) GetByField(f *Field) interface{} {
	return b.getByField(f)
}

// Get returns stored field value by name.
// field is scalar -> scalar is returned
// field is an array of scalars -> []T is returned
// field is a nested object -> *dynobuffers.Buffer is returned.
// field is an array of nested objects -> *dynobuffers.ObjectArray is returned.
// field is not set, set to nil or no such field in the Scheme -> nil
// `Get()` will not consider modifications made by Set, Append, ApplyJSONAndToBytes, ApplyMapBuffer, ApplyMap
func (b *Buffer) Get(name string) interface{} {
	f, ok := b.Scheme.FieldsMap[name]
	if !ok {
		return nil
	}
	return b.getByField(f)
}

func (b *Buffer) getArrIntf(f *Field) interface{} {
	uOffsetT := b.getFieldUOffsetTByOrder(f.Order)
	if uOffsetT == 0 {
		return nil
	}
	switch f.Ft {
	case FieldTypeInt32:
		return getImplIInt32Array(b, uOffsetT)
	case FieldTypeInt64:
		return getImplIInt64Array(b, uOffsetT)
	case FieldTypeFloat32:
		return getImplIFloat32Array(b, uOffsetT)
	case FieldTypeFloat64:
		return getImplIFloat64Array(b, uOffsetT)
	case FieldTypeByte:
		return getImplIByteArray(b, uOffsetT)
	case FieldTypeString:
		return getImplIStringArray(b, uOffsetT)
	case FieldTypeBool:
		return getImplIBoolArray(b, uOffsetT)
	default:
		return b.getByUOffsetT(f, uOffsetT)
	}
}

func (b *Buffer) GetInt32Array(name string) IInt32Array {
	uOffsetT := b.getFieldUOffsetT(name)
	if uOffsetT == 0 {
		return nil
	}
	return getImplIInt32Array(b, uOffsetT)
}

func (b *Buffer) GetInt64Array(name string) IInt64Array {
	uOffsetT := b.getFieldUOffsetT(name)
	if uOffsetT == 0 {
		return nil
	}
	return getImplIInt64Array(b, uOffsetT)
}

func (b *Buffer) GetFloat32Array(name string) IFloat32Array {
	uOffsetT := b.getFieldUOffsetT(name)
	if uOffsetT == 0 {
		return nil
	}
	return getImplIFloat32Array(b, uOffsetT)
}

func (b *Buffer) GetFloat64Array(name string) IFloat64Array {
	uOffsetT := b.getFieldUOffsetT(name)
	if uOffsetT == 0 {
		return nil
	}
	return getImplIFloat64Array(b, uOffsetT)
}

func (b *Buffer) GetStringArray(name string) IStringArray {
	uOffsetT := b.getFieldUOffsetT(name)
	if uOffsetT == 0 {
		return nil
	}
	return getImplIStringArray(b, uOffsetT)
}

func (b *Buffer) GetByteArray(name string) IByteArray {
	uOffsetT := b.getFieldUOffsetT(name)
	if uOffsetT == 0 {
		return nil
	}
	return getImplIByteArray(b, uOffsetT)
}

func (b *Buffer) GetBoolArray(name string) IBoolArray {
	uOffsetT := b.getFieldUOffsetT(name)
	if uOffsetT == 0 {
		return nil
	}
	return getImplIBoolArray(b, uOffsetT)
}

func getImplIInt32Array(b *Buffer, uOffsetT flatbuffers.UOffsetT) IInt32Array {
	return implIInt32Array{
		abstractArray: abstractArray{
			len:     b.tab.VectorLen(uOffsetT - b.tab.Pos),
			uOffset: b.tab.Vector(uOffsetT - b.tab.Pos),
			tab:     b.tab,
		},
	}
}

func getImplIInt64Array(b *Buffer, uOffsetT flatbuffers.UOffsetT) IInt64Array {
	return implIInt64Array{
		abstractArray: abstractArray{
			len:     b.tab.VectorLen(uOffsetT - b.tab.Pos),
			uOffset: b.tab.Vector(uOffsetT - b.tab.Pos),
			tab:     b.tab,
		},
	}
}

func getImplIFloat32Array(b *Buffer, uOffsetT flatbuffers.UOffsetT) IFloat32Array {
	return implIFloat32Array{
		abstractArray: abstractArray{
			len:     b.tab.VectorLen(uOffsetT - b.tab.Pos),
			uOffset: b.tab.Vector(uOffsetT - b.tab.Pos),
			tab:     b.tab,
		},
	}
}

func getImplIFloat64Array(b *Buffer, uOffsetT flatbuffers.UOffsetT) IFloat64Array {
	return implIFloat64Array{
		abstractArray: abstractArray{
			len:     b.tab.VectorLen(uOffsetT - b.tab.Pos),
			uOffset: b.tab.Vector(uOffsetT - b.tab.Pos),
			tab:     b.tab,
		},
	}
}

func getImplIStringArray(b *Buffer, uOffsetT flatbuffers.UOffsetT) IStringArray {
	return implIStringArray{
		abstractArray: abstractArray{
			len:     b.tab.VectorLen(uOffsetT - b.tab.Pos),
			uOffset: b.tab.Vector(uOffsetT - b.tab.Pos),
			tab:     b.tab,
		},
	}
}

func getImplIByteArray(b *Buffer, uOffsetT flatbuffers.UOffsetT) IByteArray {
	return implIByteArray{
		abstractArray: abstractArray{
			len:     b.tab.VectorLen(uOffsetT - b.tab.Pos),
			uOffset: b.tab.Vector(uOffsetT - b.tab.Pos),
			tab:     b.tab,
		},
	}
}

func getImplIBoolArray(b *Buffer, uOffsetT flatbuffers.UOffsetT) IBoolArray {
	return implIBoolArray{
		abstractArray: abstractArray{
			len:     b.tab.VectorLen(uOffsetT - b.tab.Pos),
			uOffset: b.tab.Vector(uOffsetT - b.tab.Pos),
			tab:     b.tab,
		},
	}
}

// ReadBuffer creates Buffer from bytes using provided Scheme
func ReadBuffer(bytes []byte, Scheme *Scheme) *Buffer {
	b := NewBuffer(Scheme)
	b.Reset(bytes)
	return b
}

// Set sets field value by name.
// Call ToBytes() to get modified byte array
// Value for byte array field could be base64 string or []byte
// `Get()` will not consider modifications made by Set, Append, ApplyJSONAndToBytes, ApplyMapBuffer, ApplyMap
// Rewrites previous modifications made by Set, Append, ApplyJSONAndToBytes, ApplyMapBuffer, ApplyMap
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
	m := b.modifiedFields[f.Order]

	if m == nil {
		m = &modifiedField{}
	}

	if bNested, ok := value.(*Buffer); ok {
		if bNested == nil {
			// Set(*Buffer(nil)) could be called at UnmarshalJSONObject()
			// will have problems on ToBytes() because value != nil here
			value = nil
		} else {
			bNested.owner = b
		}
	}

	if m.value != nil {
		if releaseable, ok := m.value.(interface{ Release() }); ok {
			b.toRelease = append(b.toRelease, releaseable)
		}
	}

	m.value = value
	m.isAppend = false
	m.isReleased = false

	b.modifiedFields[f.Order] = m

	b.setModified()
}

// Append appends an array field. toAppend could be a single value or an array of values
// Value for byte array field could be base64 string or []byte
// Rewrites previous modifications made by Set, Append, ApplyJSONAndToBytes, ApplyMapBuffer, ApplyMap
// `Get()` will not consider modifications made by Set, Append, ApplyJSONAndToBytes, ApplyMapBuffer, ApplyMap
// Nil or empty array is provided -> equals to unset the field
func (b *Buffer) Append(name string, toAppend interface{}) {
	f, ok := b.Scheme.FieldsMap[name]
	if !ok {
		return
	}
	b.append(f, toAppend)
}

func (b *Buffer) append(f *Field, toAppend interface{}) {

	b.prepareModifiedFields()

	m := b.modifiedFields[f.Order]

	if m == nil {
		m = &modifiedField{}
	}

	m.value = toAppend
	m.isAppend = true
	m.isReleased = false

	b.modifiedFields[f.Order] = m

	b.setModified()
}

// ApplyMapBuffer modifies Buffer with JSON specified by jsonMap
// ToBytes() will return byte array with initial + applied data
// float is provided for an int field -> no error, integer part only is used (gojay feature), whereas is an error on ApplyMap()
// Values for byte arrays are expected to be base64 strings
// `Get()` will not consider modifications made by Set, Append, ApplyJSONAndToBytes, ApplyMapBuffer, ApplyMap
// Rewrites previous modifications made by Set, Append, ApplyJSONAndToBytes, ApplyMapBuffer, ApplyMap
func (b *Buffer) ApplyMapBuffer(jsonMap []byte) error {
	if len(jsonMap) == 0 {
		return nil
	}
	b.prepareModifiedFields()
	return gojay.UnmarshalJSONObjectWithPool(jsonMap[:], b)
}

// ApplyJSONAndToBytes sets field values described by provided json and returns new FlatBuffer byte array with inital + applied data
// See `ApplyMapBuffer` for details
func (b *Buffer) ApplyJSONAndToBytes(jsonBytes []byte) (res []byte, nilled []string, err error) {
	if err := b.ApplyMapBuffer(jsonBytes); err != nil {
		return nil, nil, err
	}
	return b.ToBytesNilled()
}

// ToBytesNilled constructs resulting byte array (nil values are not stored) and list of field names which were nilled using Set, ApplyJSON, ApplyMap
// useful in cases when we should know which fields were null on load from map or JSON
// nilled fields of nested objects are not considered
func (b *Buffer) ToBytesNilled() (res []byte, nilledFields []string, err error) {
	if res, err = b.ToBytes(); err != nil {
		return
	}
	for i, mf := range b.modifiedFields {
		if mf != nil && !mf.isReleased && mf.value == nil {
			nilledFields = append(nilledFields, b.Scheme.Fields[i].Name)
		}
	}
	return
}

// ApplyMap sets field values described by provided map[string]interface{}
// Resulting buffer has no value (or has nil value) for a mandatory field -> error
// Value type and field type are incompatible (e.g. string for numberic field) -> error
// Value and field types differs but value fits into field -> no error. Examples:
//
//	255 fits into float, double, int, long, byte;
//	256 does not fit into byte
//	math.MaxInt64 does not fit into int32
//
// Unexisting field is provided -> error
// Byte arrays could be base64 strings or []byte
// Array element is nil -> error (not supported)
// Note: float is provided for an int field -> error, whereas no error on ApplyJSONAndToBytes() (gojay feature)
// `Get()` will not consider modifications made by Set, Append, ApplyJSONAndToBytes, ApplyMapBuffer, ApplyMap
// Rewrites previous modifications made by Set, Append, ApplyJSONAndToBytes, ApplyMapBuffer, ApplyMap
// Non-empty array is applied over an existing array -> incomming array is appended to the existing
// Nil or empty array is applied over an existing array -> existing array is unset
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
						buffers.Release()
						return fmt.Errorf("element value of array field %s must be an object, %#v provided", fn, dataNestedIntf)
					}

					buffers.Slice[i] = NewBuffer(f.FieldScheme)
					buffers.Slice[i].owner = b
					if err := buffers.Slice[i].ApplyMap(dataNested); err != nil {
						buffers.Release()
						return err
					}
				}

				b.append(f, buffers)
			} else {
				bNested := NewBuffer(f.FieldScheme)
				bNested.owner = b
				dataNested, ok := fv.(map[string]interface{})
				if !ok {
					bNested.Release()
					return fmt.Errorf("value of field %s must be an object, %#v provided", fn, fv)
				}
				if err := bNested.ApplyMap(dataNested); err != nil {
					bNested.Release()
					return err
				}
				b.set(f, bNested)
			}
		} else if f.IsArray {
			b.append(f, fv)
		} else {
			b.set(f, fv)
		}
	}
	return nil
}

// UnmarshalJSONObject unmarshals JSON into the current Buffer using Gojay. Conforms to gojay.UnmarshalerJSONObject interface
func (b *Buffer) UnmarshalJSONObject(dec *gojay.Decoder, fn string) (err error) {
	f, ok := b.Scheme.FieldsMap[fn]
	if !ok {
		return fmt.Errorf("field %s does not exist in the scheme", fn)
	}
	m := b.modifiedFields[f.Order]
	if f.Ft == FieldTypeObject {
		if f.IsArray {
			buffers := getBufferSlice(0)
			buffers.Scheme = f.FieldScheme
			buffers.Owner = b

			if err = dec.Array(buffers); err != nil {
				buffers.Release()
				return err
			}

			if m != nil {
				if prevBufSlice, ok := m.value.(*buffersSlice); ok {
					b.toRelease = append(b.toRelease, prevBufSlice)
				}
			}

			if len(buffers.Slice) == 0 {
				buffers.Release()
				b.set(f, nil)
			} else {
				b.append(f, buffers)
			}

		} else {
			var bNested *Buffer
			if err = dec.ObjectOrNull(len(f.FieldScheme.Fields), func() gojay.UnmarshalerJSONObject {
				bNested = NewBuffer(f.FieldScheme)
				bNested.owner = b
				bNested.prepareModifiedFields()
				return bNested
			}); err != nil {
				return err
			}

			b.set(f, bNested)
		}
	} else {
		if f.IsArray {
			switch f.Ft {
			case FieldTypeByte:
				base64StrBytes, isNull, err := dec.StringBytesOrNull()
				if err != nil {
					return err
				}
				if len(base64StrBytes) == 0 || isNull {
					b.set(f, nil)
					return nil
				}
				bytes := make([]byte, base64.StdEncoding.DecodedLen(len(base64StrBytes)))
				n, err := base64.StdEncoding.Decode(bytes, base64StrBytes)
				if err != nil {
					return fmt.Errorf("the string %s considered as base64-encoded value for field %s: %s", string(base64StrBytes), f.QualifiedName(), err)
				}
				b.append(f, bytes[:n])
			case FieldTypeBool:
				arr := []bool{}
				if err = dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) (err error) {
					val, isNull, err := dec.BoolOrNull()
					if err != nil {
						return err
					}
					if isNull {
						return nullArrayElementError(f)
					}
					arr = append(arr, val)
					return nil
				})); err == nil {
					if len(arr) == 0 {
						b.set(f, nil)
					} else {
						b.append(f, arr)
					}
				}
			case FieldTypeFloat64:
				arr := []float64{}
				if err = dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) (err error) {
					val, isNull, err := dec.Float64OrNull()
					if err != nil {
						return err
					}
					if isNull {
						return nullArrayElementError(f)
					}
					arr = append(arr, val)
					return nil
				})); err == nil {
					if len(arr) == 0 {
						b.set(f, nil)
					} else {
						b.append(f, arr)
					}
				}
			case FieldTypeFloat32:
				arr := []float32{}
				if err = dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) (err error) {
					val, isNull, err := dec.Float32OrNull()
					if err != nil {
						return err
					}
					if isNull {
						return nullArrayElementError(f)
					}
					arr = append(arr, val)
					return nil
				})); err == nil {
					if len(arr) == 0 {
						b.set(f, nil)
					} else {
						b.append(f, arr)
					}
				}
			case FieldTypeInt32:
				arr := []int32{}
				if err = dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) (err error) {
					val, isNull, err := dec.Int32OrNull()
					if err != nil {
						return err
					}
					if isNull {
						return nullArrayElementError(f)
					}
					arr = append(arr, val)
					return nil
				})); err == nil {
					if len(arr) == 0 {
						b.set(f, nil)
					} else {
						b.append(f, arr)
					}
				}
			case FieldTypeInt64:
				arr := []int64{}
				if err = dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) (err error) {
					val, isNull, err := dec.Int64OrNull()
					if err != nil {
						return err
					}
					if isNull {
						return nullArrayElementError(f)
					}
					arr = append(arr, val)
					return nil
				})); err == nil {
					if len(arr) == 0 {
						b.set(f, nil)
					} else {
						b.append(f, arr)
					}
				}
			case FieldTypeString:
				arr := [][]byte{}
				if err = dec.Array(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) (err error) {
					val, isNull, err := dec.StringBytesOrNull()
					if err != nil {
						return err
					}
					if isNull {
						return nullArrayElementError(f)
					}
					arr = append(arr, val)
					return nil
				})); err == nil {
					if len(arr) == 0 {
						b.set(f, nil)
					} else {
						b.append(f, arr)
					}
				}
			}
			if err != nil {
				return err
			}
		} else {
			var isNull bool
			switch f.Ft {
			case FieldTypeString:
				var val []byte
				if val, isNull, err = dec.StringBytesOrNull(); err == nil && !isNull {
					b.set(f, val)
				}
			case FieldTypeBool:
				var val bool
				if val, isNull, err = dec.BoolOrNull(); err == nil && !isNull {
					b.set(f, val)
				}
			case FieldTypeByte, FieldTypeInt32:
				// ok to write int32 into byte field. Will fail on ToBytes() if value does not fit into byte
				var val int32
				if val, isNull, err = dec.Int32OrNull(); err == nil && !isNull {
					b.set(f, float64(val))
				}
			case FieldTypeFloat64:
				var val float64
				if val, isNull, err = dec.Float64OrNull(); err == nil && !isNull {
					b.set(f, val)
				}
			case FieldTypeFloat32:
				var val float32
				if val, isNull, err = dec.Float32OrNull(); err == nil && !isNull {
					b.set(f, val)
				}
			case FieldTypeInt64:
				var val int64
				if val, isNull, err = dec.Int64OrNull(); err == nil && !isNull {
					b.set(f, val)
				}
			}
			if err == nil && isNull {
				b.set(f, nil)
			}
		}
	}
	return
}

func nullArrayElementError(f *Field) error {
	return fmt.Errorf("null JSON array element is met for field %s", f.QualifiedName())
}

// NKeys returns Schemes's root fields amount. Conforms to gojay.UnmarshalerJSONObject interface
func (b *Buffer) NKeys() int {
	return len(b.Scheme.FieldsMap)
}

// ToBytes returns new FlatBuffer byte array with fields modified by Set() and fields which initially had values
// Note: initial byte array and current modifications are kept
func (b *Buffer) ToBytes() ([]byte, error) {
	if !b.isModified && len(b.tab.Bytes) > 0 {
		return b.tab.Bytes, nil
	}

	b.builder.Reset()

	uOffset, err := b.encodeBuffer(b.builder)
	if err != nil {
		return nil, err
	}

	if uOffset != 0 {
		return b.builder.FinishedBytes(), nil
	}

	return nil, nil
}

// ToBytesWithBuilder same as ToBytes but uses builder
// builder.Reset() is invoked
func (b *Buffer) ToBytesWithBuilder(builder *flatbuffers.Builder) error {
	if !b.isModified && len(b.tab.Bytes) > 0 { // mandatory fields should be checked
		builder.Bytes = b.tab.Bytes
		return nil
	}
	_, err := b.encodeBuffer(builder)
	return err
}

func (b *Buffer) prepareModifiedFields() {
	if len(b.Scheme.Fields) > cap(b.modifiedFields) {
		b.modifiedFields = make([]*modifiedField, len(b.Scheme.Fields))
	} else {
		b.modifiedFields = b.modifiedFields[:len(b.Scheme.Fields)]
	}
}

// if modifiedField is set to non-nil but the value is empty array, object or string -> set modifiedField.value = nil to easier calculate nilledFields at ToBytesNilled()
func (b *Buffer) encodeBuffer(bl *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	offsets := getOffsetSlice(len(b.Scheme.Fields))
	defer putOffsetSlice(offsets)

	b.prepareModifiedFields()

	var err error

	for _, f := range b.Scheme.Fields {
		if f.IsArray {
			arrayUOffsetT := flatbuffers.UOffsetT(0)
			modifiedField := b.modifiedFields[f.Order]
			if modifiedField != nil && !modifiedField.isReleased {
				if modifiedField.value != nil {
					var toAppendToIntf interface{} = nil
					if modifiedField.isAppend {
						toAppendToIntf = b.getArrIntf(f)
					}
					if arrayUOffsetT, err = b.encodeArray(bl, f, modifiedField.value, toAppendToIntf); err != nil {
						return 0, err
					}
					if arrayUOffsetT == 0 {
						if f.Ft == FieldTypeObject {
							b.toRelease = append(b.toRelease, modifiedField.value)
						}
						modifiedField.value = nil
					}
				}
			} else {
				if uOffsetT := b.getFieldUOffsetTByOrder(f.Order); uOffsetT != 0 {
					arrayUOffsetT = b.copyArray(bl, uOffsetT, f)
				}
			}
			(*offsets)[f.Order].arr = arrayUOffsetT
		} else if f.Ft == FieldTypeObject {
			nestedUOffsetT := flatbuffers.UOffsetT(0)
			modifiedField := b.modifiedFields[f.Order]
			if modifiedField != nil && !modifiedField.isReleased {
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
					if nestedUOffsetT == 0 {
						b.toRelease = append(b.toRelease, modifiedField.value)
						modifiedField.value = nil
					}
				}
			} else {
				if uOffsetT := b.getFieldUOffsetTByOrder(f.Order); uOffsetT != 0 {
					bufToWrite := b.getByUOffsetT(f, uOffsetT) // can not be nil
					if storeObjectsAsBytes {
						nestedBytes, _ := bufToWrite.(*Buffer).ToBytes() // no errors should be here
						nestedUOffsetT = bl.CreateByteVector(nestedBytes)
					} else {
						nestedUOffsetT, _ = bufToWrite.(*Buffer).encodeBuffer(bl) // no errors should be here
					}
				}
			}
			(*offsets)[f.Order].obj = nestedUOffsetT
		} else if f.Ft == FieldTypeString {
			stringUOffsetT := flatbuffers.UOffsetT(0)
			modifiedStringField := b.modifiedFields[f.Order]
			if modifiedStringField != nil && !modifiedStringField.isReleased {
				if modifiedStringField.value != nil {
					switch toWrite := modifiedStringField.value.(type) {
					case string:
						if len(toWrite) > 0 {
							stringUOffsetT = bl.CreateString(toWrite)
						}
					case []byte:
						if len(toWrite) > 0 {
							stringUOffsetT = bl.CreateByteString(toWrite)
						}
					default:
						return 0, fmt.Errorf("string required but %#v provided for field %s", modifiedStringField.value, f.QualifiedName())
					}
					if stringUOffsetT == 0 {
						modifiedStringField.value = nil
					}

				}
			} else {
				if offset := b.getFieldUOffsetTByOrder(f.Order); offset != 0 {
					stringUOffsetT = bl.CreateByteString(b.tab.ByteVector(offset))
				}
			}
			(*offsets)[f.Order].str = stringUOffsetT
		}
	}

	isStarted := false
	beforePrepend := func() {
		if !isStarted {
			bl.StartObject(len(b.Scheme.Fields))
			isStarted = true
		}
	}
	for _, f := range b.Scheme.Fields {
		isSet := false
		offsetToWrite := flatbuffers.UOffsetT(0)
		if f.IsArray {
			offsetToWrite = (*offsets)[f.Order].arr
		} else {
			switch f.Ft {
			case FieldTypeString:
				offsetToWrite = (*offsets)[f.Order].str
			case FieldTypeObject:
				offsetToWrite = (*offsets)[f.Order].obj
			default:
				modifiedField := b.modifiedFields[f.Order]
				if modifiedField != nil && !modifiedField.isReleased {
					if isSet = modifiedField.value != nil; isSet {
						if !encodeFixedSizeValue(bl, f, modifiedField.value, beforePrepend) {
							return 0, fmt.Errorf("wrong value %T(%#v) provided for field %s", modifiedField.value, modifiedField.value, f.QualifiedName())
						}
					}
				} else {
					isSet = copyFixedSizeValue(bl, b, f, beforePrepend)
				}
			}
		}
		if f.IsMandatory && offsetToWrite == 0 && !isSet {
			return 0, fmt.Errorf("mandatory field %s is not set", f.QualifiedName())
		}
		if offsetToWrite > 0 {
			beforePrepend()
			bl.PrependUOffsetTSlot(f.Order, offsetToWrite, 0)
		}
	}

	if isStarted {
		res := bl.EndObject()
		bl.Finish(res)
		return res, nil
	}
	return 0, nil
}

// HasValue returns if specified field exists in the scheme and its value is set to non-nil
func (b *Buffer) HasValue(name string) bool {
	return b.getFieldUOffsetT(name) != 0
}

// Reset sets current underlying byte array and clears modified fields. Useful for *Buffer instance reuse
// Note: bytes must match the Buffer's scheme
func (b *Buffer) Reset(bytes []byte) {
	b.tab.Bytes = bytes
	if len(b.tab.Bytes) == 0 {
		b.tab.Pos = 0
	} else {
		b.tab.Pos = flatbuffers.GetUOffsetT(b.tab.Bytes)
	}

	if b.modifiedFields != nil {
		b.releaseFields()
	}

	b.isModified = false
}

func (b *Buffer) copyArray(bl *flatbuffers.Builder, arrayUOffsetT flatbuffers.UOffsetT, f *Field) flatbuffers.UOffsetT {
	l := b.tab.VectorLen(arrayUOffsetT - b.tab.Pos)
	uOffsetT := b.tab.Vector(arrayUOffsetT - b.tab.Pos)
	switch f.Ft {
	case FieldTypeInt32:
		bl.StartVector(flatbuffers.SizeInt32, l, flatbuffers.SizeInt32)
		for i := 0; i < l; i++ {
			elem := b.tab.GetInt32(uOffsetT + flatbuffers.UOffsetT((l-i-1)*flatbuffers.SizeInt32))
			bl.PrependInt32(elem)
		}
		return bl.EndVector(l)
	case FieldTypeBool:
		bl.StartVector(flatbuffers.SizeBool, l, flatbuffers.SizeBool)
		for i := 0; i < l; i++ {
			elem := b.tab.GetBool(uOffsetT + flatbuffers.UOffsetT((l-i-1)*flatbuffers.SizeBool))
			bl.PrependBool(elem)
		}
		return bl.EndVector(l)
	case FieldTypeInt64:
		bl.StartVector(flatbuffers.SizeInt64, l, flatbuffers.SizeInt64)
		for i := 0; i < l; i++ {
			elem := b.tab.GetInt64(uOffsetT + flatbuffers.UOffsetT((l-i-1)*flatbuffers.SizeInt64))
			bl.PrependInt64(elem)
		}
		return bl.EndVector(l)
	case FieldTypeFloat32:
		bl.StartVector(flatbuffers.SizeFloat32, l, flatbuffers.SizeFloat32)
		for i := 0; i < l; i++ {
			elem := b.tab.GetFloat32(uOffsetT + flatbuffers.UOffsetT((l-i-1)*flatbuffers.SizeFloat32))
			bl.PrependFloat32(elem)
		}
		return bl.EndVector(l)
	case FieldTypeFloat64:
		bl.StartVector(flatbuffers.SizeFloat64, l, flatbuffers.SizeFloat64)
		for i := 0; i < l; i++ {
			elem := b.tab.GetFloat64(uOffsetT + flatbuffers.UOffsetT((l-i-1)*flatbuffers.SizeFloat64))
			bl.PrependFloat64(elem)
		}
		return bl.EndVector(l)
	case FieldTypeByte:
		return bl.CreateByteVector(b.tab.Bytes[uOffsetT : l+int(uOffsetT)]) // copied there
	case FieldTypeString:
		os := getUOffsetSlice(l)
		for i := 0; i < l; i++ {
			elementUOffsetT := uOffsetT + flatbuffers.UOffsetT((l-i-1)*flatbuffers.SizeUOffsetT)
			(*os)[i] = bl.CreateByteString(b.tab.ByteVector(elementUOffsetT))
		}
		bl.StartVector(flatbuffers.SizeUOffsetT, l, flatbuffers.SizeUOffsetT)
		for i := 0; i < l; i++ {
			bl.PrependUOffsetT((*os)[i])
		}
		putUOffsetSlice(os)
		return bl.EndVector(l)
	default:
		objectArray := b.getByUOffsetT(f, arrayUOffsetT)
		resUOffsetT, _ := b.encodeArray(bl, f, objectArray, nil) // no errors should be here
		return resUOffsetT
	}
}

func encodeStringArr(f *Field, value interface{}, bl *flatbuffers.Builder, toAppendToIntf interface{}) (flatbuffers.UOffsetT, bool) {
	toAppendTo, _ := toAppendToIntf.(IStringArray)
	toAppendToLen := 0
	if toAppendTo != nil {
		toAppendToLen = toAppendTo.Len()
	}
	switch valArr := value.(type) {
	case []string:
		// on Set, Append, ApplyMap
		if len(valArr) == 0 {
			return 0, true
		}
		l := len(valArr) + toAppendToLen
		stringUOffsetTs := getUOffsetSlice(l)
		for i := 0; i < toAppendToLen; i++ {
			(*stringUOffsetTs)[i] = bl.CreateString(toAppendTo.At(i))
		}
		for i := 0; i < len(valArr); i++ {
			(*stringUOffsetTs)[i+toAppendToLen] = bl.CreateString(valArr[i])
		}
		bl.StartVector(flatbuffers.SizeUOffsetT, l, flatbuffers.SizeUOffsetT)
		for i := 0; i < l; i++ {
			bl.PrependUOffsetT((*stringUOffsetTs)[i])
		}
		putUOffsetSlice(stringUOffsetTs)
		return bl.EndVector(l), true
	case []interface{}:
		// on ApplyMap
		if len(valArr) == 0 {
			return 0, true
		}
		l := len(valArr) + toAppendToLen
		stringUOffsetTs := getUOffsetSlice(l)
		defer putUOffsetSlice(stringUOffsetTs)
		for i := 0; i < toAppendToLen; i++ {
			(*stringUOffsetTs)[i] = bl.CreateString(toAppendTo.At(i))
		}
		for i := 0; i < len(valArr); i++ {
			str, ok := valArr[i].(string)
			if !ok {
				return 0, false
			}
			(*stringUOffsetTs)[i+toAppendToLen] = bl.CreateString(str)
		}
		bl.StartVector(flatbuffers.SizeUOffsetT, l, flatbuffers.SizeUOffsetT)
		for i := 0; i < l; i++ {
			bl.PrependUOffsetT((*stringUOffsetTs)[i])
		}
		return bl.EndVector(l), true
	case [][]byte:
		// came from JSON
		if len(valArr) == 0 {
			return 0, true
		}
		l := len(valArr) + toAppendToLen
		stringUOffsetTs := getUOffsetSlice(l)
		defer putUOffsetSlice(stringUOffsetTs)
		for i := 0; i < toAppendToLen; i++ {
			(*stringUOffsetTs)[i] = bl.CreateString(toAppendTo.At(i))
		}
		for i := 0; i < len(valArr); i++ {
			(*stringUOffsetTs)[i+toAppendToLen] = bl.CreateString(string(valArr[i]))
		}
		bl.StartVector(flatbuffers.SizeUOffsetT, l, flatbuffers.SizeUOffsetT)
		for i := 0; i < l; i++ {
			bl.PrependUOffsetT((*stringUOffsetTs)[i])
		}
		return bl.EndVector(l), true
	}
	return 0, false
}

func encodeByteArr(f *Field, value interface{}, bl *flatbuffers.Builder, toAppendToIntf interface{}) (flatbuffers.UOffsetT, bool) {
	toAppendTo, _ := toAppendToIntf.(IByteArray)
	var target []byte
	switch arr := value.(type) {
	case []byte:
		target = arr
	case string:
		var err error
		if target, err = base64.StdEncoding.DecodeString(arr); err != nil {
			return 0, false
		}
	default:
		return 0, false
	}
	if len(target) == 0 {
		return 0, true
	}
	if toAppendTo != nil {
		target = append(toAppendTo.Bytes(), target...)
	}
	return bl.CreateByteVector(target), true
}

func encodeInt32Arr(f *Field, value interface{}, bl *flatbuffers.Builder, toAppendToIntf interface{}) (flatbuffers.UOffsetT, bool) {
	toAppendTo, _ := toAppendToIntf.(IInt32Array)
	toAppendToLen := 0
	if toAppendTo != nil {
		toAppendToLen = toAppendTo.Len()
	}
	if arr, ok := value.([]int32); !ok {
		intfs, ok := value.([]interface{})
		if !ok {
			return 0, false
		}
		if len(intfs) == 0 {
			return 0, true // nil or empty array is set or appended -> unset
		}
		l := len(intfs) + toAppendToLen
		bl.StartVector(flatbuffers.SizeInt32, l, flatbuffers.SizeInt32)
		for i := 0; i < toAppendToLen; i++ {
			bl.PrependInt32(toAppendTo.At(i))
		}
		for _, intf := range intfs {
			float64Src, ok := intf.(float64)
			if !ok || !IsFloat64ValueFitsIntoField(f, float64Src) {
				return 0, false
			}
			bl.PrependInt32(int32(float64Src))
		}
		return bl.EndVector(l), true
	} else {
		if len(arr) == 0 {
			return 0, true
		}
		l := len(arr) + toAppendToLen
		bl.StartVector(flatbuffers.SizeInt32, l, flatbuffers.SizeInt32)
		for i := 0; i < toAppendToLen; i++ {
			bl.PrependInt32(toAppendTo.At(i))
		}
		for _, int32Elem := range arr {
			bl.PrependInt32(int32Elem)
		}
		return bl.EndVector(l), true
	}
}

func encodeBoolArr(f *Field, value interface{}, bl *flatbuffers.Builder, toAppendToIntf interface{}) (flatbuffers.UOffsetT, bool) {
	toAppendTo, _ := toAppendToIntf.(IBoolArray)
	toAppendToLen := 0
	if toAppendTo != nil {
		toAppendToLen = toAppendTo.Len()
	}
	if arr, ok := value.([]bool); !ok {
		intfs, ok := value.([]interface{})
		if !ok {
			return 0, false
		}
		if len(intfs) == 0 {
			return 0, true
		}
		l := len(intfs) + toAppendToLen
		bl.StartVector(flatbuffers.SizeBool, l, flatbuffers.SizeBool)
		for i := 0; i < toAppendToLen; i++ {
			bl.PrependBool(toAppendTo.At(i))
		}
		for _, intf := range intfs {
			b, ok := intf.(bool)
			if !ok {
				return 0, false
			}
			bl.PrependBool(b)
		}
		return bl.EndVector(l), true
	} else {
		if len(arr) == 0 {
			return 0, true
		}
		l := len(arr) + toAppendToLen
		bl.StartVector(flatbuffers.SizeBool, l, flatbuffers.SizeBool)
		for i := 0; i < toAppendToLen; i++ {
			bl.PrependBool(toAppendTo.At(i))
		}
		for _, boolElem := range arr {
			bl.PrependBool(boolElem)
		}
		return bl.EndVector(l), true
	}
}

func encodeInt64Arr(f *Field, value interface{}, bl *flatbuffers.Builder, toAppendToIntf interface{}) (flatbuffers.UOffsetT, bool) {
	toAppendTo, _ := toAppendToIntf.(IInt64Array)
	toAppendToLen := 0
	if toAppendTo != nil {
		toAppendToLen = toAppendTo.Len()
	}
	if arr, ok := value.([]int64); !ok {
		intfs, ok := value.([]interface{})
		if !ok {
			return 0, false
		}
		if len(intfs) == 0 {
			return 0, true
		}
		l := len(intfs) + toAppendToLen
		bl.StartVector(flatbuffers.SizeInt64, l, flatbuffers.SizeInt64)
		for i := 0; i < toAppendToLen; i++ {
			bl.PrependInt64(toAppendTo.At(i))
		}
		for _, intf := range intfs {
			float64Src, ok := intf.(float64)
			if !ok || !IsFloat64ValueFitsIntoField(f, float64Src) {
				return 0, false
			}
			bl.PrependInt64(int64(float64Src))
		}
		return bl.EndVector(l), true
	} else {
		if len(arr) == 0 {
			return 0, true
		}
		l := len(arr) + toAppendToLen
		bl.StartVector(flatbuffers.SizeInt64, l, flatbuffers.SizeInt64)
		for i := 0; i < toAppendToLen; i++ {
			bl.PrependInt64(toAppendTo.At(i))
		}
		for _, int64Elem := range arr {
			bl.PrependInt64(int64Elem)
		}
		return bl.EndVector(l), true
	}
}

func encodeFloat32Arr(f *Field, value interface{}, bl *flatbuffers.Builder, toAppendToIntf interface{}) (flatbuffers.UOffsetT, bool) {
	toAppendTo, _ := toAppendToIntf.(IFloat32Array)
	toAppendToLen := 0
	if toAppendTo != nil {
		toAppendToLen = toAppendTo.Len()
	}
	if arr, ok := value.([]float32); !ok {
		intfs, ok := value.([]interface{})
		if !ok {
			return 0, false
		}
		if len(intfs) == 0 {
			return 0, true
		}
		l := len(intfs) + toAppendToLen
		bl.StartVector(flatbuffers.SizeFloat32, l, flatbuffers.SizeFloat32)
		for i := 0; i < toAppendToLen; i++ {
			bl.PrependFloat32(toAppendTo.At(i))
		}
		for _, intf := range intfs {
			float64Src, ok := intf.(float64)
			if !ok || !IsFloat64ValueFitsIntoField(f, float64Src) {
				return 0, false
			}
			bl.PrependFloat32(float32(float64Src))
		}
		return bl.EndVector(l), true
	} else {
		if len(arr) == 0 {
			return 0, true
		}
		l := len(arr) + toAppendToLen
		bl.StartVector(flatbuffers.SizeFloat32, l, flatbuffers.SizeFloat32)
		for i := 0; i < toAppendToLen; i++ {
			bl.PrependFloat32(toAppendTo.At(i))
		}
		for _, float32Elem := range arr {
			bl.PrependFloat32(float32Elem)
		}
		return bl.EndVector(l), true
	}
}

func encodeFloat64Arr(f *Field, value interface{}, bl *flatbuffers.Builder, toAppendToIntf interface{}) (flatbuffers.UOffsetT, bool) {
	toAppendTo, _ := toAppendToIntf.(IFloat64Array)
	toAppendToLen := 0
	if toAppendTo != nil {
		toAppendToLen = toAppendTo.Len()
	}
	if arr, ok := value.([]float64); !ok {
		intfs, ok := value.([]interface{})
		if !ok {
			return 0, false
		}
		if len(intfs) == 0 {
			return 0, true
		}
		l := len(intfs) + toAppendToLen
		bl.StartVector(flatbuffers.SizeFloat64, l, flatbuffers.SizeFloat64)
		for i := 0; i < toAppendToLen; i++ {
			bl.PrependFloat64(toAppendTo.At(i))
		}
		for _, intf := range intfs {
			float64Src, ok := intf.(float64)
			if !ok {
				return 0, false
			}
			bl.PrependFloat64(float64Src)
		}
		return bl.EndVector(l), true
	} else {
		if len(arr) == 0 {
			return 0, true
		}
		l := len(arr) + toAppendToLen
		bl.StartVector(flatbuffers.SizeFloat64, l, flatbuffers.SizeFloat64)
		for i := 0; i < toAppendToLen; i++ {
			bl.PrependFloat64(toAppendTo.At(i))
		}
		for _, float64Elem := range arr {
			bl.PrependFloat64(float64Elem)
		}
		return bl.EndVector(l), true
	}
}

func (b *Buffer) encodeArray(bl *flatbuffers.Builder, f *Field, value interface{}, toAppendToIntf interface{}) (uOffsetT flatbuffers.UOffsetT, err error) {
	ok := false
	switch f.Ft {
	case FieldTypeInt32:
		uOffsetT, ok = encodeInt32Arr(f, value, bl, toAppendToIntf)
	case FieldTypeBool:
		uOffsetT, ok = encodeBoolArr(f, value, bl, toAppendToIntf)
	case FieldTypeInt64:
		uOffsetT, ok = encodeInt64Arr(f, value, bl, toAppendToIntf)
	case FieldTypeFloat32:
		uOffsetT, ok = encodeFloat32Arr(f, value, bl, toAppendToIntf)
	case FieldTypeFloat64:
		uOffsetT, ok = encodeFloat64Arr(f, value, bl, toAppendToIntf)
	case FieldTypeByte:
		uOffsetT, ok = encodeByteArr(f, value, bl, toAppendToIntf)
	case FieldTypeString:
		uOffsetT, ok = encodeStringArr(f, value, bl, toAppendToIntf)
	default:
		nestedUOffsetTs := getUOffsetSlice(0)
		defer putUOffsetSlice(nestedUOffsetTs)
		switch arr := value.(type) {
		case *buffersSlice:
			// came on ApplyMap() or UnmarshalJSONObjectWithPool()
			for i := 0; i < len(arr.Slice); i++ {
				if arr.Slice[i] == nil {
					return 0, fmt.Errorf("nil element of array field %s is provided. Nils are not supported for array elements", f.QualifiedName())
				}
				if storeObjectsAsBytes {
					nestedBytes, err := arr.Slice[i].ToBytes()
					if err != nil {
						return 0, err
					}
					*nestedUOffsetTs = append(*nestedUOffsetTs, bl.CreateByteVector(nestedBytes))
				} else {
					nestedUOffsetT, err := arr.Slice[i].encodeBuffer(bl)
					if err != nil {
						return 0, err
					}
					*nestedUOffsetTs = append(*nestedUOffsetTs, nestedUOffsetT)
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
					*nestedUOffsetTs = append(*nestedUOffsetTs, bl.CreateByteVector(nestedBytes))
				} else {
					nestedUOffsetT, err := arr[i].encodeBuffer(bl)
					if err != nil {
						return 0, err
					}
					*nestedUOffsetTs = append(*nestedUOffsetTs, nestedUOffsetT)
				}
			}
		case *ObjectArray:
			// re-encoding existing array
			for arr.Next() {
				if storeObjectsAsBytes {
					nestedBytes, _ := arr.Buffer.ToBytes()
					*nestedUOffsetTs = append(*nestedUOffsetTs, bl.CreateByteVector(nestedBytes))
				} else {
					nestedUOffsetT, _ := arr.Buffer.encodeBuffer(bl) // should be no errors here
					*nestedUOffsetTs = append(*nestedUOffsetTs, nestedUOffsetT)
				}
			}

		default:
			return 0, fmt.Errorf("%#v provided for field %s is not an array of nested objects", value, f.QualifiedName())
		}

		if len(*nestedUOffsetTs) == 0 {
			return 0, nil // empty nested objects array -> skip
		}

		if toAppendToIntf != nil {
			toAppendToArr := toAppendToIntf.(*ObjectArray)

			toAppendToUOffsetTs := getUOffsetSlice(toAppendToArr.Len)
			defer putUOffsetSlice(toAppendToUOffsetTs)

			for i := 0; toAppendToArr.Next(); i++ {
				if storeObjectsAsBytes {
					bufBytes, _ := toAppendToArr.Buffer.ToBytes()
					(*toAppendToUOffsetTs)[i] = bl.CreateByteVector(bufBytes)
				} else {
					(*toAppendToUOffsetTs)[i], _ = toAppendToArr.Buffer.encodeBuffer(bl)
				}
			}

			*toAppendToUOffsetTs = append(*toAppendToUOffsetTs, *nestedUOffsetTs...)

			nestedUOffsetTs = toAppendToUOffsetTs
		}

		bl.StartVector(flatbuffers.SizeUOffsetT, len(*nestedUOffsetTs), flatbuffers.SizeUOffsetT)
		for i := 0; i < len(*nestedUOffsetTs); i++ {
			bl.PrependUOffsetT((*nestedUOffsetTs)[i])
		}

		o := bl.EndVector(len(*nestedUOffsetTs))

		return o, nil
	}
	if !ok {
		if f.Ft == FieldTypeString {
			return 0, fmt.Errorf("%#v provided for field %s which can not be converted to []string", value, f.QualifiedName())
		} else if f.Ft == FieldTypeByte {
			return 0, fmt.Errorf("[]byte or base64-encoded string required but %#v provided for field %s", value, f.QualifiedName())
		}
		return 0, fmt.Errorf("[]%s required but %#v provided for field %s", fieldTypesNamesMap[f.Ft], value, f.QualifiedName())
	}
	return
}

func copyFixedSizeValue(dest *flatbuffers.Builder, src *Buffer, f *Field, beforePrepend func()) bool {
	offset := src.getFieldUOffsetTByOrder(f.Order)
	if offset == 0 {
		return false
	}
	beforePrepend()
	switch f.Ft {
	case FieldTypeInt32:
		dest.PrependInt32(src.tab.GetInt32(offset))
	case FieldTypeInt64:
		dest.PrependInt64(src.tab.GetInt64(offset))
	case FieldTypeFloat32:
		dest.PrependFloat32(src.tab.GetFloat32(offset))
	case FieldTypeFloat64:
		dest.PrependFloat64(src.tab.GetFloat64(offset))
	case FieldTypeByte:
		dest.PrependByte(src.tab.GetByte(offset))
	case FieldTypeBool:
		dest.PrependBool(src.tab.GetBool(offset))
	}
	dest.Slot(f.Order)
	return true
}

// IsFloat64ValueFitsIntoField checks if target type of field enough to fit float64 value
// e.g. float64(1) could be applied to any numeric field, float64(256) to any numeric except FieldTypeByte etc
// Useful to check float64 values came from JSON
func IsFloat64ValueFitsIntoField(f *Field, float64Src float64) bool {
	if float64Src == 0 {
		return true
	}
	if float64Src == float64(int32(float64Src)) {
		res := f.Ft == FieldTypeInt32 || f.Ft == FieldTypeInt64 || f.Ft == FieldTypeFloat64 || f.Ft == FieldTypeFloat32
		if float64Src > 0 && float64Src <= 255 {
			return res || f.Ft == FieldTypeByte
		}
		return res
	} else if float64Src == float64(int64(float64Src)) {
		return f.Ft == FieldTypeInt64 || f.Ft == FieldTypeFloat64
	} else {
		return f.Ft == FieldTypeFloat64 || f.Ft == FieldTypeFloat32
	}
}

func encodeFixedSizeValue(bl *flatbuffers.Builder, f *Field, value interface{}, beforePrepend func()) bool {
	switch val := value.(type) {
	case bool:
		if f.Ft != FieldTypeBool {
			return false
		}
		beforePrepend()
		bl.PrependBool(val)
	case float64:
		if !IsFloat64ValueFitsIntoField(f, val) {
			return false
		}
		switch f.Ft {
		case FieldTypeInt32:
			beforePrepend()
			bl.PrependInt32(int32(val))
		case FieldTypeInt64:
			beforePrepend()
			bl.PrependInt64(int64(val))
		case FieldTypeFloat32:
			beforePrepend()
			bl.PrependFloat32(float32(val))
		case FieldTypeFloat64:
			beforePrepend()
			bl.PrependFloat64(val)
		default:
			beforePrepend()
			bl.PrependByte(byte(val))
		}
	case float32:
		if f.Ft != FieldTypeFloat32 {
			return false
		}
		beforePrepend()
		bl.PrependFloat32(val)
	case int64:
		if f.Ft != FieldTypeInt64 {
			return false
		}
		beforePrepend()
		bl.PrependInt64(val)
	case int32:
		if f.Ft != FieldTypeInt32 {
			return false
		}
		beforePrepend()
		bl.PrependInt32(val)
	case byte:
		if f.Ft != FieldTypeByte {
			return false
		}
		beforePrepend()
		bl.PrependByte(val)
	case int:
		switch f.Ft {
		case FieldTypeInt32:
			if math.Abs(float64(val)) > math.MaxInt32 {
				return false
			}
			beforePrepend()
			bl.PrependInt32(int32(val))
		case FieldTypeInt64:
			if math.Abs(float64(val)) > math.MaxInt64 {
				return false
			}
			beforePrepend()
			bl.PrependInt64(int64(val))
		case FieldTypeFloat32:
			if math.Abs(float64(val)) > math.MaxFloat32 {
				return false
			}
			beforePrepend()
			bl.PrependFloat32(float32(val))
		case FieldTypeFloat64:
			beforePrepend()
			bl.PrependFloat64(float64(val))
		default:
			if math.Abs(float64(val)) > 255 {
				return false
			}
			beforePrepend()
			bl.PrependByte(byte(val))
		}
	default:
		return false
	}
	bl.Slot(f.Order)
	return true
}

// IsNil returns if current buffer means nothing
// need to comply to gojay.MarshalerJSONObject
func (b *Buffer) IsNil() bool {
	if b.isModified {
		return false
	}
	if len(b.tab.Bytes) > 0 {
		// same approach is used on GetNames()
		vTable := flatbuffers.UOffsetT(flatbuffers.SOffsetT(b.tab.Pos) - b.tab.GetSOffsetT(b.tab.Pos))
		vOffsetT := b.tab.GetVOffsetT(vTable)
		for vTableOffset := flatbuffers.VOffsetT(4); vTableOffset < vOffsetT; vTableOffset += 2 {
			if b.tab.GetVOffsetT(vTable+flatbuffers.UOffsetT(vTableOffset)) > 0 {
				return false
			}
		}
	}
	return true
}

// MarshalJSONObject encodes current Buffer into JSON using gojay. Complies to gojay.MarshalerJSONObject interface
func (b *Buffer) MarshalJSONObject(enc *gojay.Encoder) {
	b.prepareModifiedFields()
	for _, f := range b.Scheme.Fields {
		var value interface{}
		modifiedField := b.modifiedFields[f.Order]
		if modifiedField != nil && !modifiedField.isReleased {
			value = modifiedField.value
		} else {
			if f.IsArray {
				value = b.getArrIntf(f)
			} else {
				value = b.getByField(f)
			}
		}
		if value == nil {
			continue
		}
		arrLen := 0
		if f.IsArray {
			arrLen = getArrayLen(value)
			if arrLen == 0 {
				continue
			}
		}
		if f.Ft == FieldTypeObject {
			if f.IsArray {
				switch arr := value.(type) {
				case *ObjectArray:
					enc.AddArrayKey(f.Name, gojay.EncodeArrayFunc(func(enc *gojay.Encoder) {
						for arr.Next() {
							enc.AddObject(arr.Buffer)
						}
					}))
				case *buffersSlice:
					enc.AddArrayKey(f.Name, gojay.EncodeArrayFunc(func(enc *gojay.Encoder) {
						for _, buffer := range arr.Slice {
							if buffer != nil {
								enc.AddObject(buffer)
							}
						}
					}))
				case []*Buffer:
					enc.AddArrayKey(f.Name, gojay.EncodeArrayFunc(func(enc *gojay.Encoder) {
						for _, buffer := range arr {
							if buffer != nil {
								enc.AddObject(buffer)
							}
						}
					}))
				}
			} else {
				b := value.(*Buffer)
				if !b.IsNil() {
					enc.AddObjectKey(f.Name, b)
				}
			}
		} else {
			if f.IsArray {
				var encodeFunc func(i int, enc *gojay.Encoder)
				switch f.Ft {
				case FieldTypeString:
					encodeFunc = func(i int, enc *gojay.Encoder) {
						switch arr := value.(type) {
						case IStringArray:
							enc.String(arr.At(i))
						case []string:
							enc.String(arr[i])
						}
					}
				case FieldTypeInt32:
					encodeFunc = func(i int, enc *gojay.Encoder) {
						switch arr := value.(type) {
						case IInt32Array:
							enc.Int32(arr.At(i))
						case []int32:
							enc.Int32(arr[i])
						}
					}
				case FieldTypeBool:
					encodeFunc = func(i int, enc *gojay.Encoder) {
						switch arr := value.(type) {
						case IBoolArray:
							enc.Bool(arr.At(i))
						case []bool:
							enc.Bool(arr[i])
						}
					}
				case FieldTypeFloat64:
					encodeFunc = func(i int, enc *gojay.Encoder) {
						switch arr := value.(type) {
						case IFloat64Array:
							enc.Float64(arr.At(i))
						case []float64:
							enc.Float64(arr[i])
						}
					}
				case FieldTypeFloat32:
					encodeFunc = func(i int, enc *gojay.Encoder) {
						switch arr := value.(type) {
						case IFloat32Array:
							enc.Float32(arr.At(i))
						case []float32:
							enc.Float32(arr[i])
						}
					}
				case FieldTypeInt64:
					encodeFunc = func(i int, enc *gojay.Encoder) {
						switch arr := value.(type) {
						case IInt64Array:
							enc.Int64(arr.At(i))
						case []int64:
							enc.Int64(arr[i])
						}
					}
				}
				if f.Ft == FieldTypeByte {
					// note: val is always []byte here. base64 string decoded to []byte on UnmarshalJSONObject()
					var bytes []byte
					switch val := value.(type) {
					case []byte:
						bytes = val
					case IByteArray:
						bytes = val.Bytes()
					}
					base64Str := base64.StdEncoding.EncodeToString(bytes)
					enc.StringKey(f.Name, base64Str)
				} else {
					enc.ArrayKey(f.Name, gojay.EncodeArrayFunc(func(enc *gojay.Encoder) {
						for i := 0; i < arrLen; i++ {
							encodeFunc(i, enc)
						}
					}))
				}
			} else {
				enc.AddInterfaceKey(f.Name, value)
			}
		}
	}
}

// ToJSON returns JSON key->value string
// empty buffer -> "{}"
func (b *Buffer) ToJSON() []byte {
	buf := bytes.NewBuffer(nil)
	enc := gojay.BorrowEncoder(buf)
	defer enc.Release()
	enc.EncodeObject(b)
	return buf.Bytes()
}

// GetBytes is an alias for ToBytes(). Simply returns underlying buffer if no modifications. Returns nil on error
func (b *Buffer) GetBytes() []byte {
	bytes, _ := b.ToBytes()
	return bytes
}

// ToJSONMap returns map[string]interface{} representation of the buffer compatible to json
// result map is built using inital data + current modifications
// numeric field types are kept (not float64 as json.Unmarshal() does)
// nested object, array, array element is empty or nil -> skip
// empty buffer -> empty map is returned
func (b *Buffer) ToJSONMap() map[string]interface{} {
	res := map[string]interface{}{}
	b.prepareModifiedFields()
	for _, f := range b.Scheme.Fields {
		var storedVal interface{}
		modifiedField := b.modifiedFields[f.Order]
		if modifiedField != nil && !modifiedField.isReleased {
			storedVal = modifiedField.value
		} else {
			storedVal = b.getByField(f)
		}
		if storedVal == nil {
			continue
		}
		if f.Ft == FieldTypeObject {
			if f.IsArray {
				targetArr := []interface{}{}
				switch arr := storedVal.(type) {
				case *ObjectArray:
					for arr.Next() {
						if elem := arr.Buffer.ToJSONMap(); len(elem) > 0 {
							targetArr = append(targetArr, elem)
						}
					}
				case *buffersSlice:
					// came on ApplyMap()
					buffers, _ := storedVal.(*buffersSlice)
					for _, buffer := range buffers.Slice {
						if elem := buffer.ToJSONMap(); len(elem) > 0 {
							targetArr = append(targetArr, elem)
						}
					}
				case []*Buffer:
					// explicit Set() was called
					for _, buffer := range arr {
						if elem := buffer.ToJSONMap(); len(elem) > 0 {
							targetArr = append(targetArr, elem)
						}
					}
				}
				if len(targetArr) > 0 {
					res[f.Name] = targetArr
				}
			} else {
				if nested := storedVal.(*Buffer).ToJSONMap(); len(nested) > 0 {
					res[f.Name] = nested
				}
			}
		} else {
			if f.Ft == FieldTypeByte && storedVal == "" {
				// empty base64 string for byte array -> no bytes, skip
				continue
			}
			// speed is the same as on explicit storedVal to slice casting + len check
			v := reflect.ValueOf(storedVal)
			if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
				if v.Len() == 0 {
					// empty array -> skip
					continue
				}
			}
			res[f.Name] = storedVal
		}
	}
	return res
}

// IterateFields calls `callback` for each fields which has a value.
// `names` empty -> calback is called for all fields which has a value
// `names` not empty -> callback is called for each specified name if according field has a value
// callbeck returns false -> iteration stops
func (b *Buffer) IterateFields(names []string, callback func(name string, value interface{}) bool) {
	if len(b.tab.Bytes) == 0 {
		return
	}
	if len(names) == 0 {
		for _, f := range b.Scheme.Fields {
			if value := b.getByField(f); value != nil {
				if !callback(f.Name, value) {
					return
				}
			}
		}
	} else {
		for _, name := range names {
			if val := b.Get(name); val != nil {
				if !callback(name, val) {
					return
				}
			}
		}
	}
}

func (b *Buffer) IsModified() bool {
	return b.isModified
}

// applies all current modifications to the underlying byte array and clears all modifications
// note: takes one byte array allocation
func (b *Buffer) CommitChanges() error {
	bytes, err := b.ToBytes()
	if err != nil {
		return err
	}
	bytes = copyBytes(bytes)
	b.Reset(bytes)
	return nil
}

// NewScheme creates new empty Scheme
func NewScheme() *Scheme {
	return &Scheme{"", map[string]*Field{}, []*Field{}}
}

// AddField adds field
func (s *Scheme) AddField(name string, ft FieldType, isMandatory bool) *Scheme {
	s.AddFieldC(name, ft, nil, isMandatory, false)
	return s
}

// AddArray adds array field
func (s *Scheme) AddArray(name string, elementType FieldType, isMandatory bool) *Scheme {
	s.AddFieldC(name, elementType, nil, isMandatory, true)
	return s
}

// AddNested adds nested object field
func (s *Scheme) AddNested(name string, nested *Scheme, isMandatory bool) *Scheme {
	s.AddFieldC(name, FieldTypeObject, nested, isMandatory, false)
	return s
}

// AddNestedArray adds array of nested objects field
func (s *Scheme) AddNestedArray(name string, nested *Scheme, isMandatory bool) *Scheme {
	s.AddFieldC(name, FieldTypeObject, nested, isMandatory, true)
	return s
}

// AddFieldC adds new finely-tuned field
func (s *Scheme) AddFieldC(name string, ft FieldType, nested *Scheme, isMandatory bool, IsArray bool) *Scheme {
	newField := &Field{name, ft, len(s.FieldsMap), isMandatory, nested, s, IsArray}
	s.FieldsMap[name] = newField
	s.Fields = append(s.Fields, newField)
	return s
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
					val, _ = f.FieldScheme.MarshalYAML() // no errors possible
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

// UnmarshalYAML unmarshals Scheme from yaml. Conforms to yaml.Unmarshaler interface
// fields will be replaced with the ones came from the yaml
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
//
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
			if err != nil {
				return nil, err
			}
			nestedScheme.Name = fieldName
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

func getArrayLen(value interface{}) int {
	switch typed := value.(type) {
	case interface{ Len() int }:
		return typed.Len()
	case string:
		return len(typed)
	case *ObjectArray:
		return typed.Len
	case *buffersSlice:
		return len(typed.Slice)
	}
	v := reflect.ValueOf(value)
	if v.Type().Kind() == reflect.Array || v.Type().Kind() == reflect.Slice {
		return v.Len()
	}
	return 0
}

func init() {
	for name, ft := range yamlFieldTypesMap {
		fieldTypesNamesMap[ft] = name
	}
}

func copyBytes(src []byte) []byte {
	res := make([]byte, len(src))
	copy(res, src)
	return res
}
