/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"encoding/json"
	"reflect"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
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
Price: float
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

var arraysAllTypesYaml = `
ints..: int
longs..: long
floats..: float
doubles..: double
strings..: string
boolTrues..: bool
boolFalses..: bool
bytes..: byte
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
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

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
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, s)
	actual = b.Get("name")
	assert.Nil(t, actual)
	actual = b.Get("price")
	assert.Equal(t, float32(0.124), actual.(float32))
	actual = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))

	// set untyped int value
	b.Set("quantity", 45)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
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
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
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
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
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
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, s)
	b.Set("int", nil)
	b.Set("long", nil)
	b.Set("float", nil)
	b.Set("double", nil)
	b.Set("string", nil)
	b.Set("boolFalse", nil)
	b.Set("boolTrue", nil)
	b.Set("byte", nil)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
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
	bytesNew, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

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
	bytesOld, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

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
	_, err = YamlToScheme(`
		nested:
		  nestedField: wrongType`)
	assert.NotNil(t, err)
	_, err = YamlToScheme(`
		nested:
		  wrong`)
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
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
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
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
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
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
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
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
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
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	b = ReadBuffer(bytes, s)
	b.Set("price", float32(0.124))
	b.Set("name", "new cola")
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

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
	scheme.AddField("mandatory", FieldTypeInt, true)
	bytes, err := yaml.Marshal(scheme)
	schemeNew := NewScheme()
	err = yaml.Unmarshal(bytes, &schemeNew)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, reflect.DeepEqual(scheme.Fields, schemeNew.Fields))
	assert.True(t, reflect.DeepEqual(scheme.fieldsMap, schemeNew.fieldsMap))

	schemeNew = NewScheme()
	err = yaml.Unmarshal([]byte("wrong "), &schemeNew)
	if err == nil {
		t.Fatal()
	}
}

func TestSchemeNestedToFromYAML(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nes", schemeNested, true)
	schemeRoot.AddNestedArray("nesarr", schemeNested, true)
	schemeRoot.AddField("last", FieldTypeInt, false)
	bytes, err := yaml.Marshal(schemeRoot)
	schemeNew := NewScheme()
	err = yaml.Unmarshal(bytes, &schemeNew)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, reflect.DeepEqual(schemeRoot.Fields, schemeNew.Fields))
	assert.True(t, reflect.DeepEqual(schemeRoot.fieldsMap, schemeNew.fieldsMap))

	schemeNew = NewScheme()
	err = yaml.Unmarshal([]byte("wrong "), &schemeNew)
	if err == nil {
		t.Fatal()
	}
}

// func TestCanBeAssigned(t *testing.T) {
// 	scheme, err := YamlToScheme(allTypesYaml)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	assert.False(t, scheme.CanBeAssigned("unexisting", 0.123))

// 	assert.False(t, scheme.CanBeAssigned("int", 0.123))
// 	assert.True(t, scheme.CanBeAssigned("int", 0.0))
// 	assert.True(t, scheme.CanBeAssigned("int", -0.0))
// 	assert.False(t, scheme.CanBeAssigned("int", true))
// 	assert.False(t, scheme.CanBeAssigned("int", "sdsd"))
// 	assert.False(t, scheme.CanBeAssigned("int", math.MaxFloat32))
// 	assert.False(t, scheme.CanBeAssigned("int", math.MaxFloat64))
// 	assert.False(t, scheme.CanBeAssigned("int", math.MaxInt64))
// 	assert.False(t, scheme.CanBeAssigned("int", math.MaxInt32))
// 	assert.True(t, scheme.CanBeAssigned("int", float64(math.MaxInt32)))
// 	assert.False(t, scheme.CanBeAssigned("int", float64(5000000000000))) //int32 overflow
// 	assert.False(t, scheme.CanBeAssigned("int", byte(1)))

// 	assert.False(t, scheme.CanBeAssigned("long", 0.123))
// 	assert.True(t, scheme.CanBeAssigned("long", 0.0))
// 	assert.True(t, scheme.CanBeAssigned("long", -0.0))
// 	assert.False(t, scheme.CanBeAssigned("long", true))
// 	assert.False(t, scheme.CanBeAssigned("long", "sdsd"))
// 	assert.False(t, scheme.CanBeAssigned("long", math.MaxFloat32))
// 	assert.False(t, scheme.CanBeAssigned("long", math.MaxFloat64))
// 	assert.False(t, scheme.CanBeAssigned("long", math.MaxInt32))
// 	assert.True(t, scheme.CanBeAssigned("long", float64(math.MaxInt32)))
// 	assert.False(t, scheme.CanBeAssigned("long", float64(math.MaxInt64))) // unsupported
// 	assert.False(t, scheme.CanBeAssigned("long", math.MaxInt64))
// 	assert.False(t, scheme.CanBeAssigned("long", float64(500000000000000000000000))) // int64 overflow
// 	assert.True(t, scheme.CanBeAssigned("long", float64(5000000000000)))
// 	assert.True(t, scheme.CanBeAssigned("long", float64(math.MaxInt32)))

// 	assert.False(t, scheme.CanBeAssigned("string", 0.123))
// 	assert.False(t, scheme.CanBeAssigned("string", 0.0))
// 	assert.False(t, scheme.CanBeAssigned("string", -0.0))
// 	assert.False(t, scheme.CanBeAssigned("string", true))
// 	assert.True(t, scheme.CanBeAssigned("string", "sdsd"))
// 	assert.False(t, scheme.CanBeAssigned("string", math.MaxFloat32))
// 	assert.False(t, scheme.CanBeAssigned("string", math.MaxFloat64))
// 	assert.False(t, scheme.CanBeAssigned("string", float64(5000000000000)))
// 	assert.False(t, scheme.CanBeAssigned("string", float64(math.MaxInt32)))

// 	assert.True(t, scheme.CanBeAssigned("float", 0.123))
// 	assert.True(t, scheme.CanBeAssigned("float", 0.0))
// 	assert.True(t, scheme.CanBeAssigned("float", -0.0))
// 	assert.False(t, scheme.CanBeAssigned("float", true))
// 	assert.False(t, scheme.CanBeAssigned("float", "sdsd"))
// 	assert.True(t, scheme.CanBeAssigned("float", float64(math.MaxFloat32)))
// 	assert.True(t, scheme.CanBeAssigned("float", float64(math.MaxFloat64)))
// 	assert.False(t, scheme.CanBeAssigned("float", float64(5000000000000)))
// 	assert.True(t, scheme.CanBeAssigned("float", float64(math.MaxInt32)))

// 	assert.True(t, scheme.CanBeAssigned("double", 0.123))
// 	assert.True(t, scheme.CanBeAssigned("double", 0.0))
// 	assert.True(t, scheme.CanBeAssigned("double", -0.0))
// 	assert.False(t, scheme.CanBeAssigned("double", true))
// 	assert.False(t, scheme.CanBeAssigned("double", "sdsd"))
// 	assert.True(t, scheme.CanBeAssigned("double", float64(math.MaxFloat32)))
// 	assert.True(t, scheme.CanBeAssigned("double", float64(math.MaxFloat64)))
// 	assert.True(t, scheme.CanBeAssigned("double", float64(5000000000000)))
// 	assert.True(t, scheme.CanBeAssigned("double", float64(math.MaxInt32)))
// 	assert.True(t, scheme.CanBeAssigned("double", float64(math.MaxInt64)))

// 	assert.False(t, scheme.CanBeAssigned("boolTrue", 0.123))
// 	assert.False(t, scheme.CanBeAssigned("boolTrue", 0.0))
// 	assert.False(t, scheme.CanBeAssigned("boolTrue", -0.0))
// 	assert.True(t, scheme.CanBeAssigned("boolTrue", true))
// 	assert.False(t, scheme.CanBeAssigned("boolTrue", "sdsd"))
// 	assert.False(t, scheme.CanBeAssigned("boolTrue", float64(math.MaxFloat32)))
// 	assert.False(t, scheme.CanBeAssigned("boolTrue", float64(math.MaxFloat64)))
// 	assert.False(t, scheme.CanBeAssigned("boolTrue", float64(5000000000000)))
// 	assert.False(t, scheme.CanBeAssigned("boolTrue", float64(math.MaxInt32)))
// 	assert.False(t, scheme.CanBeAssigned("boolTrue", float64(math.MaxInt64)))

// 	assert.False(t, scheme.CanBeAssigned("byte", 0.123))
// 	assert.True(t, scheme.CanBeAssigned("byte", 0.0))
// 	assert.True(t, scheme.CanBeAssigned("byte", -0.0))
// 	assert.False(t, scheme.CanBeAssigned("byte", true))
// 	assert.False(t, scheme.CanBeAssigned("byte", "sdsd"))
// 	assert.False(t, scheme.CanBeAssigned("byte", float64(-255)))
// 	assert.True(t, scheme.CanBeAssigned("byte", float64(255)))
// 	assert.False(t, scheme.CanBeAssigned("byte", float64(-256)))
// 	assert.False(t, scheme.CanBeAssigned("byte", float64(256)))
// 	assert.False(t, scheme.CanBeAssigned("byte", float64(math.MaxFloat32)))
// 	assert.False(t, scheme.CanBeAssigned("byte", float64(math.MaxFloat64)))
// 	assert.False(t, scheme.CanBeAssigned("byte", float64(5000000000000)))
// 	assert.False(t, scheme.CanBeAssigned("byte", float64(math.MaxInt32)))
// 	assert.False(t, scheme.CanBeAssigned("byte", float64(math.MaxInt64)))
// }

// func TestCanBeAssignedMandatory(t *testing.T) {
// 	scheme, err := YamlToScheme(schemeMandatory)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.True(t, scheme.CanBeAssigned("name", nil))
// 	assert.False(t, scheme.CanBeAssigned("price", nil))
// }

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

func TestApplyJsonErrors(t *testing.T) {
	scheme, err := YamlToScheme(schemeMandatory)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(scheme)

	// unset field
	json := []byte(`{"name": null, "price": 0.124, "unknown": 42}`)
	bytes, err := b.ApplyJSONAndToBytes(json)
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, scheme)
	assert.Nil(t, b.Get("name"))
	assert.Equal(t, float32(0.124), b.Get("price"))

	// wrong type -> error
	json = []byte(`{"name": "str", "price": "wrong type", "unknown": 42}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	assert.Nil(t, bytes)
	assert.NotNil(t, err)

	// unset mandatory field -> error
	json = []byte(`{"name": "str", "price": null, "unknown": 42}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	assert.Nil(t, bytes)
	assert.NotNil(t, err)

	// mandatory field is not set -> error
	json = []byte(`{"name": "str", "unknown": 42}`)
	b = NewBuffer(scheme)
	bytes, err = b.ApplyJSONAndToBytes(json)
	assert.Nil(t, bytes)
	assert.NotNil(t, err)

	// wrong json -> error
	json = []byte(`wrong`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	assert.Nil(t, bytes)
	assert.NotNil(t, err)
}

func TestApplyJSONAdvanced(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nested", schemeNested, true)
	schemeRoot.AddNestedArray("nestedarr", schemeNested, false)
	schemeRoot.AddArray("ids", FieldTypeInt, false)

	bRoot := NewBuffer(schemeRoot)
	bytes, err := bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str1", "nested": {"price": 0.123,"quantity":42},"nestedarr":[{"price": 0.124,"quantity":43},{"price": 0.125,"quantity":44}], "ids": [45,46]}`))
	if err != nil {
		t.Fatal(err)
	}
	bRoot = ReadBuffer(bytes, schemeRoot)
	assert.Equal(t, "str1", bRoot.Get("name"))
	ints := bRoot.Get("ids").(*Array).GetAll()
	intf := ints[0]
	intVal, ok := intf.(int32)
	assert.True(t, ok)
	assert.Equal(t, int32(45), intVal)
	assert.True(t, reflect.DeepEqual([]interface{}{int32(45), int32(46)}, ints))
	bRoot.Set("ids", ints)
	bytes, err = bRoot.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	bRoot = ReadBuffer(bytes, schemeRoot)
	assert.True(t, reflect.DeepEqual([]interface{}{int32(45), int32(46)}, bRoot.Get("ids").(*Array).GetAll()))

}

