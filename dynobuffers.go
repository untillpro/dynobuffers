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
	"strings"
	"unicode"

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
	Name        string
	Ft          FieldType
	order       int
	IsMandatory bool
	scheme      *Scheme // != nil for FieldTypeNested only
}

type modifiedField struct {
	Field
	value       interface{}
	strUOffsetT flatbuffers.UOffsetT
}

// Scheme s.e.
type Scheme struct {
	fieldsMap    map[string]*Field
	Fields       []*Field
	stringFields []*Field
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
		return string(b.tab.ByteVector(o)), true
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
	return string(b.tab.ByteVector(o)), true
}

func (b *Buffer) getByField(f *Field) interface{} {
	uOffsetT := b.getFieldUOffsetTByOrder(f.order)
	if uOffsetT == 0 {
		return nil
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
		if len(b.modifiedFields) == 0 {
			b.modifiedFields = make([]*modifiedField, len(b.scheme.Fields))
		}
		if b.modifiedFields[f.order] == nil {
			b.set(f, res)
		}
		return res
	default:
		return string(b.tab.ByteVector(uOffsetT))
	}
}

// Get returns stored field value by name.
// nil -> field is unset or no such field in the scheme
func (b *Buffer) Get(name string) interface{} {
	f, ok := b.scheme.fieldsMap[name]
	if !ok {
		return nil
	}
	return b.getByField(f)

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
	if len(b.modifiedFields) == 0 {
		b.modifiedFields = make([]*modifiedField, len(b.scheme.Fields))
	}
	b.modifiedFields[f.order] = &modifiedField{Field{f.Name, f.Ft, f.order, f.IsMandatory, nil}, value, 0}
}

// ApplyJSONAndToBytes sets field values described by provided json and returns new FlatBuffer byte array
// returns error if value of incompatible type is provided or mandatory field is not set\unset
// no error if unexisting field is provided
func (b *Buffer) ApplyJSONAndToBytes(jsonBytes []byte) ([]byte, error) {
	dest := map[string]interface{}{}
	err := json.Unmarshal(jsonBytes, &dest)
	if err != nil {
		return nil, err
	}
	err = b.applyMap(dest)
	if err != nil {
		return nil, err
	}

	return b.ToBytes()
}

