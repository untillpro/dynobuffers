/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"errors"

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

var fieldTypesMap = map[string]FieldType{
	"int":    FieldTypeInt,
	"long":   FieldTypeLong,
	"float":  FieldTypeFloat,
	"double": FieldTypeDouble,
	"string": FieldTypeString,
	"bool":   FieldTypeBool,
	"byte":   FieldTypeByte,
}

var fixedSizeFieldsSizesMap = map[FieldType]int{
	FieldTypeInt:    4,
	FieldTypeLong:   8,
	FieldTypeFloat:  4,
	FieldTypeDouble: 8,
	FieldTypeBool:   1,
	FieldTypeByte:   1,
}

type fieldModification struct {
	value interface{}
	isSet bool
}

// Buffer s.e.
type Buffer struct {
	bl           *flatbuffers.Builder
	schema       *Schema
	fields       map[string]interface{}
	stringFields map[string]string
	tab          flatbuffers.Table
}

// NewBuffer s.e.
func NewBuffer(schema *Schema) *Buffer {
	b := &Buffer{}
	b.schema = schema
	return b
}

// ReadInt32 s.e.
func ReadInt32(tab flatbuffers.Table, name string, schema *Schema) int32 {
	order := schema.fieldsOrder[name]
	o := flatbuffers.UOffsetT(tab.Offset(flatbuffers.VOffsetT((order + 2) * 2)))
	if o != 0 {
		return tab.GetInt32(o + tab.Pos)
	}
	return int32(0)
}

// ReadFloat32 s.e.
func ReadFloat32(tab flatbuffers.Table, name string, schema *Schema) float32 {
	order := schema.fieldsOrder[name]
	o := flatbuffers.UOffsetT(tab.Offset(flatbuffers.VOffsetT((order + 2) * 2)))
	if o != 0 {
		return tab.GetFloat32(o + tab.Pos)
	}
	return float32(0)
}

// GetInt32 s.e.
func (b *Buffer) GetInt32(name string) (int32, bool) {
	order := b.schema.fieldsOrder[name]
	o := flatbuffers.UOffsetT(b.tab.Offset(flatbuffers.VOffsetT((order + 2) * 2)))
	if o != 0 {
		return b.tab.GetInt32(o + b.tab.Pos), true
	}
	return int32(0), true
}

// GetFloat32 s.e.
func (b *Buffer) GetFloat32(name string) (float32, bool) {
	order := b.schema.fieldsOrder[name]
	o := flatbuffers.UOffsetT(b.tab.Offset(flatbuffers.VOffsetT((order + 2) * 2)))
	if o != 0 {
		return b.tab.GetFloat32(o + b.tab.Pos), true
	}
	return float32(0), true
}

// Get returns stored field value by name.
// If Set() was called for the field then old value is still returned.
// Value == nil && !isSet -> is not set or no such field in schema.
func (b *Buffer) Get(name string) (value interface{}, isSet bool) {
	order := b.schema.fieldsOrder[name]
	o := flatbuffers.UOffsetT(b.tab.Offset(flatbuffers.VOffsetT((order + 2) * 2)))
	switch b.schema.fieldTypes[name] {
	case FieldTypeInt:
		if o != 0 {
			return b.tab.GetInt32(o + b.tab.Pos), true
		}
		return int32(0), true
	case FieldTypeLong:
		if o != 0 {
			return b.tab.GetInt64(o + b.tab.Pos), true
		}
		return int64(0), true
	case FieldTypeFloat:
		if o != 0 {
			return b.tab.GetFloat32(o + b.tab.Pos), true
		}
		return float32(0), true
	case FieldTypeDouble:
		if o != 0 {
			return b.tab.GetFloat64(o + b.tab.Pos), true
		}
		return float64(0), true
	case FieldTypeByte:
		if o != 0 {
			return b.tab.GetByte(o + b.tab.Pos), true
		}
		return byte(0), true
	case FieldTypeBool:
		if o != 0 {
			return b.tab.GetBool(o + b.tab.Pos), true
		}
		return false, true
	case FieldTypeString:
		if o != 0 {
			return string(b.tab.ByteVector(o + b.tab.Pos)), true
		}
		return "", true
	}
	return nil, false
}

// ReadBuffer creates Buffer from bytes using schema
func ReadBuffer(bytes []byte, schema *Schema) *Buffer {
	b := NewBuffer(schema)
	rootUOffsetT := flatbuffers.GetUOffsetT(bytes)
	b.tab.Bytes = bytes
	b.tab.Pos = rootUOffsetT
	return b
}

// Set sets field value by name.
// Value type must be in [int32, int64, float32, float64, string, bool], error otherwise.
// Byte buffer is not modified
func (b *Buffer) Set(name string, value interface{}) {
	ft := b.schema.fieldTypes[name]
	if ft == FieldTypeString {
		if b.stringFields == nil {
			b.stringFields = map[string]string{}
		}
		b.stringFields[name] = value.(string)
	} else {
		if b.fields == nil {
			b.fields = map[string]interface{}{}
		}
		b.fields[name] = value
	}
}

// Unset field to remove it on ToBytes().
// Note: Get() still returns previous value
func (b *Buffer) Unset(name string) {
	panic("not implemented")
}