func TestApplyJsonAllTypes(t *testing.T) {
	scheme, err := YamlToScheme(allTypesYaml)
	if err != nil {
		t.Fatal(err)
	}
	json := []byte(`{"string": "str", "long": 42, "int": 43, "float": 0.124, "double": 0.125, "byte": 6, "boolTrue": true, "boolFalse": false, "unknown": -1}`)
	b := NewBuffer(scheme)
	bytes, err := b.ApplyJSONAndToBytes(json)
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, scheme)
	testFieldValues(t, b, int32(43), int64(42), float32(0.124), float64(0.125), "str", 6)
	assert.True(t, b.Get("boolTrue").(bool))
	assert.False(t, b.Get("boolFalse").(bool))
}

func TestNestedBasic(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nes", schemeNested, true)
	schemeRoot.AddField("last", FieldTypeInt, false)

	b := NewBuffer(schemeRoot)
	bNested := NewBuffer(schemeNested)
	bNested.Set("price", 0.123)
	bNested.Set("quantity", 42)
	b.Set("name", "str")
	b.Set("nes", bNested)
	b.Set("last", 42)
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	b = ReadBuffer(bytes, schemeRoot)
	assert.Equal(t, "str", b.Get("name"))
	assert.Equal(t, int32(42), b.Get("last"))

	bNested = b.Get("nes").(*Buffer)
	assert.Equal(t, int32(42), bNested.Get("quantity"))
	assert.Equal(t, float32(0.123), bNested.Get("price"))
}

