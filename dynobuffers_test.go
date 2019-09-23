/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var schemaStr = `
name: string
price: float
quantity: int
`

var schemaStrNew = `
name: string
price: float
quantity: int
newField: long
`

var allTypesYaml = `
int: int
long: long
float: float
double: double
string: string
boolTrue: bool
boolFalse: bool
byte: byte
`

func TestBasicUsage(t *testing.T) {
	s, err := YamlToSchema(schemaStrNew)
	if err != nil {
		t.Fatal(err)
	}

	// create new from sratch
	b := NewBuffer(s)
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	b.Set("unknownField", "") // no errors, nothing will be me on ToBytes()
	bytes := b.ToBytes()

	// create from bytes
	b = ReadBuffer(bytes, s)
	actual := b.Get("name")
	assert.Equal(t, "cola", actual.(string))
	actual = b.Get("price")
	assert.Equal(t, float32(0.123), actual.(float32))
	actual = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))
	actual = b.Get("newField")

	// modify existing
	b.Set("price", float32(0.124))
	b.Set("name", nil) // set to nil is equivalent to unset
	bytes = b.ToBytes()
	b = ReadBuffer(bytes, s)
	actual = b.Get("name")
	assert.Nil(t, actual)
	actual = b.Get("price")
	assert.Equal(t, float32(0.124), actual.(float32))
	actual = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))

	// unset or unknown field -> nil
	actual = b.Get("unknownField")
	assert.Nil(t, actual)
	_, ok := b.GetInt("unknownField")
	assert.False(t, ok)
	// field was not set -> nil
	actual = b.Get("newField")
	assert.Nil(t, actual)
}

func TestNilFields(t *testing.T) {
	s, err := YamlToSchema(allTypesYaml)
	if err != nil {
		t.Fatal(err)
	}

	// test initially unset
	b := NewBuffer(s)
	testNilFields(t, b)
	bytes := b.ToBytes()
	b = ReadBuffer(bytes, s)
	testNilFields(t, b)

	//test initially set to nil
	b.Set("int", nil)
	b.Set("long", nil)
	b.Set("float", nil)
	b.Set("double", nil)
	b.Set("string", nil)
	b.Set("boolFalse", nil)
	b.Set("boolTrue", nil)
	b.Set("byte", nil)
	bytes = b.ToBytes()
	b = ReadBuffer(bytes, s)
	testNilFields(t, b)

	// test unset
	b.Set("int", int32(1))
	b.Set("long", int64(2))
	b.Set("float", float32(3))
	b.Set("double", float64(4))
	b.Set("string", "str")
	b.Set("boolFalse", false)
	b.Set("boolTrue", true)
	b.Set("byte", byte(5))
	bytes = b.ToBytes()
	b = ReadBuffer(bytes, s)
	b.Set("int", nil)
	b.Set("long", nil)
	b.Set("float", nil)
	b.Set("double", nil)
	b.Set("string", nil)
	b.Set("boolFalse", nil)
	b.Set("boolTrue", nil)
	b.Set("byte", nil)
	bytes = b.ToBytes()
	b = ReadBuffer(bytes, s)
	testNilFields(t, b)
}

func testNilFields(t *testing.T, b *Buffer) {
	assert.Nil(t, b.Get("int"))
	_, ok := b.GetInt("int")
	assert.False(t, ok)
	assert.Nil(t, b.Get("long"))
	_, ok = b.GetLong("long")
	assert.False(t, ok)
	assert.Nil(t, b.Get("float"))
	_, ok = b.GetFloat("float")
	assert.False(t, ok)
	assert.Nil(t, b.Get("double"))
	_, ok = b.GetDouble("double")
	assert.False(t, ok)
	assert.Nil(t, b.Get("string"))
	_, ok = b.GetString("string")
	assert.False(t, ok)
	assert.Nil(t, b.Get("boolFalse"))
	_, ok = b.GetBool("boolFalse")
	assert.False(t, ok)
	assert.Nil(t, b.Get("boolTrue"))
	_, ok = b.GetBool("boolTrue")
	assert.False(t, ok)
	assert.Nil(t, b.Get("byte"))
	_, ok = b.GetByte("byte")
	assert.False(t, ok)
}

