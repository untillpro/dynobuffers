/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"encoding/binary"
	"errors"
	"math"

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
	bytes                  []byte
	storedFieldsInfo       []byte
	modifiedFields         map[string]*fieldModification
	fixedSizeValuesOffsets map[string]int
	varSizeValuesOffsets   map[string]int
	schema                 *Schema
}

// NewBuffer s.e.
func NewBuffer(schema *Schema) *Buffer {
	b := &Buffer{}
	b.modifiedFields = map[string]*fieldModification{}
	b.schema = schema
	b.fixedSizeValuesOffsets = map[string]int{}
	b.varSizeValuesOffsets = map[string]int{}
	return b
}

// Get returns stored field value by name.
// If Set() was called for the field then old value is still returned.
// Value == nil && !isSet -> is not set or no such field in schema.
func (b *Buffer) Get(name string) (value interface{}, isSet bool) {
	fieldOrder, ok := b.schema.fieldsOrder[name]
	if !ok {
		return nil, false
	}
	bitNum := uint(fieldOrder-int(fieldOrder/8)*8) * 2
	if hasBit(b.storedFieldsInfo, bitNum) {
		if hasBit(b.storedFieldsInfo, bitNum+1) {
			// field is set to nil
			return nil, true
		}
	} else {
		// field is unset or no field in the Schema
		return nil, false
	}
	if offset, ok := b.fixedSizeValuesOffsets[name]; ok {
		// field is fixed-size
		switch b.schema.fieldTypes[name] {
		case FieldTypeInt:
			return int32(binary.LittleEndian.Uint32(b.bytes[offset : offset+4])), true
		case FieldTypeLong:
			return int64(binary.LittleEndian.Uint64(b.bytes[offset : offset+8])), true
		case FieldTypeFloat:
			bits := binary.LittleEndian.Uint32(b.bytes[offset : offset+4])
			value = math.Float32frombits(bits)
			return value, true
		case FieldTypeDouble:
			bits := binary.LittleEndian.Uint64(b.bytes[offset : offset+8])
			value = math.Float64frombits(bits)
			return value, true
		case FieldTypeBool:
			return b.bytes[offset] != 0, true
		case FieldTypeByte:
			return b.bytes[offset], true
		}
	}
	// field is string
	varSizeValueOffset := b.varSizeValuesOffsets[name]
	offset := int32(binary.LittleEndian.Uint32(b.bytes[varSizeValueOffset : varSizeValueOffset+4]))
	size := int32(binary.LittleEndian.Uint32(b.bytes[varSizeValueOffset+4 : varSizeValueOffset+8]))
	return string(b.bytes[offset : offset+size]), true
}

// ReadBuffer creates Buffer from bytes using schema
func ReadBuffer(bytes []byte, schema *Schema) *Buffer {
	b := NewBuffer(schema)
	b.bytes = bytes

	varSizeValuesOffsetsPos := int(binary.LittleEndian.Uint32(bytes[:4]))
	fixedSizeValuesPos := int(binary.LittleEndian.Uint32(bytes[4:8]))
	b.storedFieldsInfo = bytes[8:varSizeValuesOffsetsPos]

	for _, fieldName := range schema.fieldsOrderedList {
		ft := schema.fieldTypes[fieldName]
		if fixedFieldSize, ok := fixedSizeFieldsSizesMap[ft]; ok {
			// fixed-size
			b.fixedSizeValuesOffsets[fieldName] = fixedSizeValuesPos
			fixedSizeValuesPos += fixedFieldSize
		} else {
			b.varSizeValuesOffsets[fieldName] = varSizeValuesOffsetsPos
			varSizeValuesOffsetsPos += 8
		}
	}

	return b
}

// Set sets field value by name.
// Value type must be in [int32, int64, float32, float64, string, bool], error otherwise.
// Byte buffer is not modified
func (b *Buffer) Set(name string, value interface{}) error {
	if value != nil {
		ft:=intfToFieldType(value)
		if ft == FieldTypeUnspecified {
			return errors.New("value is of unsupported type")
		}
		if schemaFieldType, ok := b.schema.fieldTypes[name]; ok {
			if schemaFieldType != ft {
				return errors.New("value type differs from field type")
			} 
		} else {
			return errors.New("no such field in the schema")
		}
	}
	
	b.modifiedFields[name] = &fieldModification{value, true}
	return nil
}

// Unset field to remove it on ToBytes().
// Note: Get() still returns previous value
func (b *Buffer) Unset(name string) {
	b.modifiedFields[name] = &fieldModification{nil, false}
}