func TestNestedAdvanced(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, false)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nes", schemeNested, false)
	schemeRoot.AddField("last", FieldTypeInt, false)
	b := NewBuffer(schemeRoot)
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// test initial
	b = ReadBuffer(bytes, schemeRoot)
	assert.Nil(t, b.Get("name"))
	assert.Nil(t, b.Get("nes"))
	assert.Nil(t, b.Get("last"))

	// nested still nil after Get() and not modify
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, schemeRoot)
	assert.Nil(t, b.Get("nes"))

	// fill
	bNested := NewBuffer(schemeNested)
	bNested.Set("price", 0.123)
	bNested.Set("quantity", 42)
	b.Set("name", "str")
	b.Set("nes", bNested)
	b.Set("last", 42)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// modify nested
	b = ReadBuffer(bytes, schemeRoot)
	bNested = b.Get("nes").(*Buffer)
	bNested.Set("quantity", 43)
	bNested.Set("price", 0.124)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, schemeRoot)
	assert.Equal(t, "str", b.Get("name"))
	assert.Equal(t, int32(42), b.Get("last"))
	bNested = b.Get("nes").(*Buffer)
	assert.Equal(t, int32(43), bNested.Get("quantity"))
	assert.Equal(t, float32(0.124), bNested.Get("price"))

	// unset nested
	b.Set("nes", nil)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, schemeRoot)
	assert.Equal(t, "str", b.Get("name"))
	assert.Equal(t, int32(42), b.Get("last"))
	assert.Nil(t, b.Get("nes"))
}

