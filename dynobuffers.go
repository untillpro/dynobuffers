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

	flatbuffers "github.com/google/flatbuffers/go"
	"gopkg.in/yaml.v2"
)

// FieldType s.e.
type FieldType int

const (
	// FieldTypeUnspecified - wrong type
	FieldTypeUnspecified FieldType = iota
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
		return b.tab.GetInt32(o + b.tab.Pos), true
	}
	return int32(0), false
}

// GetFloat returns float32 value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetFloat(name string) (float32, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetFloat32(o + b.tab.Pos), true
	}
	return float32(0), false
}

// GetString returns string value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetString(name string) (string, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return string(b.tab.ByteVector(o + b.tab.Pos)), true
	}
	return "", false
}

// GetLong returns int64 value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetLong(name string) (int64, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetInt64(o + b.tab.Pos), true
	}
	return int64(0), false
}

// GetDouble returns float64 value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetDouble(name string) (float64, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetFloat64(o + b.tab.Pos), true
	}
	return float64(0), false
}

// GetByte returns byte value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetByte(name string) (byte, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetByte(o + b.tab.Pos), true
	}
	return byte(0), false
}

// GetBool returns bool value by name and if the scheme contains the field and if the value was set to non-nil
func (b *Buffer) GetBool(name string) (bool, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetBool(o + b.tab.Pos), true
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
	return flatbuffers.UOffsetT(b.tab.Offset(flatbuffers.VOffsetT((order + 2) * 2)))
}

func (b *Buffer) getByStringField(f *Field) (string, bool) {
	o := b.getFieldUOffsetTByOrder(f.order)
	if o == 0 {
		return "", false
	}
	return string(b.tab.ByteVector(o + b.tab.Pos)), true
}

func (b *Buffer) getByField(f *Field) interface{} {
	o := b.getFieldUOffsetTByOrder(f.order)
	if o == 0 {
		return nil
	}
	uOffsetT := o + b.tab.Pos
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
	if len(b.modifiedFields) == 0 {
		b.modifiedFields = make([]*modifiedField, len(b.scheme.Fields))
	}
	b.modifiedFields[f.order] = &modifiedField{Field{name, f.Ft, f.order, f.IsMandatory}, value, 0}
}

// ToBytes returns new FlatBuffer byte array with fields modified by Set() and fields which initially had values
// Note: initial byte array and current modifications are kept
func (b *Buffer) ToBytes() ([]byte, error) {
	bl := flatbuffers.NewBuilder(0)

	strUOffsetTs := make([]flatbuffers.UOffsetT, len(b.scheme.Fields))
	if len(b.modifiedFields) == 0 {
		b.modifiedFields = make([]*modifiedField, len(b.scheme.Fields))
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
		isSet  := false
		if f.Ft == FieldTypeString {
			if strUOffsetTs[f.order] > 0 {
				bl.PrependUOffsetTSlot(f.order, strUOffsetTs[f.order], 0)
				isSet = true
			}
		} else {
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
			return nil, fmt.Errorf("Field %s is mandatory but not set", f.Name)
		}
	}
	endUOffsetT := bl.EndObject()
	bl.Finish(endUOffsetT)
	return bl.FinishedBytes(), nil
}

func copyNonStringField(dest *flatbuffers.Builder, src *Buffer, f *Field) bool {
	offset := src.getFieldUOffsetTByOrder(f.order)
	if offset == 0 {
		return false
	}
	offset += src.tab.Pos
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
			buf.WriteString("\"" + f.Name + "\": ")
			e.Encode(value)
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
	newField := &Field{name, ft, len(s.fieldsMap), isMandatory}
	s.fieldsMap[name] = newField
	s.Fields = append(s.Fields, newField)
	if ft == FieldTypeString {
		s.stringFields = append(s.stringFields, newField)
	}
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

// CanBeAssigned checks if correct value is provided for the field. Returns if scheme contains such field and its type equal to the value type
// Numbers must be float64 only
func (s *Scheme) CanBeAssigned(fieldName string, fieldValue interface{}) bool {
	f, ok := s.fieldsMap[fieldName]
	if !ok {
		return false
	}
	switch f.Ft {
	case FieldTypeBool:
		_, ok := fieldValue.(bool)
		return ok
	case FieldTypeString:
		_, ok := fieldValue.(string)
		return ok
	default:
		float64Src, ok := fieldValue.(float64)
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
	buf := bytes.NewBufferString("")
	for _, f := range s.Fields {
		for ftStr, curFt := range yamlFieldTypesMap {
			if curFt == f.Ft {
				buf.WriteString(f.Name + ": " + ftStr + "\n")
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
	scheme := NewScheme()
	yamlParsed := yaml.MapSlice{}
	err := yaml.Unmarshal([]byte(yamlStr), &yamlParsed)
	if err != nil {
		return nil, err
	}
	for _, mapItem := range yamlParsed {
		if typeStr, ok := mapItem.Value.(string); ok {
			isMandatory := typeStr[0:1] == "~"
			if isMandatory {
				typeStr = typeStr[1:]
			}
			if ft, ok := yamlFieldTypesMap[typeStr]; ok {
				scheme.AddField(mapItem.Key.(string), ft, isMandatory)
			} else {
				return nil, errors.New("unknown field type: " + typeStr)
			}
		}
	}
	return scheme, nil
}