// ToBytes returns initial byte array with modifications made by Set().
// Note: current Buffer still keep initial byte array, current modifications are not discarded
func (b *Buffer) ToBytes() []byte {
	storedFieldsInfo := make([]byte, int(len(b.schema.fieldsOrderedList)*2/8)+1)
	copy(storedFieldsInfo, b.storedFieldsInfo)
	fixedSizeValues := []byte{}
	varSizeValues := []byte{}
	varSizeValuesOffsets := []byte{}
	varSizeValuesLengths := []int{}
	bitNum := uint(0)
	for _, fieldName := range b.schema.fieldsOrderedList {
		ft := b.schema.fieldTypes[fieldName]
		if fm, ok := b.modifiedFields[fieldName]; ok {
			if fm.isSet {
				storedFieldsInfo = setBit(storedFieldsInfo, bitNum)
				if fm.value == nil {
					storedFieldsInfo = setBit(storedFieldsInfo, bitNum+1)
				} else {
					if _, ok := fixedSizeFieldsSizesMap[ft]; ok {
						fixedSizeValues = append(fixedSizeValues, encodeIntf(fm.value, ft)...)
					} else {
						// var-size value is modified
						mIntfBytes := encodeIntf(fm.value, ft)
						varSizeValues = append(varSizeValues, mIntfBytes...)
						varSizeValuesLengths = append(varSizeValuesLengths, len(mIntfBytes))
					}
				}
			} else {
				storedFieldsInfo = clearBit(storedFieldsInfo, bitNum)
				storedFieldsInfo = clearBit(storedFieldsInfo, bitNum+1)
			}
		} else if hasBit(b.storedFieldsInfo, bitNum) {
			// not modified but had a value
			storedFieldsInfo = setBit(storedFieldsInfo, bitNum)
			if hasBit(b.storedFieldsInfo, bitNum+1) {
				// was set to null
				storedFieldsInfo = setBit(storedFieldsInfo, bitNum+1)
			} else {
				// had value
				if fieldSize, ok := fixedSizeFieldsSizesMap[ft]; ok {
					offset := b.fixedSizeValuesOffsets[fieldName]
					fixedSizeValues = append(fixedSizeValues, b.bytes[offset:offset+fieldSize]...)
				} else {
					varSizeValueOffset := b.varSizeValuesOffsets[fieldName]
					offset := int32(binary.LittleEndian.Uint32(b.bytes[varSizeValueOffset : varSizeValueOffset+4]))
					size := int32(binary.LittleEndian.Uint32(b.bytes[varSizeValueOffset+4 : varSizeValueOffset+8]))
					varSizeValues = append(varSizeValues, b.bytes[offset:offset+size]...)
					varSizeValuesLengths = append(varSizeValuesLengths, int(size))
				}
			}
		} 
		bitNum += 2
	}

	res := make([]byte, 8)
	res = append(res, storedFieldsInfo...)

	varSizeValuesOffsetsLen := len(varSizeValuesLengths) * 4 * 2
	varSaizeValueOffset := len(res) + len(fixedSizeValues) + varSizeValuesOffsetsLen
	for _, varSizeValueLen := range varSizeValuesLengths {
		tmp := make([]byte, 4)
		binary.LittleEndian.PutUint32(tmp, uint32(varSaizeValueOffset))
		varSizeValuesOffsets = append(varSizeValuesOffsets, tmp...)
		binary.LittleEndian.PutUint32(tmp, uint32(varSizeValueLen))
		varSizeValuesOffsets = append(varSizeValuesOffsets, tmp...)
		varSaizeValueOffset += varSizeValueLen
	}

	res = append(res, varSizeValuesOffsets...)
	res = append(res, fixedSizeValues...)
	res = append(res, varSizeValues...)
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(len(storedFieldsInfo)+8))
	copy(res[:4], tmp)
	binary.LittleEndian.PutUint32(tmp, uint32(len(storedFieldsInfo)+8+len(varSizeValuesOffsets)))
	copy(res[4:8], tmp)
	return res
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

func encodeIntf(intf interface{}, ft FieldType) []byte {
	switch ft {
	case FieldTypeInt:
		res := make([]byte, 4)
		binary.LittleEndian.PutUint32(res, uint32(intf.(int32)))
		return res
	case FieldTypeLong:
		res := make([]byte, 8)
		binary.LittleEndian.PutUint64(res, uint64(intf.(int64)))
		return res
	case FieldTypeFloat:
		bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(bytes, math.Float32bits(intf.(float32)))
		return bytes
	case FieldTypeDouble:
		bytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(bytes, math.Float64bits(intf.(float64)))
		return bytes
	case FieldTypeString:
		str := intf.(string)
		return []byte(str)
	case FieldTypeBool:
		if intf.(bool) {
			return []byte{1}
		}
		return []byte{0}
	default: // FieldTypeByte
		return []byte{intf.(byte)}
	}
}

func setBit(bytes []byte, pos uint) []byte {
	byteNum := int(pos / 8)
	if byteNum >= len(bytes) {
		tmp := make([]byte, byteNum-len(bytes)+1)
		bytes = append(bytes, tmp...)
	}
	bytes[byteNum] |= (1 << uint(pos-uint(byteNum*8)))
	return bytes
}

func hasBit(bytes []byte, pos uint) bool {
	byteNum := int(pos / 8)
	if byteNum >= len(bytes) {
		return false
	}
	val := bytes[byteNum] & (1 << uint(pos-uint(byteNum*8)))
	return val > 0
}

func clearBit(bytes []byte, pos uint) []byte {
	byteNum := int(pos / 8)
	if byteNum >= len(bytes) {
		tmp := make([]byte, byteNum-len(bytes)+1)
		bytes = append(bytes, tmp...)
	}
	mask := ^(1 << pos)
	tmp := int(bytes[byteNum])
	tmp &= mask
	bytes[byteNum] = byte(tmp)
	return bytes
}
