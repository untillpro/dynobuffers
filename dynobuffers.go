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
	schema               *Schema
	modifiedFields       map[string]interface{}
	modifiedStringFields map[string]interface{}
	tab                  flatbuffers.Table
}

type field struct {
	ft    FieldType
	order int
}

// Schema s.e.
type Schema struct {
	fields            map[string]*field
	fieldsOrderedList []string
}

// NewBuffer creates new empty Buffer
func NewBuffer(schema *Schema) *Buffer {
	b := &Buffer{}
	b.schema = schema
	return b
}

// GetInt returns int32 value by name and if the schema contains the field and the value was set
func (b *Buffer) GetInt(name string) (int32, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetInt32(o + b.tab.Pos), true
	}
	return int32(0), false
}

// GetFloat returns float32 value by name and if the schema contains the field and if the value was set
func (b *Buffer) GetFloat(name string) (float32, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetFloat32(o + b.tab.Pos), true
	}
	return float32(0), false
}

// GetString returns string value by name and if the schema contains the field and if the value was set
func (b *Buffer) GetString(name string) (string, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return string(b.tab.ByteVector(o + b.tab.Pos)), true
	}
	return "", false
}

// GetLong returns int64 value by name and if the schema contains the field and if the value was set
func (b *Buffer) GetLong(name string) (int64, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetInt64(o + b.tab.Pos), true
	}
	return int64(0), false
}

// GetDouble returns float64 value by name and if the schema contains the field and if the value was set
func (b *Buffer) GetDouble(name string) (float64, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetFloat64(o + b.tab.Pos), true
	}
	return float64(0), false
}

// GetByte returns byte value by name and if the schema contains the field and if the value was set
func (b *Buffer) GetByte(name string) (byte, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetByte(o + b.tab.Pos), true
	}
	return byte(0), false
}

// GetBool returns bool value by name and if the schema contains the field and if the value was set
func (b *Buffer) GetBool(name string) (bool, bool) {
	o := b.getFieldUOffsetT(name)
	if o != 0 {
		return b.tab.GetBool(o + b.tab.Pos), true
	}
	return false, false
}
 
func (b *Buffer) getFieldUOffsetT(name string) flatbuffers.UOffsetT {
	if f, ok := b.schema.fields[name]; ok {
		return b.getFieldUOffsetTByOrder(f.order)
	}
	return 0
}

func (b *Buffer) getFieldUOffsetTByOrder(order int) flatbuffers.UOffsetT {
	return flatbuffers.UOffsetT(b.tab.Offset(flatbuffers.VOffsetT((order + 2) * 2)))
}

// Get returns stored field value by name.
// nil -> field is unset or no such field in the schema
func (b *Buffer) Get(name string) interface{} {
	if len(b.tab.Bytes) == 0 {
		return nil
	}
	f, ok := b.schema.fields[name]
	if !ok {
		return nil
	}
	o := b.getFieldUOffsetTByOrder(f.order)
	switch f.ft {
	case FieldTypeInt:
		if o != 0 {
			return b.tab.GetInt32(o + b.tab.Pos)
		}
		return nil
	case FieldTypeLong:
		if o != 0 {
			return b.tab.GetInt64(o + b.tab.Pos)
		}
		return nil
	case FieldTypeFloat:
		if o != 0 {
			return b.tab.GetFloat32(o + b.tab.Pos)
		}
		return nil
	case FieldTypeDouble:
		if o != 0 {
			return b.tab.GetFloat64(o + b.tab.Pos)
		}
		return nil
	case FieldTypeByte:
		if o != 0 {
			return b.tab.GetByte(o + b.tab.Pos)
		}
		return nil
	case FieldTypeBool:
		if o != 0 {
			return b.tab.GetBool(o + b.tab.Pos)
		}
		return nil
	default:
		if o != 0 {
			return string(b.tab.ByteVector(o + b.tab.Pos))
		}
		return nil
	}
}

// ReadBuffer creates Buffer from bytes using provided Schema
func ReadBuffer(bytes []byte, schema *Schema) *Buffer {
	b := NewBuffer(schema)
	rootUOffsetT := flatbuffers.GetUOffsetT(bytes)
	b.tab.Bytes = bytes
	b.tab.Pos = rootUOffsetT
	return b
}