func TestWriteNewReadOld(t *testing.T) {
	schemaNew, err := YamlToSchema(schemaStrNew)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(schemaNew)
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	b.Set("newField", int64(1))
	bytesNew := b.ToBytes()

	schemaOld, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytesNew, schemaOld)

	actual := b.Get("name")
	assert.Equal(t, "cola", actual.(string))
	actual = b.Get("price")
	assert.Equal(t, float32(0.123), actual.(float32))
	actual = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))

	actual = b.Get("newField")
	assert.Nil(t, actual)
}

func TestWriteOldReadNew(t *testing.T) {
	schemaOld, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(schemaOld)
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	bytesOld := b.ToBytes()

	schemaNew, err := YamlToSchema(schemaStrNew)
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytesOld, schemaNew)

	actual := b.Get("name")
	assert.Equal(t, "cola", actual.(string))
	actual = b.Get("price")
	assert.Equal(t, float32(0.123), actual.(float32))
	actual = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))

	actual = b.Get("newField")
	assert.Nil(t, actual)
}

func TestYamlToSchemaErrors(t *testing.T) {
	_, err := YamlToSchema("wrong yaml")
	assert.NotNil(t, err)
	_, err = YamlToSchema("name: wrongType")
	assert.NotNil(t, err)
}

func TestSchemaHasField(t *testing.T) {
	s, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, s.HasField("name"))
	assert.True(t, s.HasField("price"))
	assert.True(t, s.HasField("quantity"))
	assert.False(t, s.HasField("unexisting"))
}

func TestToBytesFilledUnmodified(t *testing.T) {
	b := getBufferAllFields(t, int32(1), int64(2), float32(3), float64(4), "str", byte(5))
	bytes := b.ToBytes()
	b = ReadBuffer(bytes, b.schema)
	testFieldValues(t, b, int32(1), int64(2), float32(3), float64(4), "str", byte(5))
}

func TestFieldTypes(t *testing.T) {
	b := getBufferAllFields(t, int32(1), int64(2), float32(3), float64(4), "str", byte(5))
	testFieldValues(t, b, int32(1), int64(2), float32(3), float64(4), "str", byte(5))
}

func TestDefaultValuesAreValidValues(t *testing.T) {
	// in FlatBuffers setting field to default value is considered as the field is unset
	b := getBufferAllFields(t, int32(0), int64(0), float32(0), float64(0), "", byte(0))
	testFieldValues(t, b, int32(0), int64(0), float32(0), float64(0), "", byte(0))
}

func getBufferAllFields(t *testing.T, expectedInt32 int32, expectedInt64 int64, expectedFloat32 float32, expectedFloat64 float64, expectedString string, expectedByte byte) *Buffer {
	s, err := YamlToSchema(allTypesYaml)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(s)
	b.Set("int", expectedInt32)
	b.Set("long", expectedInt64)
	b.Set("float", expectedFloat32)
	b.Set("double", expectedFloat64)
	b.Set("string", expectedString)
	b.Set("boolFalse", false)
	b.Set("boolTrue", true)
	b.Set("byte", expectedByte)
	bytes := b.ToBytes()
	b = ReadBuffer(bytes, s)
	return b
}

