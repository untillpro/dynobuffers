/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var schemeStr = `
name: string
price: float
quantity: int
`

var schemeStrNew = `
name: string
price: float
quantity: int
newField: long
`

var schemeMandatory = `
name: string
price: ~float
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
	s, err := YamlToScheme(schemeStrNew)
	if err != nil {
		t.Fatal(err)
	}

	// create new from sratch
	b := NewBuffer(s)
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	b.Set("unknownField", "") // no errors, nothing will be made on ToBytes()
	bytes, _ := b.ToBytes()

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
	b.Set("name", nil) // set to nil means unset
	bytes, _ = b.ToBytes()
	b = ReadBuffer(bytes, s)
	actual = b.Get("name")
	assert.Nil(t, actual)
	actual = b.Get("price")
	assert.Equal(t, float32(0.124), actual.(float32))
	actual = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))

	// set untyped int value
	b.Set("quantity", 45)
	bytes, _ = b.ToBytes()
	b = ReadBuffer(bytes, s)
	actual = b.Get("quantity")
	assert.Equal(t, int32(45), actual.(int32))

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
	s, err := YamlToScheme(allTypesYaml)
	if err != nil {
		t.Fatal(err)
	}

	// test initially unset
	b := NewBuffer(s)
	testNilFields(t, b)
	bytes, _ := b.ToBytes()
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
	bytes, _ = b.ToBytes()
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
	bytes, _ = b.ToBytes()
	b = ReadBuffer(bytes, s)
	b.Set("int", nil)
	b.Set("long", nil)
	b.Set("float", nil)
	b.Set("double", nil)
	b.Set("string", nil)
	b.Set("boolFalse", nil)
	b.Set("boolTrue", nil)
	b.Set("byte", nil)
	bytes, _ = b.ToBytes()
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
	schemeNew, err := YamlToScheme(schemeStrNew)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(schemeNew)
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	b.Set("newField", int64(1))
	bytesNew, _ := b.ToBytes()

	schemeOld, err := YamlToScheme(schemeStr)
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytesNew, schemeOld)

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
	schemeOld, err := YamlToScheme(schemeStr)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(schemeOld)
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	bytesOld, _ := b.ToBytes()

	schemeNew, err := YamlToScheme(schemeStrNew)
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytesOld, schemeNew)

	actual := b.Get("name")
	assert.Equal(t, "cola", actual.(string))
	actual = b.Get("price")
	assert.Equal(t, float32(0.123), actual.(float32))
	actual = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))

	actual = b.Get("newField")
	assert.Nil(t, actual)
}

func TestYamlToSchemeErrors(t *testing.T) {
	_, err := YamlToScheme("wrong yaml")
	assert.NotNil(t, err)
	_, err = YamlToScheme("name: wrongType")
	assert.NotNil(t, err)
}

func TestSchemeHasField(t *testing.T) {
	s, err := YamlToScheme(schemeStr)
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
	bytes, _ := b.ToBytes()
	b = ReadBuffer(bytes, b.scheme)
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
	s, err := YamlToScheme(allTypesYaml)
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
	bytes, _ := b.ToBytes()
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
	scheme, err := YamlToScheme(schemeStr)
	if err != nil {
		t.Fatal(err)
	}

	b := NewBuffer(scheme)
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
	b = NewBuffer(scheme)
	b.Set("name", "cola")
	b.Set("quantity", int32(42))
	jsonStr = b.ToJSON()
	dest = map[string]interface{}{}
	json.Unmarshal([]byte(jsonStr), &dest)
	assert.True(t, len(dest) == 2)
	assert.Equal(t, "cola", dest["name"])
	assert.Equal(t, float64(42), dest["quantity"])

	// test field not set on ReadBuffer
	bytes, _ := b.ToBytes()
	b = ReadBuffer(bytes, scheme)

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
	bytes, _ = b.ToBytes()
	b = ReadBuffer(bytes, scheme)
	jsonStr = b.ToJSON()
	dest = map[string]interface{}{}
	json.Unmarshal([]byte(jsonStr), &dest)
	assert.True(t, len(dest) == 1)
	assert.Equal(t, "cola", dest["name"])
}

func TestDifferentOrder(t *testing.T) {
	s, err := YamlToScheme(schemeStr)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(s)
	b.Set("quantity", int32(42))
	b.Set("Name", "cola")
	b.Set("price", float32(0.123))
	bytes, _ := b.ToBytes()

	b = ReadBuffer(bytes, s)
	b.Set("price", float32(0.124))
	b.Set("name", "new cola")
	bytes, _ = b.ToBytes()

	b = ReadBuffer(bytes, s)
	actual := b.Get("name")
	assert.Equal(t, "new cola", actual.(string))
	actual = b.Get("price")
	assert.Equal(t, float32(0.124), actual.(float32))
	actual = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))
}

func TestSchemeToFromYaml(t *testing.T) {
	scheme, err := YamlToScheme(schemeStr)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := yaml.Marshal(scheme)
	schemeNew := NewScheme()
	err = yaml.Unmarshal(bytes, &schemeNew)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, reflect.DeepEqual(scheme.Fields, schemeNew.Fields))
	assert.True(t, reflect.DeepEqual(scheme.fieldsMap, schemeNew.fieldsMap))
	assert.True(t, reflect.DeepEqual(scheme.stringFields, schemeNew.stringFields))

	schemeNew = NewScheme()
	err = yaml.Unmarshal([]byte("wrong "), &schemeNew)
	if err == nil {
		t.Fatal()
	}
}

func TestCanBeAssigned(t *testing.T) {
	scheme, err := YamlToScheme(allTypesYaml)
	if err != nil {
		t.Fatal(err)
	}

	assert.False(t, scheme.CanBeAssigned("unexisting", 0.123))

	assert.False(t, scheme.CanBeAssigned("int", 0.123))
	assert.True(t, scheme.CanBeAssigned("int", 0.0))
	assert.True(t, scheme.CanBeAssigned("int", -0.0))
	assert.False(t, scheme.CanBeAssigned("int", true))
	assert.False(t, scheme.CanBeAssigned("int", "sdsd"))
	assert.False(t, scheme.CanBeAssigned("int", math.MaxFloat32))
	assert.False(t, scheme.CanBeAssigned("int", math.MaxFloat64))
	assert.False(t, scheme.CanBeAssigned("int", math.MaxInt64))
	assert.False(t, scheme.CanBeAssigned("int", math.MaxInt32))
	assert.True(t, scheme.CanBeAssigned("int", float64(math.MaxInt32)))
	assert.False(t, scheme.CanBeAssigned("int", float64(5000000000000))) //int32 overflow
	assert.False(t, scheme.CanBeAssigned("int", byte(1)))

	assert.False(t, scheme.CanBeAssigned("long", 0.123))
	assert.True(t, scheme.CanBeAssigned("long", 0.0))
	assert.True(t, scheme.CanBeAssigned("long", -0.0))
	assert.False(t, scheme.CanBeAssigned("long", true))
	assert.False(t, scheme.CanBeAssigned("long", "sdsd"))
	assert.False(t, scheme.CanBeAssigned("long", math.MaxFloat32))
	assert.False(t, scheme.CanBeAssigned("long", math.MaxFloat64))
	assert.False(t, scheme.CanBeAssigned("long", math.MaxInt32))
	assert.True(t, scheme.CanBeAssigned("long", float64(math.MaxInt32)))
	assert.False(t, scheme.CanBeAssigned("long", float64(math.MaxInt64))) // unsupported
	assert.False(t, scheme.CanBeAssigned("long", math.MaxInt64))
	assert.False(t, scheme.CanBeAssigned("long", float64(500000000000000000000000))) // int64 overflow
	assert.True(t, scheme.CanBeAssigned("long", float64(5000000000000)))
	assert.True(t, scheme.CanBeAssigned("long", float64(math.MaxInt32)))

	assert.False(t, scheme.CanBeAssigned("string", 0.123))
	assert.False(t, scheme.CanBeAssigned("string", 0.0))
	assert.False(t, scheme.CanBeAssigned("string", -0.0))
	assert.False(t, scheme.CanBeAssigned("string", true))
	assert.True(t, scheme.CanBeAssigned("string", "sdsd"))
	assert.False(t, scheme.CanBeAssigned("string", math.MaxFloat32))
	assert.False(t, scheme.CanBeAssigned("string", math.MaxFloat64))
	assert.False(t, scheme.CanBeAssigned("string", float64(5000000000000)))
	assert.False(t, scheme.CanBeAssigned("string", float64(math.MaxInt32)))

	assert.True(t, scheme.CanBeAssigned("float", 0.123))
	assert.True(t, scheme.CanBeAssigned("float", 0.0))
	assert.True(t, scheme.CanBeAssigned("float", -0.0))
	assert.False(t, scheme.CanBeAssigned("float", true))
	assert.False(t, scheme.CanBeAssigned("float", "sdsd"))
	assert.True(t, scheme.CanBeAssigned("float", float64(math.MaxFloat32)))
	assert.True(t, scheme.CanBeAssigned("float", float64(math.MaxFloat64)))
	assert.False(t, scheme.CanBeAssigned("float", float64(5000000000000)))
	assert.True(t, scheme.CanBeAssigned("float", float64(math.MaxInt32)))

	assert.True(t, scheme.CanBeAssigned("double", 0.123))
	assert.True(t, scheme.CanBeAssigned("double", 0.0))
	assert.True(t, scheme.CanBeAssigned("double", -0.0))
	assert.False(t, scheme.CanBeAssigned("double", true))
	assert.False(t, scheme.CanBeAssigned("double", "sdsd"))
	assert.True(t, scheme.CanBeAssigned("double", float64(math.MaxFloat32)))
	assert.True(t, scheme.CanBeAssigned("double", float64(math.MaxFloat64)))
	assert.True(t, scheme.CanBeAssigned("double", float64(5000000000000)))
	assert.True(t, scheme.CanBeAssigned("double", float64(math.MaxInt32)))
	assert.True(t, scheme.CanBeAssigned("double", float64(math.MaxInt64)))

	assert.False(t, scheme.CanBeAssigned("boolTrue", 0.123))
	assert.False(t, scheme.CanBeAssigned("boolTrue", 0.0))
	assert.False(t, scheme.CanBeAssigned("boolTrue", -0.0))
	assert.True(t, scheme.CanBeAssigned("boolTrue", true))
	assert.False(t, scheme.CanBeAssigned("boolTrue", "sdsd"))
	assert.False(t, scheme.CanBeAssigned("boolTrue", float64(math.MaxFloat32)))
	assert.False(t, scheme.CanBeAssigned("boolTrue", float64(math.MaxFloat64)))
	assert.False(t, scheme.CanBeAssigned("boolTrue", float64(5000000000000)))
	assert.False(t, scheme.CanBeAssigned("boolTrue", float64(math.MaxInt32)))
	assert.False(t, scheme.CanBeAssigned("boolTrue", float64(math.MaxInt64)))

	assert.False(t, scheme.CanBeAssigned("byte", 0.123))
	assert.True(t, scheme.CanBeAssigned("byte", 0.0))
	assert.True(t, scheme.CanBeAssigned("byte", -0.0))
	assert.False(t, scheme.CanBeAssigned("byte", true))
	assert.False(t, scheme.CanBeAssigned("byte", "sdsd"))
	assert.False(t, scheme.CanBeAssigned("byte", float64(-255)))
	assert.True(t, scheme.CanBeAssigned("byte", float64(255)))
	assert.False(t, scheme.CanBeAssigned("byte", float64(-256)))
	assert.False(t, scheme.CanBeAssigned("byte", float64(256)))
	assert.False(t, scheme.CanBeAssigned("byte", float64(math.MaxFloat32)))
	assert.False(t, scheme.CanBeAssigned("byte", float64(math.MaxFloat64)))
	assert.False(t, scheme.CanBeAssigned("byte", float64(5000000000000)))
	assert.False(t, scheme.CanBeAssigned("byte", float64(math.MaxInt32)))
	assert.False(t, scheme.CanBeAssigned("byte", float64(math.MaxInt64)))
}

func TestMandatoryFields(t *testing.T) {
	scheme, err := YamlToScheme(schemeMandatory)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(scheme)
	bytes, err := b.ToBytes()
	assert.NotNil(t, err)
	assert.Nil(t, bytes)

	b.Set("name", "str")
	bytes, err = b.ToBytes()
	assert.NotNil(t, err)
	assert.Nil(t, bytes)

	b.Set("price", 0.123)
	bytes, err = b.ToBytes()
	assert.Nil(t, err)
	assert.NotNil(t, bytes)

	b = ReadBuffer(bytes, scheme) 
	bytes, err = b.ToBytes()
	assert.Nil(t, err)
	assert.NotNil(t, bytes)

}