// ToBytes returns initial byte array with modifications made by Set().
// Note: current Buffer still keep initial byte array, current modifications are not discarded
func (b *Buffer) ToBytes() []byte {
	bl := flatbuffers.NewBuilder(0)
	stringUOffsetTs := map[string]flatbuffers.UOffsetT{}
	for _, fieldName := range b.schema.fieldsOrderedList {
		if b.schema.fieldTypes[fieldName] == FieldTypeString {
			strToWrite := ""
			if modifiedString, ok := b.stringFields[fieldName]; ok {
				strToWrite = modifiedString
			} else {
				actual, _ := b.Get(fieldName)
				strToWrite = actual.(string)
			}
			stringUOffsetTs[fieldName] = bl.CreateString(strToWrite)
		}
	}
	bl.StartObject(len(b.schema.fieldsOrderedList))
	for i, fieldName := range b.schema.fieldsOrderedList {
		if strUOffsetT, ok := stringUOffsetTs[fieldName]; ok {
			bl.PrependUOffsetTSlot(i, strUOffsetT, 0)
		} else {
			ft := b.schema.fieldTypes[fieldName]
			if value, ok := b.fields[fieldName]; !ok {
				if b.bl != nil {
					// get existing only if the object was read
					value, _ = b.Get(fieldName)
					encodeValue(bl, fieldName, ft, i, value)
				}
			} else {
				encodeValue(bl, fieldName, ft, i, value)
			}
		}
	}
	endUOffsetT := bl.EndObject()
	bl.Finish(endUOffsetT)
	return bl.FinishedBytes()
}

func encodeValue(bl *flatbuffers.Builder, fieldName string, ft FieldType, order int, value interface{}) {
	switch ft {
	case FieldTypeInt:
		bl.PrependInt32Slot(order, value.(int32), 0)
	case FieldTypeLong:
		bl.PrependInt64Slot(order, value.(int64), 0)
	case FieldTypeFloat:
		bl.PrependFloat32Slot(order, value.(float32), 0)
	case FieldTypeDouble:
		bl.PrependFloat64Slot(order, value.(float64), 0)
	case FieldTypeByte:
		bl.PrependByteSlot(order, value.(byte), 0)
	case FieldTypeBool:
		bl.PrependBoolSlot(order, value.(bool), false)
	}
}

// ToJSON s.e.
func (b *Buffer) ToJSON() string {
	// buf := bytes.NewBufferString("")
	// e := json.NewEncoder(buf)
	// buf.WriteString("{")
	// for _, fieldName := range b.schema.fieldsOrderedList {
	// 	if fm, ok := b.modifiedFields[fieldName]; ok {
	// 		if fm.isSet {
	// 			buf.WriteString("\"" + fieldName + "\": ")
	// 			e.Encode(fm.value)
	// 			buf.WriteString(",")
	// 		}
	// 	} else {
	// 		// not modified but had a value
	// 		value, isSet := b.Get(fieldName)
	// 		if isSet {
	// 			buf.WriteString("\"" + fieldName + "\": ")
	// 			e.Encode(value)
	// 			buf.WriteString(",")
	// 		}
	// 	}
	// }
	// if buf.Len() > 1 {
	// 	buf.Truncate(buf.Len() - 1)
	// }
	// buf.WriteString("}")
	// return strings.Replace(buf.String(), "\n", "", -1)
	panic("not implemented")
}

// Schema s.e.
type Schema struct {
	fieldTypes        map[string]FieldType
	fieldsOrder       map[string]int
	fieldsOrderedList []string
}

// NewSchema create new empty schema for manual
func NewSchema() *Schema {
	return &Schema{map[string]FieldType{}, map[string]int{}, []string{}}
}

// AddField appends schema with new field
func (s *Schema) AddField(name string, ft FieldType) {
	s.fieldTypes[name] = ft
	s.fieldsOrder[name] = len(s.fieldsOrderedList)
	s.fieldsOrderedList = append(s.fieldsOrderedList, name)
}

// YamlToSchema creates Schema by provided yaml `fieldName: yamlFieldType`
//
// Field types:
// `int` -> `int32`\r\n
// - `long` -> `int64`
// - `float` -> `float32`
// - `double` -> `float64`
// - `bool` -> `bool`
// - `string` -> `string`
//
// Example:
// `name: string
// price: float
// quantity: int`
func YamlToSchema(yamlStr string) (*Schema, error) {
	schema := NewSchema()
	yamlParsed := yaml.MapSlice{}
	err := yaml.Unmarshal([]byte(yamlStr), &yamlParsed)
	if err != nil {
		return nil, err
	}
	for _, mapItem := range yamlParsed {
		if typeStr, ok := mapItem.Value.(string); ok {
			if ft, ok := fieldTypesMap[typeStr]; ok {
				schema.AddField(mapItem.Key.(string), ft)
			} else {
				return nil, errors.New("unknown field type: " + typeStr)
			}
		}
	}
	return schema, nil
}

func intfToFieldType(intf interface{}) FieldType {
	switch intf.(type) {
	case int32:
		return FieldTypeInt
	case int64:
		return FieldTypeLong
	case float32:
		return FieldTypeFloat
	case float64:
		return FieldTypeDouble
	case string:
		return FieldTypeString
	case bool:
		return FieldTypeBool
	case byte:
		return FieldTypeByte
	}
	return FieldTypeUnspecified
}