func TestNestedMandatory(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nes", schemeNested, true)
	schemeRoot.AddField("last", FieldTypeInt, false)
	b := NewBuffer(schemeRoot)
	bytes, err := b.ToBytes()
	if err == nil {
		t.Fatal()
	}

	// fill
	bNested := NewBuffer(schemeNested)
	bNested.Set("price", 0.123)
	bNested.Set("quantity", 42)
	b.Set("name", "str")
	b.Set("nes", bNested)
	b.Set("last", 42)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// read and to bytes. Nested should be copied
	b = ReadBuffer(bytes, schemeRoot)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// unset mandatory in nested
	b = ReadBuffer(bytes, schemeRoot)
	bNested = b.Get("nes").(*Buffer)
	bNested.Set("quantity", nil)
	bytes, err = b.ToBytes()
	if err == nil {
		t.Fatal()
	}

	// unset nested mandatory
	bNested.Set("quantity", 1)
	b.Set("nes", nil)
	bytes, err = b.ToBytes()
	if err == nil {
		t.Fatal(err)
	}
}

func TestApplyNestedJSON(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nes", schemeNested, true)
	schemeRoot.AddField("last", FieldTypeInt, false)

	b := NewBuffer(schemeRoot)
	bNested := NewBuffer(schemeNested)
	bNested.Set("price", 0.123)
	bNested.Set("quantity", 42)
	b.Set("name", "str")
	b.Set("nes", bNested)
	b.Set("last", 42)
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, schemeRoot)
	jsonStr := b.ToJSON()

	b = NewBuffer(schemeRoot)
	bytes, err = b.ApplyJSONAndToBytes([]byte(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, schemeRoot)
	bNested = b.Get("nes").(*Buffer)
	assert.Equal(t, int32(42), bNested.Get("quantity"))
	assert.Equal(t, float32(0.123), bNested.Get("price"))
	assert.Equal(t, "str", b.Get("name"))
	assert.Equal(t, int32(42), b.Get("last"))

	// error if mandatory nested object is null
	b = NewBuffer(schemeRoot)
	_, err = b.ApplyJSONAndToBytes([]byte(`{"name":"str","nes":null,"last":42}`))
	if err == nil {
		t.Fatal()
	}
	b = NewBuffer(schemeRoot)
	_, err = b.ApplyJSONAndToBytes([]byte(`{"name":"str","last":42}`))
	if err == nil {
		t.Fatal()
	}

	// error if mandatory field in nested object is null
	b = NewBuffer(schemeRoot)
	_, err = b.ApplyJSONAndToBytes([]byte(`{"name":"str","nes":{"price": 1,"quantity":null},"last":42}`))
	if err == nil {
		t.Fatal()
	}
	b = NewBuffer(schemeRoot)
	_, err = b.ApplyJSONAndToBytes([]byte(`{"name":"str","nes":{"price": 1},"last":42}`))
	if err == nil {
		t.Fatal()
	}
	b = NewBuffer(schemeRoot)
	_, err = b.ApplyJSONAndToBytes([]byte(`{"name":"str","nes":{},"last":42}`))
	if err == nil {
		t.Fatal()
	}
}