func testFieldValues(t *testing.T, b *Buffer, expectedInt32 int32, expectedInt64 int64, expectedFloat32 float32, expectedFloat64 float64, expectedString string, expectedByte byte) {
	actual := b.Get("int")
	actualInt, ok := b.GetInt("int")
	assert.Equal(t, expectedInt32, actual)
	assert.Equal(t, expectedInt32, actualInt)
	assert.True(t, ok)

	actual = b.Get("long")
	actualLong, ok := b.GetLong("long")
	assert.Equal(t, expectedInt64, actual)
	assert.Equal(t, expectedInt64, actualLong)
	assert.True(t, ok)

	actual = b.Get("float")
	actualFloat, ok := b.GetFloat("float")
	assert.Equal(t, expectedFloat32, actual)
	assert.Equal(t, expectedFloat32, actualFloat)
	assert.True(t, ok)

	actual = b.Get("double")
	actualDouble, ok := b.GetDouble("double")
	assert.Equal(t, expectedFloat64, actual)
	assert.Equal(t, expectedFloat64, actualDouble)
	assert.True(t, ok)

	actual = b.Get("string")
	actualString, ok := b.GetString("string")
	assert.Equal(t, expectedString, actual)
	assert.Equal(t, expectedString, actualString)
	assert.True(t, ok)

	actual = b.Get("byte")
	actualByte, ok := b.GetByte("byte")
	assert.Equal(t, expectedByte, actual)
	assert.Equal(t, expectedByte, actualByte)
	assert.True(t, ok)

	actual = b.Get("boolTrue")
	actualBool, ok := b.GetBool("boolTrue")
	assert.Equal(t, true, actual)
	assert.Equal(t, true, actualBool)
	assert.True(t, ok)

	actual = b.Get("boolFalse")
	actualBool, ok = b.GetBool("boolFalse")
	assert.Equal(t, false, actual)
	assert.Equal(t, false, actualBool)
	assert.True(t, ok)
}

func TestToJSON(t *testing.T) {
	schema, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}

	b := NewBuffer(schema)
	dest := map[string]interface{}{}
	jsonStr := b.ToJSON()
	json.Unmarshal([]byte(jsonStr), &dest)
	assert.True(t, len(dest) == 0)

	// basic test
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	jsonStr = b.ToJSON()
	json.Unmarshal([]byte(jsonStr), &dest)
	assert.True(t, len(dest) == 3)
	assert.Equal(t, "cola", dest["name"])
	assert.Equal(t, float64(0.123), dest["price"])
	assert.Equal(t, float64(42), dest["quantity"])

	// test field initially not set
	b = NewBuffer(schema)
	b.Set("name", "cola")
	b.Set("quantity", int32(42))
	jsonStr = b.ToJSON()
	dest = map[string]interface{}{}
	json.Unmarshal([]byte(jsonStr), &dest)
	assert.True(t, len(dest) == 2)
	assert.Equal(t, "cola", dest["name"])
	assert.Equal(t, float64(42), dest["quantity"])

	// test field not set on ReadBuffer
	bytes := b.ToBytes()
	b = ReadBuffer(bytes, schema)

	jsonStr = b.ToJSON()
	dest = map[string]interface{}{}
	json.Unmarshal([]byte(jsonStr), &dest)
	assert.True(t, len(dest) == 2)
	assert.Equal(t, "cola", dest["name"])
	assert.Equal(t, float64(42), dest["quantity"])

	// test unset field
	b.Set("quantity", nil)
	jsonStr = b.ToJSON()
	dest = map[string]interface{}{}
	json.Unmarshal([]byte(jsonStr), &dest)
	assert.True(t, len(dest) == 1)
	assert.Equal(t, "cola", dest["name"])

	// test read unset field
	bytes = b.ToBytes()
	b = ReadBuffer(bytes, schema)
	jsonStr = b.ToJSON()
	dest = map[string]interface{}{}
	json.Unmarshal([]byte(jsonStr), &dest)
	assert.True(t, len(dest) == 1)
	assert.Equal(t, "cola", dest["name"])
}

func TestDifferentOrder(t *testing.T) {
	s, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(s)
	b.Set("quantity", int32(42))
	b.Set("Name", "cola")
	b.Set("price", float32(0.123))
	bytes := b.ToBytes()

	b = ReadBuffer(bytes, s)
	b.Set("price", float32(0.124))
	b.Set("name", "new cola")
	bytes = b.ToBytes()

	b = ReadBuffer(bytes, s)
	actual := b.Get("name")
	assert.Equal(t, "new cola", actual.(string))
	actual = b.Get("price")
	assert.Equal(t, float32(0.124), actual.(float32))
	actual = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))
}
