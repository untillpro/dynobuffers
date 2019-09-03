/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"
	"strings"

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
	bytes       []byte
	storageMask []byte
	fields      map[string]*fieldInfo
	schema      *Schema
}

type fieldInfo struct {
	order         int
	offset        int
	ft            FieldType
	isFixedSize   bool
	isModified    bool
	modifiedValue interface{}
	isSet         bool
	isNil         bool
	fixedSize     int
	name          string
}

// NewBuffer s.e.
func NewBuffer(schema *Schema) *Buffer {
	b := &Buffer{}
	b.schema = schema
	b.fields = map[string]*fieldInfo{}
	return b
}

// Get returns stored field value by name.
// If Set() was called for the field then old value is still returned.
// Value == nil && !isSet -> is not set or no such field in schema.
func (b *Buffer) Get(name string) (value interface{}, isSet bool) {
	fi, ok := b.fields[name]
	if !ok {
		return nil, false
	}
	bitNum := fi.order * 2
	if hasBit(b.storageMask, bitNum+1) {
		// field is set to nil
		return nil, true
	}
	if fi.isFixedSize {
		// field is fixed-size
		switch fi.ft {
		case FieldTypeInt:
			return int32(binary.LittleEndian.Uint32(b.bytes[fi.offset : fi.offset+4])), true
		case FieldTypeLong:
			return int64(binary.LittleEndian.Uint64(b.bytes[fi.offset : fi.offset+8])), true
		case FieldTypeFloat:
			bits := binary.LittleEndian.Uint32(b.bytes[fi.offset : fi.offset+4])
			value = math.Float32frombits(bits)
			return value, true
		case FieldTypeDouble:
			bits := binary.LittleEndian.Uint64(b.bytes[fi.offset : fi.offset+8])
			value = math.Float64frombits(bits)
			return value, true
		case FieldTypeBool:
			return b.bytes[fi.offset] != 0, true
		case FieldTypeByte:
			return b.bytes[fi.offset], true
		}
	} else {
		// field is string
		offset := int32(binary.LittleEndian.Uint32(b.bytes[fi.offset : fi.offset+4]))
		size := int32(binary.LittleEndian.Uint32(b.bytes[fi.offset+4 : fi.offset+8]))
		return string(b.bytes[offset : offset+size]), true
	}
	return nil, false
}

// ReadBuffer creates Buffer from bytes using schema
func ReadBuffer(bytes []byte, schema *Schema) *Buffer {
	b := NewBuffer(schema)
	b.bytes = bytes

	varSizeValuesOffsetsPos := int(binary.LittleEndian.Uint32(bytes[:4]))
	fixedSizeValuesPos := int(binary.LittleEndian.Uint32(bytes[4:8]))
	b.storageMask = bytes[8:varSizeValuesOffsetsPos]

	for i, fiSchema := range schema.fieldsOrderedList {
		if !hasBit(b.storageMask, i*2) {
			continue
		}
		fi := &fieldInfo{}
		fi.fixedSize = fiSchema.fixedSize
		fi.ft = fiSchema.ft
		fi.isFixedSize = fiSchema.isFixedSize
		fi.name = fiSchema.name
		fi.order = fiSchema.order
		fi.isSet = true
		fi.isNil = hasBit(b.storageMask, i*2+1)

		if fi.isFixedSize {
			// fixed-size
			if !fi.isNil {
				fi.offset = fixedSizeValuesPos
				fixedSizeValuesPos += fi.fixedSize
			}
		} else {
			if !fi.isNil {
				fi.offset = varSizeValuesOffsetsPos
				varSizeValuesOffsetsPos += 8
			}
		}
		b.fields[fi.name] = fi
	}

	return b
}

// Set sets field value by name.
// Value type must be in [int32, int64, float32, float64, string, bool], error otherwise.
// Byte buffer is not modified
func (b *Buffer) Set(name string, value interface{}) error {
	var schemaFieldType FieldType
	if value != nil {
		ft := intfToFieldType(value)
		if sft, ok := b.schema.fieldTypes[name]; ok {
			schemaFieldType = sft
			if schemaFieldType != ft {
				return errors.New("value type differs from field type")
			}
		} else {
			return errors.New("no such field in the schema")
		}
	}
	fi, ok := b.fields[name]
	if !ok {
		fi = &fieldInfo{}
		fi.ft = schemaFieldType
		if fixedFieldSize, ok := fixedSizeFieldsSizesMap[schemaFieldType]; ok {
			fi.fixedSize = fixedFieldSize
			fi.isFixedSize = true
		} else {
			fi.isFixedSize = false
		}
		fi.isSet = true
		fi.isNil = false
		b.fields[name] = fi
	}
	fi.isModified = true
	fi.modifiedValue = value
	return nil
}

// Unset field to remove it on ToBytes().
// Note: Get() still returns previous value
func (b *Buffer) Unset(name string) {
	fi, ok := b.fields[name]
	if !ok {
		fi = &fieldInfo{}
		b.fields[name] = fi
	}
	fi.isModified = true
	fi.isSet = false
	fi.modifiedValue = nil
}