func TestApplyNestedArrayJSON(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNestedArray("nes", schemeNested, true)
	schemeRoot.AddField("last", FieldTypeInt, false)

	b := NewBuffer(schemeRoot)
	bNestedArr := []*Buffer{}
	bNested := NewBuffer(schemeNested)
	bNested.Set("price", 0.123)
	bNested.Set("quantity", 42)
	bNestedArr = append(bNestedArr, bNested)
	bNested = NewBuffer(schemeNested)
	bNested.Set("price", 0.124)
	bNested.Set("quantity", 43)
	bNestedArr = append(bNestedArr, bNested)

	b.Set("name", "str")
	b.Set("nes", bNestedArr)
	b.Set("last", 42)
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	b = ReadBuffer(bytes, schemeRoot)
	jsonStr := b.ToJSON()
	b = NewBuffer(schemeRoot)
	bytes, err = b.ApplyJSONAndToBytes([]byte(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, schemeRoot)
	bNested = b.GetByIndex("nes", 0).(*Buffer)
	assert.Equal(t, int32(42), bNested.Get("quantity"))
	assert.Equal(t, float32(0.123), bNested.Get("price"))
	bNested = b.GetByIndex("nes", 1).(*Buffer)
	assert.Equal(t, int32(43), bNested.Get("quantity"))
	assert.Equal(t, float32(0.124), bNested.Get("price"))
	assert.Equal(t, "str", b.Get("name"))
	assert.Equal(t, int32(42), b.Get("last"))

	// error if not array is provided for an array field
	bytes, err = b.ApplyJSONAndToBytes([]byte(`{"name":"str","nes":{"price":0.123,"quantity":42},"last":42}`))
	if err == nil {
		t.Fatal()
	}

	// error if not object is provided as an array element
	bytes, err = b.ApplyJSONAndToBytes([]byte(`{"name":"str","nes":[0.123],"last":42}`))
	if err == nil {
		t.Fatal()
	}

	// error if wrong failed to encode an array
	bytes, err = b.ApplyJSONAndToBytes([]byte(`{"name":"str","nes":[{"price":0.123,"quantity":"wrong value"}],"last":42}`))
	if err == nil {
		t.Fatal()
	}
}

func TestApplyArrayJSON(t *testing.T) {
	s := NewScheme()
	s.AddField("id", FieldTypeInt, false)
	s.AddArray("names", FieldTypeString, false)

	b := NewBuffer(s)
	b.Set("id", 42)
	names := []string{"str1", "str2"}
	b.Set("names", names)
	jsonStr := b.ToJSON()

	b = NewBuffer(s)
	bytes, err := b.ApplyJSONAndToBytes([]byte(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	b = ReadBuffer(bytes, s)
	assert.Equal(t, "str1", b.GetByIndex("names", 0))
	assert.Equal(t, "str2", b.GetByIndex("names", 1))
}

func TestFlatBuffersNested(t *testing.T) {
	bl := flatbuffers.NewBuilder(0)
	bl.StartObject(1)
	bl.PrependInt32(45)
	bl.Slot(0)
	nested := bl.EndObject()
	bl.Finish(nested)

	bl.StartObject(2)
	bl.PrependInt32(42)
	bl.Slot(0)
	bl.PrependUOffsetTSlot(1, nested, 0)
	root := bl.EndObject()
	bl.Finish(root)

	bytes := bl.FinishedBytes()

	tabRoot := flatbuffers.Table{}
	tabRoot.Bytes = bytes
	tabRoot.Pos = flatbuffers.GetUOffsetT(bytes)

	rootField0Offset := flatbuffers.UOffsetT(tabRoot.Offset(flatbuffers.VOffsetT((0+2)*2))) + tabRoot.Pos
	assert.Equal(t, int32(42), tabRoot.GetInt32(rootField0Offset))

	rootField1Offset := flatbuffers.UOffsetT(tabRoot.Offset(flatbuffers.VOffsetT((1+2)*2))) + tabRoot.Pos

	nestedOffset := tabRoot.Indirect(rootField1Offset)
	tabNested := flatbuffers.Table{}
	tabNested.Bytes = bytes
	tabNested.Pos = nestedOffset

	nestedField0Offset := flatbuffers.UOffsetT(tabNested.Offset(flatbuffers.VOffsetT((0+2)*2))) + tabNested.Pos
	assert.Equal(t, int32(45), tabNested.GetInt32(nestedField0Offset))

}

func TestArraysBasic(t *testing.T) {
	s := NewScheme()
	s.AddArray("longs", FieldTypeLong, false)

	// initial
	b := NewBuffer(s)
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, s)
	assert.Nil(t, b.Get("longs"))

	// set and read
	longs := []int64{5, 6}
	b.Set("longs", longs)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, s)
	assert.Equal(t, int64(5), b.GetByIndex("longs", 0))
	assert.Equal(t, int64(6), b.GetByIndex("longs", 1))
	assert.Nil(t, b.GetByIndex("longs", 3))
	assert.Nil(t, b.GetByIndex("longs", -1))
	assert.Nil(t, b.GetByIndex("unexisting", 0))

	// Non-modified array should be copied on ToBytes()
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, s)
	assert.Equal(t, int64(5), b.GetByIndex("longs", 0))
	assert.Equal(t, int64(6), b.GetByIndex("longs", 1))

	//test Array struct
	a := b.Get("longs").(*Array)
	longsActual := a.GetAll()
	assert.Equal(t, len(longs), len(longsActual))
	for i, long := range longsActual {
		assert.Equal(t, longs[i], long)
	}
	next, ok := a.GetNext()
	assert.True(t, ok)
	assert.Equal(t, int64(5), next)
	next, ok = a.GetNext()
	assert.True(t, ok)
	assert.Equal(t, int64(6), next)
	next, ok = a.GetNext()
	assert.False(t, ok)
	assert.Nil(t, next)

	// modify
	longs = []int64{7, 8, 9}
	b.Set("longs", longs)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, s)
	assert.Equal(t, int64(7), b.GetByIndex("longs", 0))
	assert.Equal(t, int64(8), b.GetByIndex("longs", 1))

	// unset
	b.Set("longs", nil)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, s)
	assert.Nil(t, b.Get("longs"))

	// set to empty
	longs = []int64{}
	b.Set("longs", longs)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, s)
	assert.Nil(t, b.GetByIndex("longs", 0))
	a = b.Get("longs").(*Array)
	next, ok = a.GetNext()
	assert.Nil(t, next)
	assert.False(t, ok)
	longsActual = a.GetAll()
	assert.True(t, len(longsActual) == 0)
}