func (b *Buffer) applyMap(data map[string]interface{}) error {
	for fn, fv := range data {
		f, ok := b.scheme.fieldsMap[fn]
		if !ok {
			continue
		}
		if f.Ft == FieldTypeNested {
			bNested := NewBuffer(f.scheme)
			dataNested, ok := fv.(map[string]interface{})
			if !ok {
				return fmt.Errorf("value of field %s must be an object, %v provided", fn, fv)
			}
			bNested.applyMap(dataNested)
			b.Set(f.Name, bNested)
		} else {
			if !b.scheme.canBeAssigned(f, fv) {
				return fmt.Errorf("value %v can not be assigned to field %s", fv, fn)
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

func (b *Buffer) encodeBuffer(bl *flatbuffers.Builder) (flatbuffers.UOffsetT, error) {
	strUOffsetTs := make([]flatbuffers.UOffsetT, len(b.scheme.Fields))
	nestedBuffers := make([]flatbuffers.UOffsetT, len(b.scheme.Fields))
	if len(b.modifiedFields) == 0 {
		b.modifiedFields = make([]*modifiedField, len(b.scheme.Fields))
	}

	for _, f := range b.scheme.Fields {
		if f.Ft != FieldTypeNested {
			continue
		}
		nestedUOffsetT := flatbuffers.UOffsetT(0)
		modifiedField := b.modifiedFields[f.order]
		if modifiedField != nil && modifiedField.value != nil {
			nestedBuffer := modifiedField.value.(*Buffer)
			nestedUOffsetTNew, err := nestedBuffer.encodeBuffer(bl)
			if err != nil {
				return 0, err
			}
			nestedUOffsetT = nestedUOffsetTNew
		}
		nestedBuffers[f.order] = nestedUOffsetT
	}

	for _, stringField := range b.scheme.stringFields {
		modifiedStringField := b.modifiedFields[stringField.order]
		if modifiedStringField != nil {
			if modifiedStringField.value != nil {
				strUOffsetTs[stringField.order] = bl.CreateString(modifiedStringField.value.(string))
			}
		} else {
			if strToWrite, ok := b.getByStringField(stringField); ok {
				strUOffsetTs[stringField.order] = bl.CreateString(strToWrite)
			}
		}
	}

	bl.StartObject(len(b.scheme.fieldsMap))
	for _, f := range b.scheme.Fields {
		isSet := false
		switch f.Ft {
		case FieldTypeString:
			if strUOffsetTs[f.order] > 0 {
				bl.PrependUOffsetTSlot(f.order, strUOffsetTs[f.order], 0)
				isSet = true
			}
		case FieldTypeNested:
			if nestedBuffers[f.order] > 0 {
				bl.PrependUOffsetTSlot(f.order, nestedBuffers[f.order], 0)
				isSet = true
			}
		default:
			modifiedField := b.modifiedFields[f.order]
			if modifiedField != nil {
				if modifiedField.value != nil {
					encodeValue(bl, f, modifiedField.value)
					isSet = true
				}
			} else {
				isSet = copyNonStringField(bl, b, f)
			}
		}
		if f.IsMandatory && !isSet {
			return 0, fmt.Errorf("Field %s is mandatory but not set", f.Name)
		}
	}
	res := bl.EndObject()
	bl.Finish(res)
	return res, nil
}

func copyNonStringField(dest *flatbuffers.Builder, src *Buffer, f *Field) bool {
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

func numberToFloat64(number interface{}) float64 {
	switch number.(type) {
	case float64:
		return number.(float64)
	case float32:
		return float64(number.(float32))
	case int64:
		return float64(number.(int64))
	case int32:
		return float64(number.(int32))
	case int:
		return float64(number.(int))
	default:
		return float64(number.(byte))
	}
}

func encodeValue(bl *flatbuffers.Builder, f *Field, value interface{}) {
	switch f.Ft {
	case FieldTypeInt:
		bl.PrependInt32(int32(numberToFloat64(value)))
	case FieldTypeLong:
		bl.PrependInt64(int64(numberToFloat64(value)))
	case FieldTypeFloat:
		bl.PrependFloat32(float32(numberToFloat64(value)))
	case FieldTypeDouble:
		bl.PrependFloat64(float64(numberToFloat64(value)))
	case FieldTypeByte:
		bl.PrependByte(byte(numberToFloat64(value)))
	case FieldTypeBool:
		bl.PrependBool(value.(bool))
	}
	bl.Slot(f.order)
}

// ToJSON returns JSON flat key->value string
func (b *Buffer) ToJSON() string {
	buf := bytes.NewBufferString("")
	e := json.NewEncoder(buf)
	buf.WriteString("{")
	for _, f := range b.scheme.Fields {
		var value interface{}
		if len(b.modifiedFields) == 0 {
			value = b.getByField(f)
		} else {
			modifiedField := b.modifiedFields[f.order]
			if modifiedField != nil {
				value = modifiedField.value
			} else {
				value = b.getByField(f)
			}
		}
		if value != nil {
			buf.WriteString("\"" + f.Name + "\":")
			if f.Ft == FieldTypeNested {
				buf.WriteString(value.(*Buffer).ToJSON())
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
	return &Scheme{map[string]*Field{}, []*Field{}, []*Field{}}
}

// AddField appends scheme with new field
func (s *Scheme) AddField(name string, ft FieldType, isMandatory bool) {
	newField := &Field{name, ft, len(s.fieldsMap), isMandatory, nil}
	s.fieldsMap[name] = newField
	s.Fields = append(s.Fields, newField)
	if ft == FieldTypeString {
		s.stringFields = append(s.stringFields, newField)
	}
}

// AddNested adds field which value is nested Scheme
func (s *Scheme) AddNested(name string, nested *Scheme, isMandatory bool) {
	newField := &Field{name, FieldTypeNested, len(s.fieldsMap), isMandatory, nested}
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

// CanBeAssigned checks if correct value is provided for the field.
// no such field -> false
// value is nil and field is mandatory -> false
// value type and field type are incompatible -> false
// value has appropriate type (e.g. string for numberic field) and fits into field (e.g. 255 fits into float, double, int, long, byte; 256 does not fit into byte etc) -> true
// Numbers must be float64 only
func (s *Scheme) CanBeAssigned(fieldName string, fieldValue interface{}) bool {
	f, ok := s.fieldsMap[fieldName]
	if !ok {
		return false
	}
	return s.canBeAssigned(f, fieldValue)
}

func (s *Scheme) canBeAssigned(f *Field, fv interface{}) bool {
	if fv == nil {
		return !f.IsMandatory
	}
	switch f.Ft {
	case FieldTypeBool:
		_, ok := fv.(bool)
		return ok
	case FieldTypeString:
		_, ok := fv.(string)
		return ok
	default:
		float64Src, ok := fv.(float64)
		if !ok {
			return false
		}
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
}

// UnmarshalText is used to conform to yaml.TextMarshaler inteface
func (s *Scheme) UnmarshalText(text []byte) error {
	newS, err := YamlToScheme(string(text))
	if err != nil {
		return err
	}
	s.fieldsMap = newS.fieldsMap
	s.Fields = newS.Fields
	s.stringFields = newS.stringFields
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

// YamlToScheme creates Scheme by provided yaml `fieldName: yamlFieldType`
//  Field types:
//    - `int` -> `int32`
//    - `long` -> `int64`
//    - `float` -> `float32`
//    - `double` -> `float64`
//    - `bool` -> `bool`
//    - `string` -> `string`
//    - `byte` -> `byte`
//  See [dynobuffers_test.go](dynobuffers_test.go) for examples
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
			fieldName := mapItem.Key.(string)
			isMandatory := unicode.IsUpper(rune(fieldName[0]))
			if isMandatory {
				fieldName = strings.ToLower(fieldName)
			}
			nestedScheme, err := mapSliceToScheme(nestedMapSlice)
			if err != nil {
				return nil, err
			}
			scheme.AddNested(fieldName, nestedScheme, isMandatory)
		} else if typeStr, ok := mapItem.Value.(string); ok {
			fieldName := mapItem.Key.(string)
			isMandatory := unicode.IsUpper(rune(fieldName[0]))
			if isMandatory {
				fieldName = strings.ToLower(fieldName)
			}
			if ft, ok := yamlFieldTypesMap[typeStr]; ok {
				scheme.AddField(fieldName, ft, isMandatory)
			} else {
				return nil, errors.New("unknown field type: " + typeStr)
			}
		}
	}
	return scheme, nil
}