// Set sets field value by name.
// Byte array is not modified.
// Call ToBytes() to get modified byte array
func (b *Buffer) Set(name string, value interface{}) {
	f, ok := b.schema.fields[name]
	if !ok {
		return
	}
	if f.ft == FieldTypeString {
		if b.modifiedStringFields == nil {
			b.modifiedStringFields = map[string]interface{}{}
		}
		b.modifiedStringFields[name] = value
	} else {
		if b.modifiedFields == nil {
			b.modifiedFields = map[string]interface{}{}
		}
		b.modifiedFields[name] = value
	}
}

// ToBytes returns byte array in FlatBuffer format with field modified by Set() and fields which initially had values
// Note: initial byte array is kept, current modifications are not discarded
func (b *Buffer) ToBytes() []byte {
	bl := flatbuffers.NewBuilder(0)
	stringUOffsetTs := map[string]flatbuffers.UOffsetT{}
	for _, fieldName := range b.schema.fieldsOrderedList {
		if b.schema.fields[fieldName].ft == FieldTypeString {
			var strToWrite interface{}
			if modifiedString, ok := b.modifiedStringFields[fieldName]; ok {
				strToWrite = modifiedString
			} else {
				strToWrite = b.Get(fieldName)
			}
			if strToWrite != nil {
				stringUOffsetTs[fieldName] = bl.CreateString(strToWrite.(string))
			}
		}
	}
	bl.StartObject(len(b.schema.fields))
	for i, fieldName := range b.schema.fieldsOrderedList {
		if strUOffsetT, ok := stringUOffsetTs[fieldName]; ok {
			bl.PrependUOffsetTSlot(i, strUOffsetT, 0)
		} else {
			ft := b.schema.fields[fieldName].ft
			if value, ok := b.modifiedFields[fieldName]; !ok {
				// get existing only if the object was read
				value = b.Get(fieldName)
				if value != nil {
					encodeValue(bl, fieldName, ft, i, value)
				}
			} else {
				if value != nil {
					encodeValue(bl, fieldName, ft, i, value)
				}
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
		bl.PrependInt32(value.(int32))
	case FieldTypeLong:
		bl.PrependInt64(value.(int64))
	case FieldTypeFloat:
		bl.PrependFloat32(value.(float32))
	case FieldTypeDouble:
		bl.PrependFloat64(value.(float64))
	case FieldTypeByte:
		bl.PrependByte(value.(byte))
	case FieldTypeBool:
		bl.PrependBool(value.(bool))
	}
	bl.Slot(order)
}

// ToJSON returns JSON flat key->value string
func (b *Buffer) ToJSON() string {
	buf := bytes.NewBufferString("")
	e := json.NewEncoder(buf)
	buf.WriteString("{")
	for _, fieldName := range b.schema.fieldsOrderedList {
		var value interface{}
		if strValue, ok := b.modifiedStringFields[fieldName]; ok {
			value = strValue
		} else if nonStrValue, ok := b.modifiedFields[fieldName]; ok {
			value = nonStrValue
		} else {
			value = b.Get(fieldName)
		}
		if value != nil {
			buf.WriteString("\"" + fieldName + "\": ")
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

// NewSchema creates new empty schema
func NewSchema() *Schema {
	return &Schema{map[string]*field{}, []string{}}
}

// AddField appends schema with new field
func (s *Schema) AddField(name string, ft FieldType) {
	s.fields[name] = &field{ft, len(s.fieldsOrderedList)}
	s.fieldsOrderedList = append(s.fieldsOrderedList, name)
}

// HasField returns if the Schema contains the specified field
func (s *Schema) HasField(name string) bool {
	_, ok := s.fields[name]
	return ok
}

/*
YamlToSchema creates Schema by provided yaml `fieldName: yamlFieldType`

Field types:
  - `int` -> `int32`
  - `long` -> `int64`
  - `float` -> `float32`
  - `double` -> `float64`
  - `bool` -> `bool`
  - `string` -> `string`
  - `byte` -> `byte`

See [dynobuffers_test.go](dynobuffers_test.go) for examples
*/
func YamlToSchema(yamlStr string) (*Schema, error) {
	schema := NewSchema()
	yamlParsed := yaml.MapSlice{}
	err := yaml.Unmarshal([]byte(yamlStr), &yamlParsed)
	if err != nil {
		return nil, err
	}
	for _, mapItem := range yamlParsed {
		if typeStr, ok := mapItem.Value.(string); ok {
			if ft, ok := yamlFieldTypesMap[typeStr]; ok {
				schema.AddField(mapItem.Key.(string), ft)
			} else {
				return nil, errors.New("unknown field type: " + typeStr)
			}
		}
	}
	return schema, nil
}