func TestArraysAllTypes(t *testing.T) {
	s, err := YamlToScheme(arraysAllTypesYaml)
	if err != nil {
		t.Fatal(err)
	}

	ints := []int32{1, 2}
	longs := []int64{3, 4}
	floats := []float32{0.5, 0.6}
	doubles := []float64{0.7, 0.8}
	trueBools := []bool{true, false}
	falseBools := []bool{false, true}
	bytesArr := []byte{9, 10}
	strings := []string{"str1", "str2"}

	b := NewBuffer(s)
	b.Set("ints", ints)
	b.Set("longs", longs)
	b.Set("floats", floats)
	b.Set("doubles", doubles)
	b.Set("boolTrues", trueBools)
	b.Set("boolFalses", falseBools)
	b.Set("bytes", bytesArr)
	b.Set("strings", strings)
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	b = ReadBuffer(bytes, s)
	assert.Equal(t, int32(1), b.GetByIndex("ints", 0))
	assert.Equal(t, int32(2), b.GetByIndex("ints", 1))
	assert.Equal(t, int64(3), b.GetByIndex("longs", 0))
	assert.Equal(t, int64(4), b.GetByIndex("longs", 1))
	assert.Equal(t, float32(0.5), b.GetByIndex("floats", 0))
	assert.Equal(t, float32(0.6), b.GetByIndex("floats", 1))
	assert.Equal(t, float64(0.7), b.GetByIndex("doubles", 0))
	assert.Equal(t, float64(0.8), b.GetByIndex("doubles", 1))
	assert.Equal(t, true, b.GetByIndex("boolTrues", 0))
	assert.Equal(t, false, b.GetByIndex("boolTrues", 1))
	assert.Equal(t, false, b.GetByIndex("boolFalses", 0))
	assert.Equal(t, true, b.GetByIndex("boolFalses", 1))
	assert.Equal(t, byte(9), b.GetByIndex("bytes", 0))
	assert.Equal(t, byte(10), b.GetByIndex("bytes", 1))
	assert.Equal(t, "str1", b.GetByIndex("strings", 0))
	assert.Equal(t, "str2", b.GetByIndex("strings", 1))
}