// ToBytes returns initial byte array with modifications made by Set().
// Note: current Buffer still keep initial byte array, current modifications are not discarded
func (b *Buffer) ToBytes() []byte {
	storedFieldsInfo := make([]byte, int(len(b.schema.fieldsOrderedList)*2/8)+1)
	copy(storedFieldsInfo, b.storageMask)
	fixedSizeValues := []byte{}
	varSizeValues := []byte{}
	varSizeValuesOffsets := []byte{}
	varSizeValuesLengths := []int{}
	bitNum := 0
	for _, fiOrdered := range b.schema.fieldsOrderedList {
		fi, ok := b.fields[fiOrdered.name]
		if !ok {
			break
		}
		if fi.isModified {
			if fi.isSet {
				storedFieldsInfo = setBit(storedFieldsInfo, bitNum)
				if fi.modifiedValue == nil {
					storedFieldsInfo = setBit(storedFieldsInfo, bitNum+1)
				} else {
					if fi.isFixedSize {
						fixedSizeValues = append(fixedSizeValues, encodeIntf(fi.modifiedValue, fi.ft)...)
					} else {
						// var-size value is modified
						mIntfBytes := encodeIntf(fi.modifiedValue, fi.ft)
						varSizeValues = append(varSizeValues, mIntfBytes...)
						varSizeValuesLengths = append(varSizeValuesLengths, len(mIntfBytes))
					}
				}
			} else {
				storedFieldsInfo = clearBit(storedFieldsInfo, bitNum)
				storedFieldsInfo = clearBit(storedFieldsInfo, bitNum+1)
			}
		} else if fi.isSet {
			// not modified but had a value
			storedFieldsInfo = setBit(storedFieldsInfo, bitNum)
			if fi.isNil {
				// was set to null
				storedFieldsInfo = setBit(storedFieldsInfo, bitNum+1)
			} else {
				// had value
				if fi.isFixedSize {
					fixedSizeValues = append(fixedSizeValues, b.bytes[fi.offset:fi.offset+fi.fixedSize]...)
				} else {
					varSizeValueOffset := fi.offset
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

// ToJSON s.e.
func (b *Buffer) ToJSON() string {
	buf := bytes.NewBufferString("")
	e := json.NewEncoder(buf)
	buf.WriteString("{")
	for fieldName, fi := range b.fields {
		if fi.isModified {
			if fi.isSet {
				buf.WriteString("\"" + fieldName + "\": ")
				e.Encode(fi.modifiedValue)
				buf.WriteString(",")
			}
		} else {
			// not modified but had a value
			value, isSet := b.Get(fieldName)
			if isSet {
				buf.WriteString("\"" + fieldName + "\": ")
				e.Encode(value)
				buf.WriteString(",")
			}
		}
	}
	if buf.Len() > 1 {
		buf.Truncate(buf.Len() - 1)
	}
	buf.WriteString("}")
	return strings.Replace(buf.String(), "\n", "", -1)
}

// Schema s.e.
type Schema struct {
	fieldTypes        map[string]FieldType
	fieldsOrder       map[string]int
	fieldsOrderedList []*fieldInfo
}

// NewSchema create new empty schema for manual
func NewSchema() *Schema {
	return &Schema{map[string]FieldType{}, map[string]int{}, []*fieldInfo{}}
}

// AddField appends schema with new field
func (s *Schema) AddField(name string, ft FieldType) {
	s.fieldTypes[name] = ft
	s.fieldsOrder[name] = len(s.fieldsOrderedList)
	fi := &fieldInfo{}

	if fixedSize, ok := fixedSizeFieldsSizesMap[ft]; ok {
		fi.fixedSize = fixedSize
		fi.isFixedSize = true
	} else {
		fi.isFixedSize = false
	}
	fi.order = len(s.fieldsOrderedList)
	fi.name = name
	fi.ft = ft
	s.fieldsOrderedList = append(s.fieldsOrderedList, fi)
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

func setBit(bytes []byte, pos int) []byte {
	byteNum := int(pos / 8)
	if byteNum >= len(bytes) {
		tmp := make([]byte, byteNum-len(bytes)+1)
		bytes = append(bytes, tmp...)
	}
	bytes[byteNum] |= (1 << uint(pos-byteNum*8))
	return bytes
}

func hasBit(bytes []byte, pos int) bool {
	byteNum := int(pos / 8)
	if byteNum >= len(bytes) {
		return false
	}
	val := bytes[byteNum] & (1 << uint(pos-byteNum*8))
	return val > 0
}

func clearBit(bytes []byte, pos int) []byte {
	byteNum := int(pos / 8)
	if byteNum >= len(bytes) {
		tmp := make([]byte, byteNum-len(bytes)+1)
		bytes = append(bytes, tmp...)
	}
	mask := ^(1 << uint(pos))
	tmp := int(bytes[byteNum])
	tmp &= mask
	bytes[byteNum] = byte(tmp)
	return bytes
}