func TestArraysNested(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, false)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNestedArray("nes", schemeNested, false)
	schemeRoot.AddField("last", FieldTypeInt, false)
	b := NewBuffer(schemeRoot)

	bNestedArr := []*Buffer{}
	bNested := NewBuffer(schemeNested)
	bNested.Set("price", 0.123)
	bNested.Set("quantity", 42)
	bNestedArr = append(bNestedArr, bNested)
	bNested = NewBuffer(schemeNested)
	bNested.Set("price", 0.124)
	bNested.Set("quantity", 43)
	bNestedArr = append(bNestedArr, bNested)

	b.Set("name", "str")
	b.Set("nes", bNestedArr)
	b.Set("last", 42)
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// GetByIndex
	b = ReadBuffer(bytes, schemeRoot)
	bNested = b.GetByIndex("nes", 0).(*Buffer)
	assert.Equal(t, int32(42), bNested.Get("quantity"))
	assert.Equal(t, float32(0.123), bNested.Get("price"))
	bNested = b.GetByIndex("nes", 1).(*Buffer)
	assert.Equal(t, int32(43), bNested.Get("quantity"))
	assert.Equal(t, float32(0.124), bNested.Get("price"))

	// iterator
	arr := b.Get("nes").(*Array)
	buffers := arr.GetAll()
	assert.Equal(t, int32(42), buffers[0].(*Buffer).Get("quantity"))
	assert.Equal(t, float32(0.123), buffers[0].(*Buffer).Get("price"))
	assert.Equal(t, int32(43), buffers[1].(*Buffer).Get("quantity"))
	assert.Equal(t, float32(0.124), buffers[1].(*Buffer).Get("price"))

	bufferIntf, ok := arr.GetNext()
	assert.True(t, ok)
	assert.Equal(t, int32(42), bufferIntf.(*Buffer).Get("quantity"))
	assert.Equal(t, float32(0.123), bufferIntf.(*Buffer).Get("price"))
	bufferIntf, ok = arr.GetNext()
	assert.True(t, ok)
	assert.Equal(t, int32(43), bufferIntf.(*Buffer).Get("quantity"))
	assert.Equal(t, float32(0.124), bufferIntf.(*Buffer).Get("price"))
	bufferIntf, ok = arr.GetNext()
	assert.False(t, ok)
	assert.Nil(t, bufferIntf)

	// set one of elements to nil
	buffers[0] = nil
	b.Set("nes", buffers)
	bytes, err = b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, schemeRoot)
	assert.Nil(t, b.GetByIndex("nes", 0))
	assert.NotNil(t, b.GetByIndex("nes", 1))

}

func TestCopyBytes(t *testing.T) {
	s := NewScheme()
	s.AddField("name", FieldTypeString, false)
	s.AddField("id", FieldTypeInt, false)
	s.AddArray("longs", FieldTypeLong, false)
	s.AddField("float", FieldTypeFloat, false)

	// initial
	b := NewBuffer(s)
	b.Set("name", "str")
	b.Set("id", 42)
	b.Set("longs", []int64{45, 46})
	b.Set("float", 0.123)
	bytes, err := b.ToBytes()
	if err != nil {
		t.Fatal(err)
	}

	// make buffer unmodified
	b = ReadBuffer(bytes, s)

	// force copy existing unmodified array
	bytes, err = b.ToBytes()
	b = ReadBuffer(bytes, s)
	assert.Equal(t, "str", b.Get("name"))
	assert.Equal(t, int32(42), b.Get("id"))
	arr := b.Get("longs").(*Array).GetAllIntf().([]int64)
	assert.Equal(t, 45, arr[0])
	assert.Equal(t, 46, arr[1])
	assert.Equal(t, 0.123, b.Get("float"))

}
