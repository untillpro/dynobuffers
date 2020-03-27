/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestBasicUsage(t *testing.T) {
	// Yaml representation of scheme
	var schemeYaml = `
name: string 
price: float
quantity: int
weight: long
`
	// Create Scheme from yaml representation
	s, err := YamlToScheme(schemeYaml)
	require.Nil(t, err)

	var b *Buffer
	var bytes []byte

	// Create new buffer from scheme
	{
		b = NewBuffer(s)
		b.Set("name", "cola")
		// We may NOT get it back yet
		require.Panics(t, func() { require.Equal(t, "cola", b.Get("name").(string)) })
		b.Set("price", float32(0.123))
		b.Set("quantity", int32(42))
		b.Set("unknownField", "some value") // Nothing happens here, nothing will be written to buffer
		bytes, err = b.ToBytes()
		require.Nil(t, err)
	}

	// Create from bytes
	{
		b = ReadBuffer(bytes, s)
		// Now we can Get fields
		require.Equal(t, "cola", b.Get("name").(string))
		require.Equal(t, float32(0.123), b.Get("price"))
		require.Equal(t, int32(42), b.Get("quantity"))
		// `unknownField` is set but does not exist in scheme, so it is not written
		require.Nil(t, b.Get("unknownField"))
		// `weight` field exists but not set
		require.Nil(t, b.Get("weight"))
	}

	// Modify values
	{
		b.Set("price", float32(0.124))
		b.Set("name", nil) // set to nil means `unset`
		bytes, err = b.ToBytes()
		require.Nil(t, err)
		b = ReadBuffer(bytes, s)
		require.Nil(t, b.Get("name"))
		require.Equal(t, float32(0.124), b.Get("price").(float32))
		require.Equal(t, int32(42), b.Get("quantity").(int32))
	}

	// set untyped int value
	{
		b.Set("quantity", 45)
		bytes, err = b.ToBytes()
		require.Nil(t, err)
		b = ReadBuffer(bytes, s)
		actual := b.Get("quantity")
		require.Equal(t, int32(45), actual.(int32))
	}

	// Use typed getter
	{
		_, ok := b.GetInt("unknownField")
		require.False(t, ok)
	}

	// check HasValue
	{
		b = NewBuffer(s)
		b.Set("price", float32(0.123))
		b.Set("quantity", nil)
		bytes, err = b.ToBytes()
		require.Nil(t, err, err)
		b = ReadBuffer(bytes, s)
		require.True(t, b.HasValue("price"))
		require.False(t, b.HasValue("quantity")) // set to nil
		require.False(t, b.HasValue("name"))     // not set
		require.False(t, b.HasValue("unknownField"))
	}

	// set string field to non-string -> error
	b.Set("name", 123)
	bytes, err = b.ToBytes()
	require.NotNil(t, err)

	// nil Scheme provided -> panic
	require.Panics(t, func() { NewBuffer(nil) })
	require.Panics(t, func() { ReadBuffer([]byte{}, nil) })
}

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
intsObj..:
  int: int
`

func TestNilFields(t *testing.T) {
	s, err := YamlToScheme(allTypesYaml)
	require.Nil(t, err)

	// test initially unset
	b := NewBuffer(s)
	testNilFields(t, b)
	bytes, err := b.ToBytes()
	require.Nil(t, err)
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
	require.Nil(t, err)
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
	require.Nil(t, err)
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
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testNilFields(t, b)
}

func testNilFields(t *testing.T, b *Buffer) {
	require.Nil(t, b.Get("int"))
	_, ok := b.GetInt("int")
	require.False(t, ok)
	require.Nil(t, b.Get("long"))
	_, ok = b.GetLong("long")
	require.False(t, ok)
	require.Nil(t, b.Get("float"))
	_, ok = b.GetFloat("float")
	require.False(t, ok)
	require.Nil(t, b.Get("double"))
	_, ok = b.GetDouble("double")
	require.False(t, ok)
	require.Nil(t, b.Get("string"))
	_, ok = b.GetString("string")
	require.False(t, ok)
	require.Nil(t, b.Get("boolFalse"))
	_, ok = b.GetBool("boolFalse")
	require.False(t, ok)
	require.Nil(t, b.Get("boolTrue"))
	_, ok = b.GetBool("boolTrue")
	require.False(t, ok)
	require.Nil(t, b.Get("byte"))
	_, ok = b.GetByte("byte")
	require.False(t, ok)
}

func TestWriteNewReadOld(t *testing.T) {
	schemeNew, err := YamlToScheme(schemeStrNew)
	require.Nil(t, err)
	b := NewBuffer(schemeNew)
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	b.Set("newField", int64(1))
	bytesNew, err := b.ToBytes()
	require.Nil(t, err)

	schemeOld, err := YamlToScheme(schemeStr)
	require.Nil(t, err)
	b = ReadBuffer(bytesNew, schemeOld)

	actual := b.Get("name")
	require.Equal(t, "cola", actual.(string))
	actual = b.Get("price")
	require.Equal(t, float32(0.123), actual.(float32))
	actual = b.Get("quantity")
	require.Equal(t, int32(42), actual.(int32))

	actual = b.Get("newField")
	require.Nil(t, actual)
}

func TestWriteOldReadNew(t *testing.T) {
	schemeOld, err := YamlToScheme(schemeStr)
	require.Nil(t, err)
	b := NewBuffer(schemeOld)
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	bytesOld, err := b.ToBytes()
	require.Nil(t, err)

	schemeNew, err := YamlToScheme(schemeStrNew)
	require.Nil(t, err)
	b = ReadBuffer(bytesOld, schemeNew)

	actual := b.Get("name")
	require.Equal(t, "cola", actual.(string))
	actual = b.Get("price")
	require.Equal(t, float32(0.123), actual.(float32))
	actual = b.Get("quantity")
	require.Equal(t, int32(42), actual.(int32))

	actual = b.Get("newField")
	require.Nil(t, actual)
}

func TestToBytesFilledUnmodified(t *testing.T) {
	b := getBufferAllFields(t, int32(1), int64(2), float32(3), float64(4), "str", byte(5))
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, b.Scheme)
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
	require.Nil(t, err)
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
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	return b
}

func testFieldValues(t *testing.T, b *Buffer, expectedInt32 int32, expectedInt64 int64, expectedFloat32 float32, expectedFloat64 float64, expectedString string, expectedByte byte) {
	actual := b.Get("int")
	actualInt, ok := b.GetInt("int")
	require.Equal(t, expectedInt32, actual)
	require.Equal(t, expectedInt32, actualInt)
	require.True(t, ok)

	actual = b.Get("long")
	actualLong, ok := b.GetLong("long")
	require.Equal(t, expectedInt64, actual)
	require.Equal(t, expectedInt64, actualLong)
	require.True(t, ok)

	actual = b.Get("float")
	actualFloat, ok := b.GetFloat("float")
	require.Equal(t, expectedFloat32, actual)
	require.Equal(t, expectedFloat32, actualFloat)
	require.True(t, ok)

	actual = b.Get("double")
	actualDouble, ok := b.GetDouble("double")
	require.Equal(t, expectedFloat64, actual)
	require.Equal(t, expectedFloat64, actualDouble)
	require.True(t, ok)

	actual = b.Get("string")
	actualString, ok := b.GetString("string")
	require.Equal(t, expectedString, actual)
	require.Equal(t, expectedString, actualString)
	require.True(t, ok)

	actual = b.Get("byte")
	actualByte, ok := b.GetByte("byte")
	require.Equal(t, expectedByte, actual)
	require.Equal(t, expectedByte, actualByte)
	require.True(t, ok)

	actual = b.Get("boolTrue")
	actualBool, ok := b.GetBool("boolTrue")
	require.Equal(t, true, actual)
	require.Equal(t, true, actualBool)
	require.True(t, ok)

	actual = b.Get("boolFalse")
	actualBool, ok = b.GetBool("boolFalse")
	require.Equal(t, false, actual)
	require.Equal(t, false, actualBool)
	require.True(t, ok)
}

func TestToJSONBasic(t *testing.T) {
	scheme, err := YamlToScheme(schemeStr)
	require.Nil(t, err)

	b := NewBuffer(scheme)
	dest := map[string]interface{}{}
	jsonBytes := b.ToJSON()
	json.Unmarshal(jsonBytes, &dest)
	require.True(t, len(dest) == 0)

	// basic test
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	jsonBytes = b.ToJSON()
	json.Unmarshal(jsonBytes, &dest)
	require.True(t, len(dest) == 3)
	require.Equal(t, "cola", dest["name"])
	require.Equal(t, float64(0.123), dest["price"])
	require.Equal(t, float64(42), dest["quantity"])

	// unmodified
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)
	jsonBytes = b.ToJSON()
	json.Unmarshal(jsonBytes, &dest)
	require.True(t, len(dest) == 3)
	require.Equal(t, "cola", dest["name"])
	require.Equal(t, float64(0.123), dest["price"])
	require.Equal(t, float64(42), dest["quantity"])

	// test field initially not set
	b = NewBuffer(scheme)
	b.Set("name", "cola")
	b.Set("quantity", int32(42))
	jsonBytes = b.ToJSON()
	dest = map[string]interface{}{}
	json.Unmarshal(jsonBytes, &dest)
	require.True(t, len(dest) == 2)
	require.Equal(t, "cola", dest["name"])
	require.Equal(t, float64(42), dest["quantity"])

	// test field not set on ReadBuffer
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)

	jsonBytes = b.ToJSON()
	dest = map[string]interface{}{}
	json.Unmarshal(jsonBytes, &dest)
	require.True(t, len(dest) == 2)
	require.Equal(t, "cola", dest["name"])
	require.Equal(t, float64(42), dest["quantity"])

	// test unset field
	b.Set("quantity", nil)
	jsonBytes = b.ToJSON()
	dest = map[string]interface{}{}
	json.Unmarshal(jsonBytes, &dest)
	require.True(t, len(dest) == 1)
	require.Equal(t, "cola", dest["name"])

	// test read unset field
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)
	jsonBytes = b.ToJSON()
	dest = map[string]interface{}{}
	json.Unmarshal(jsonBytes, &dest)
	require.True(t, len(dest) == 1)
	require.Equal(t, "cola", dest["name"])
}

func TestToJSONMapBasic(t *testing.T) {
	scheme, err := YamlToScheme(schemeStr)
	require.Nil(t, err)

	b := NewBuffer(scheme)
	dest := b.ToJSONMap()
	require.True(t, len(dest) == 0)

	// basic test
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	dest = b.ToJSONMap()
	require.True(t, len(dest) == 3)
	require.Equal(t, "cola", dest["name"])
	require.Equal(t, float32(0.123), dest["price"])
	require.Equal(t, int32(42), dest["quantity"])

	// unmodified
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)
	dest = b.ToJSONMap()
	require.True(t, len(dest) == 3)
	require.Equal(t, "cola", dest["name"])
	require.Equal(t, float32(0.123), dest["price"])
	require.Equal(t, int32(42), dest["quantity"])

	// test field initially not set
	b = NewBuffer(scheme)
	b.Set("name", "cola")
	b.Set("quantity", int32(42))
	dest = b.ToJSONMap()
	require.True(t, len(dest) == 2)
	require.Equal(t, "cola", dest["name"])
	require.Equal(t, int32(42), dest["quantity"])

	// test field not set on ReadBuffer
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)

	dest = b.ToJSONMap()
	require.True(t, len(dest) == 2)
	require.Equal(t, "cola", dest["name"])
	require.Equal(t, int32(42), dest["quantity"])

	// test unset field
	b.Set("quantity", nil)
	dest = b.ToJSONMap()
	require.True(t, len(dest) == 1)
	require.Equal(t, "cola", dest["name"])

	// test read unset field
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)
	dest = b.ToJSONMap()
	require.True(t, len(dest) == 1)
	require.Equal(t, "cola", dest["name"])
}

func TestToJSONMapNestedAndNestedArray(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nested", schemeNested, true)
	schemeRoot.AddNestedArray("nestedarr", schemeNested, false)
	schemeRoot.AddArray("ids", FieldTypeInt, false)
	schemeRoot.AddArray("bytes", FieldTypeByte, false)

	bRoot := NewBuffer(schemeRoot)
	bytes, err := bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str1", "nested": {"price": 0.123,"quantity":42},"nestedarr":[{"price": 0.124,"quantity":43},{"price": 0.125,"quantity":44}], "ids": [45,46]}`))
	require.Nil(t, err)
	
	m := bRoot.ToJSONMap()
	require.Equal(t, "str1", m["name"])
	nested := m["nested"].(map[string]interface{})
	require.Equal(t, float64(0.123), nested["price"])
	require.Equal(t, float64(42), nested["quantity"])
	nestedArr := m["nestedarr"].([]interface{})
	nesteArrElem := nestedArr[0].(map[string]interface{})
	require.Equal(t, float64(0.124), nesteArrElem["price"])
	require.Equal(t, float64(43), nesteArrElem["quantity"])
	require.Len(t, nesteArrElem, 2)
	nesteArrElem = nestedArr[1].(map[string]interface{})
	require.Equal(t, float64(0.125), nesteArrElem["price"])
	require.Equal(t, float64(44), nesteArrElem["quantity"])
	require.Len(t, nesteArrElem, 2)
	require.Len(t, nestedArr, 2)
	require.Equal(t, []interface{}{float64(45), float64(46)}, m["ids"])

	bRoot = ReadBuffer(bytes, schemeRoot)
	m = bRoot.ToJSONMap()
	require.Equal(t, "str1", m["name"])
	nested = m["nested"].(map[string]interface{})
	require.Equal(t, float32(0.123), nested["price"])
	require.Equal(t, int32(42), nested["quantity"])
	nestedArr = m["nestedarr"].([]interface{})
	nesteArrElem = nestedArr[0].(map[string]interface{})
	require.Equal(t, float32(0.124), nesteArrElem["price"])
	require.Equal(t, int32(43), nesteArrElem["quantity"])
	require.Len(t, nesteArrElem, 2)
	nesteArrElem = nestedArr[1].(map[string]interface{})
	require.Equal(t, float32(0.125), nesteArrElem["price"])
	require.Equal(t, int32(44), nesteArrElem["quantity"])
	require.Len(t, nesteArrElem, 2)
	require.Len(t, nestedArr, 2)
	require.Equal(t, []int32{45, 46}, m["ids"])
}

func TestDifferentOrder(t *testing.T) {
	s, err := YamlToScheme(schemeStr)
	require.Nil(t, err)
	b := NewBuffer(s)
	b.Set("quantity", int32(42))
	b.Set("Name", "cola")
	b.Set("price", float32(0.123))
	bytes, err := b.ToBytes()
	require.Nil(t, err)

	b = ReadBuffer(bytes, s)
	b.Set("price", float32(0.124))
	b.Set("name", "new cola")
	bytes, err = b.ToBytes()
	require.Nil(t, err)

	b = ReadBuffer(bytes, s)
	actual := b.Get("name")
	require.Equal(t, "new cola", actual.(string))
	actual = b.Get("price")
	require.Equal(t, float32(0.124), actual.(float32))
	actual = b.Get("quantity")
	require.Equal(t, int32(42), actual.(int32))
}

func TestSchemeToFromYAML(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeNested.Name = "nes"
	schemeNestedArr := NewScheme()
	schemeNestedArr.AddField("price", FieldTypeFloat, false)
	schemeNestedArr.AddField("quantity", FieldTypeInt, true)
	schemeNestedArr.Name = "nesarr"
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nes", schemeNested, true)
	schemeRoot.AddNestedArray("nesarr", schemeNestedArr, true)
	schemeRoot.AddField("last", FieldTypeInt, false)
	bytes, err := yaml.Marshal(schemeRoot)
	schemeNew := NewScheme()
	err = yaml.Unmarshal(bytes, &schemeNew)
	require.Nil(t, err)

	pretty.CompareConfig.TrackCycles = true

	if !reflect.DeepEqual(schemeRoot.Fields, schemeNew.Fields) {
		t.Fatal(pretty.Compare(schemeRoot.Fields, schemeNew.Fields))
	}
	if !reflect.DeepEqual(schemeRoot.FieldsMap, schemeNew.FieldsMap) {
		t.Fatal(pretty.Compare(schemeRoot.FieldsMap, schemeNew.FieldsMap))
	}

	schemeNew = NewScheme()
	err = yaml.Unmarshal([]byte("wrong "), &schemeNew)
	require.NotNil(t, err)

	_, err = YamlToScheme("wrong yaml")
	require.NotNil(t, err)
	_, err = YamlToScheme("name: wrongType")
	require.NotNil(t, err)
	_, err = YamlToScheme(`
		nested:
		  nestedField: wrongType`)
	require.NotNil(t, err)
	_, err = YamlToScheme(`
		nested:
		  wrong`)
	require.NotNil(t, err)
}

func TestMandatoryFields(t *testing.T) {
	scheme, err := YamlToScheme(schemeMandatory)
	require.Nil(t, err)
	b := NewBuffer(scheme)
	bytes, err := b.ToBytes()
	require.NotNil(t, err)
	require.Nil(t, bytes)

	b.Set("name", "str")
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
	require.Nil(t, bytes)

	b.Set("price", 0.123)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	require.NotNil(t, bytes)

	b = ReadBuffer(bytes, scheme)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	require.NotNil(t, bytes)
}

func TestApplyJsonErrors(t *testing.T) {
	scheme, err := YamlToScheme(schemeMandatory)
	require.Nil(t, err)
	b := NewBuffer(scheme)

	// unset field
	json := []byte(`{"name": null, "price": 0.124}`)
	bytes, err := b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)
	require.Nil(t, b.Get("name"))
	require.Equal(t, float32(0.124), b.Get("price"))

	// wrong type -> error
	json = []byte(`{"name": "str", "price": "wrong type"}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, bytes)
	require.NotNil(t, err)

	// unset mandatory field -> error
	json = []byte(`{"name": "str", "price": null}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, bytes)
	require.NotNil(t, err)

	// mandatory field is not set -> error
	json = []byte(`{"name": "str"}`)
	b = NewBuffer(scheme)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, bytes)
	require.NotNil(t, err)

	// wrong json -> error
	json = []byte(`wrong`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, bytes)
	require.NotNil(t, err)

	// non-object is provided -> error
	s := NewScheme()
	sNested := NewScheme()
	s.AddNested("nes", sNested, false)
	b = NewBuffer(s)
	bytes, err = b.ApplyJSONAndToBytes([]byte(`{"nes": 42}`))
	require.Nil(t, bytes)
	require.NotNil(t, err)

	// unknown field -> error
	json = []byte(`{"price": 0.124, "unknown": 42}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, bytes)
	require.NotNil(t, err)
}

func TestApplyJSONNestedAndNestedArray(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nested", schemeNested, true)
	schemeRoot.AddNestedArray("nestedarr", schemeNested, false)
	schemeRoot.AddArray("ids", FieldTypeInt, false)
	schemeRoot.AddArray("bytes", FieldTypeByte, false)

	bRoot := NewBuffer(schemeRoot)
	bytes, err := bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str1", "nested": {"price": 0.123,"quantity":42},"nestedarr":[{"price": 0.124,"quantity":43},{"price": 0.125,"quantity":44}], "ids": [45,46]}`))
	require.Nil(t, err)
	bRoot = ReadBuffer(bytes, schemeRoot)
	require.Equal(t, "str1", bRoot.Get("name"))
	ints := bRoot.Get("ids")
	require.Equal(t, []int32{45, 46}, ints)
	bRoot.Set("ids", ints)
	bytes, err = bRoot.ToBytes()
	require.Nil(t, err)
	bRoot = ReadBuffer(bytes, schemeRoot)
	ints = bRoot.Get("ids")
	require.Equal(t, []int32{45, 46}, ints)
	bNested := bRoot.Get("nested").(*Buffer)
	require.Equal(t, float32(0.123), bNested.Get("price"))
	require.Equal(t, int32(42), bNested.Get("quantity"))
	bNesteds := bRoot.Get("nestedarr").(*ObjectArray)
	bNesteds.Next()
	require.Equal(t, float32(0.124), bNesteds.Buffer.Get("price"))
	require.Equal(t, int32(43), bNesteds.Buffer.Get("quantity"))
	bNesteds.Next()
	require.Equal(t, float32(0.125), bNesteds.Buffer.Get("price"))
	require.Equal(t, int32(44), bNesteds.Buffer.Get("quantity"))
	require.False(t, bNesteds.Next())

	// error if not array is provided for an array field
	bytes, err = bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str","nestedarr":{"price":0.123,"quantity":42},"last":42}`))
	require.NotNil(t, err)

	// error if not object is provided as an array element
	bytes, err = bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str","nestedarr":[0.123],"last":42}`))
	require.NotNil(t, err)

	// error if wrong failed to encode an array
	bytes, err = bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str","nestedarr":[{"price":0.123,"quantity":"wrong value"}],"last":42}`))
	require.NotNil(t, err)

	// error if non-base64 string is provided for byte array field
	bytes, err = bRoot.ApplyJSONAndToBytes([]byte(`{"bytes":[1]}`))
	require.NotNil(t, err)

	// error if wrong base64 string is provided for byte array field
	bytes, err = bRoot.ApplyJSONAndToBytes([]byte(`{"bytes":"wrong base64"}`))
	require.NotNil(t, err)

	// error if mandatory nested object is null
	bRoot = NewBuffer(schemeRoot)
	_, err = bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str","nested":null}`))
	require.NotNil(t, err)
	bRoot = NewBuffer(schemeRoot)
	_, err = bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str"}`))
	require.NotNil(t, err)

	// error if mandatory field in nested object is null
	bRoot = NewBuffer(schemeRoot)
	_, err = bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str","nested":{"price": 1,"quantity":null}}`))
	require.NotNil(t, err)
	bRoot = NewBuffer(schemeRoot)
	_, err = bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str","nested":{"price": 1}}`))
	require.NotNil(t, err)
	bRoot = NewBuffer(schemeRoot)
	_, err = bRoot.ApplyJSONAndToBytes([]byte(`{"name":"str","nes":{}}`))
	require.NotNil(t, err)
}

func TestApplyJSONArraysAllTypesAppend(t *testing.T) {
	s, err := YamlToScheme(arraysAllTypesYaml)
	require.Nil(t, err)
	b := NewBuffer(s)
	bytes, err := b.ApplyJSONAndToBytes([]byte(`{"ints":[1,2],"longs":[3,4],"floats":[0.123,0.124],"doubles":[0.125,0.126],"strings":["str1","str2"],"boolTrues":[true,false],"boolFalses":[false,true],"bytes":"BQY=","intsObj":[{"int":7},{"int":8}]}`))
	require.Nil(t, err)

	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4}, b.Get("longs"))
	require.Equal(t, []float32{0.123, 0.124}, b.Get("floats"))
	require.Equal(t, []float64{0.125, 0.126}, b.Get("doubles"))
	require.Equal(t, []byte{5, 6}, b.Get("bytes"))
	require.Equal(t, []bool{true, false}, b.Get("boolTrues"))
	require.Equal(t, []bool{false, true}, b.Get("boolFalses"))
	require.Equal(t, []string{"str1", "str2"}, b.Get("strings"))
	buffs := b.Get("intsObj").(*ObjectArray)
	buffs.Next()
	require.Equal(t, int32(7), buffs.Buffer.Get("int"))
	buffs.Next()
	require.Equal(t, int32(8), buffs.Buffer.Get("int"))
	require.False(t, buffs.Next())

	bytes, err = b.ApplyJSONAndToBytes([]byte(`{"ints":[-1,-2],"longs":[-3,-4],"floats":[-0.123,-0.124],"doubles":[-0.125,-0.126],"strings":["","str4"],"boolTrues":[true,true],"boolFalses":[false,false],"bytes":"BQY=","intsObj":[{"int":-7},{"int":-8}]}`))
	require.Nil(t, err)

	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2, -1, -2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4, -3, -4}, b.Get("longs"))
	require.Equal(t, []float32{0.123, 0.124, -0.123, -0.124}, b.Get("floats"))
	require.Equal(t, []float64{0.125, 0.126, -0.125, -0.126}, b.Get("doubles"))
	require.Equal(t, []byte{5, 6, 5, 6}, b.Get("bytes"))
	require.Equal(t, []bool{true, false, true, true}, b.Get("boolTrues"))
	require.Equal(t, []bool{false, true, false, false}, b.Get("boolFalses"))
	require.Equal(t, []string{"str1", "str2", "", "str4"}, b.Get("strings"))
	buffs = b.Get("intsObj").(*ObjectArray)
	buffs.Next()
	require.Equal(t, int32(7), buffs.Buffer.Get("int"))
	buffs.Next()
	require.Equal(t, int32(8), buffs.Buffer.Get("int"))
	buffs.Next()
	require.Equal(t, int32(-7), buffs.Buffer.Get("int"))
	buffs.Next()
	require.Equal(t, int32(-8), buffs.Buffer.Get("int"))
	require.False(t, buffs.Next())
}

func TestApplyJSONArraysAllTypesSet(t *testing.T) {
	s, err := YamlToScheme(arraysAllTypesYaml)
	require.Nil(t, err)

	// ints
	// element of wrong type -> error
	json := []byte(`{"ints": [1, 0.123]}`)
	b := NewBuffer(s)
	_, err = b.ApplyJSONAndToBytes(json)
	require.NotNil(t, err)
	json = []byte(`{"ints": [1, 2]}`)
	b = NewBuffer(s)
	bytes, err := b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2}, b.Get("ints"))

	// longs
	// element of wrong type -> error
	json = []byte(`{"longs": [1, "str"]}`)
	_, err = b.ApplyJSONAndToBytes(json)
	require.NotNil(t, err)
	json = []byte(`{"longs": [3, 4]}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4}, b.Get("longs"))

	// floats
	// element of wrong type -> error
	json = []byte(`{"floats": [1, "str"]}`)
	_, err = b.ApplyJSONAndToBytes(json)
	require.NotNil(t, err)
	json = []byte(`{"floats": [0.123, 0.124]}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4}, b.Get("longs"))
	require.Equal(t, []float32{0.123, 0.124}, b.Get("floats"))

	// doubles
	// element of wrong type -> error
	json = []byte(`{"doubles": [0.125, "str"]}`)
	_, err = b.ApplyJSONAndToBytes(json)
	require.NotNil(t, err)
	json = []byte(`{"doubles": [0.125, 0.126]}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4}, b.Get("longs"))
	require.Equal(t, []float32{0.123, 0.124}, b.Get("floats"))
	require.Equal(t, []float64{0.125, 0.126}, b.Get("doubles"))

	// bytes
	bytesBase64 := base64.StdEncoding.EncodeToString([]byte{5, 6})
	json = []byte(fmt.Sprintf(`{"bytes": "%s"}`, bytesBase64))
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4}, b.Get("longs"))
	require.Equal(t, []float32{0.123, 0.124}, b.Get("floats"))
	require.Equal(t, []float64{0.125, 0.126}, b.Get("doubles"))
	require.Equal(t, []byte{5, 6}, b.Get("bytes"))

	// bools
	// element of wrong type -> error
	json = []byte(`{"boolTrues": ["str"]}`)
	_, err = b.ApplyJSONAndToBytes(json)
	require.NotNil(t, err)
	json = []byte(`{"boolTrues": [true, false], "boolFalses": [false, true]}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4}, b.Get("longs"))
	require.Equal(t, []float32{0.123, 0.124}, b.Get("floats"))
	require.Equal(t, []float64{0.125, 0.126}, b.Get("doubles"))
	require.Equal(t, []byte{5, 6}, b.Get("bytes"))
	require.Equal(t, []bool{true, false}, b.Get("boolTrues"))
	require.Equal(t, []bool{false, true}, b.Get("boolFalses"))

	// strings
	// element of wrong type -> error
	json = []byte(`{"strings": ["str1", 1]}`)
	_, err = b.ApplyJSONAndToBytes(json)
	require.NotNil(t, err)
	json = []byte(`{"strings": ["str1", "str2"]}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4}, b.Get("longs"))
	require.Equal(t, []float32{0.123, 0.124}, b.Get("floats"))
	require.Equal(t, []float64{0.125, 0.126}, b.Get("doubles"))
	require.Equal(t, []byte{5, 6}, b.Get("bytes"))
	require.Equal(t, []bool{true, false}, b.Get("boolTrues"))
	require.Equal(t, []bool{false, true}, b.Get("boolFalses"))
	require.Equal(t, []string{"str1", "str2"}, b.Get("strings"))

	// nested
	json = []byte(`{"intsObj": [{"int": 7}, {"int": 8}]}`)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4}, b.Get("longs"))
	require.Equal(t, []float32{0.123, 0.124}, b.Get("floats"))
	require.Equal(t, []float64{0.125, 0.126}, b.Get("doubles"))
	require.Equal(t, []byte{5, 6}, b.Get("bytes"))
	require.Equal(t, []bool{true, false}, b.Get("boolTrues"))
	require.Equal(t, []bool{false, true}, b.Get("boolFalses"))
	require.Equal(t, []string{"str1", "str2"}, b.Get("strings"))
	buffs := b.Get("intsObj").(*ObjectArray)
	buffs.Next()
	require.Equal(t, int32(7), buffs.Buffer.Get("int"))
	buffs.Next()
	require.Equal(t, int32(8), buffs.Buffer.Get("int"))
	require.False(t, buffs.Next())

	// buffer -> json -> buffer
	json = b.ToJSON()
	b = NewBuffer(s)
	bytes, err = b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{1, 2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4}, b.Get("longs"))
	require.Equal(t, []float32{0.123, 0.124}, b.Get("floats"))
	require.Equal(t, []float64{0.125, 0.126}, b.Get("doubles"))
	require.Equal(t, []byte{5, 6}, b.Get("bytes"))
	require.Equal(t, []bool{true, false}, b.Get("boolTrues"))
	require.Equal(t, []bool{false, true}, b.Get("boolFalses"))
	require.Equal(t, []string{"str1", "str2"}, b.Get("strings"))
	buffs = b.Get("intsObj").(*ObjectArray)
	buffs.Next()
	require.Equal(t, int32(7), buffs.Buffer.Get("int"))
	buffs.Next()
	require.Equal(t, int32(8), buffs.Buffer.Get("int"))
	require.False(t, buffs.Next())

	// empty arrays
	b.Set("ints", []int32{})
	b.Set("longs", []int64{})
	b.Set("floats", []float32{})
	b.Set("doubles", []float64{})
	b.Set("boolTrues", []bool{})
	b.Set("boolFalses", []bool{})
	b.Set("bytes", []byte{})
	b.Set("strings", []string{})
	b.Set("intsObj", []*Buffer{})
	bytes, err = b.ToBytes()
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{}, b.Get("ints"))
	require.Equal(t, []int64{}, b.Get("longs"))
	require.Equal(t, []float32{}, b.Get("floats"))
	require.Equal(t, []float64{}, b.Get("doubles"))
	require.Equal(t, []byte{}, b.Get("bytes"))
	require.Equal(t, []bool{}, b.Get("boolTrues"))
	require.Equal(t, []bool{}, b.Get("boolFalses"))
	require.Equal(t, []string{}, b.Get("strings"))
	require.Equal(t, 0, b.Get("intsObj").(*ObjectArray).Len)

	// empty arrays from json
	b = NewBuffer(s)
	if bytes, err = b.ApplyJSONAndToBytes([]byte(`{"ints":[],"longs":[],"floats":[],"doubles":[],"strings":[],"boolTrues":[],"boolFalses":[],"bytes":"","intsObj":[]}`)); err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, s)
	require.Equal(t, []int32{}, b.Get("ints"))
	require.Equal(t, []int64{}, b.Get("longs"))
	require.Equal(t, []float32{}, b.Get("floats"))
	require.Equal(t, []float64{}, b.Get("doubles"))
	require.Equal(t, []byte{}, b.Get("bytes"))
	require.Equal(t, []bool{}, b.Get("boolTrues"))
	require.Equal(t, []bool{}, b.Get("boolFalses"))
	require.Equal(t, []string{}, b.Get("strings"))
	require.Equal(t, 0, b.Get("intsObj").(*ObjectArray).Len)

	// null arrays from json
	b = NewBuffer(s)
	if bytes, err = b.ApplyJSONAndToBytes([]byte(`{"ints":null,"longs":null,"floats":null,"doubles":null,"strings":null,"boolTrues":null,"boolFalses":null,"bytes":null,"intsObj":null}`)); err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytes, s)
	require.Nil(t, b.Get("ints"))
	require.Nil(t, b.Get("longs"))
	require.Nil(t, b.Get("floats"))
	require.Nil(t, b.Get("doubles"))
	require.Nil(t, b.Get("bytes"))
	require.Nil(t, b.Get("boolTrues"))
	require.Nil(t, b.Get("boolFalses"))
	require.Nil(t, b.Get("strings"))
	require.Nil(t, b.Get("intsObj"))
}

func TestApplyJsonPrimitiveAllTypes(t *testing.T) {
	scheme, err := YamlToScheme(allTypesYaml)
	require.Nil(t, err)
	json := []byte(`{"string": "str", "long": 42, "int": 43, "float": 0.124, "double": 0.125, "byte": 6, "boolTrue": true, "boolFalse": false}`)
	b := NewBuffer(scheme)
	bytes, err := b.ApplyJSONAndToBytes(json)
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)
	testFieldValues(t, b, int32(43), int64(42), float32(0.124), float64(0.125), "str", 6)
	require.True(t, b.Get("boolTrue").(bool))
	require.False(t, b.Get("boolFalse").(bool))
}

func TestNestedBasic(t *testing.T) {
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot := NewScheme()
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nes", schemeNested, false)
	schemeRoot.AddField("last", FieldTypeInt, false)

	// initially nil
	b := NewBuffer(schemeRoot)
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	require.Nil(t, b.Get("name"))
	require.Nil(t, b.Get("nes"))
	require.Nil(t, b.Get("last"))

	// nested still nil after Get() and not modify
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, schemeRoot)
	require.Nil(t, b.Get("nes"))

	// fill
	bNested := NewBuffer(schemeNested)
	bNested.Set("price", 0.123)
	bNested.Set("quantity", 42)
	b.Set("name", "str")
	b.Set("nes", bNested)
	b.Set("last", 42)
	bytes, err = b.ToBytes()
	require.Nil(t, err)

	b = ReadBuffer(bytes, schemeRoot)
	require.Equal(t, "str", b.Get("name"))
	require.Equal(t, int32(42), b.Get("last"))

	bNested = b.Get("nes").(*Buffer)
	require.Equal(t, int32(42), bNested.Get("quantity"))
	require.Equal(t, float32(0.123), bNested.Get("price"))
}

func TestNestedAdvanced(t *testing.T) {
	schemeRoot := NewScheme()
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot.AddField("name", FieldTypeString, false)
	schemeRoot.AddNested("nes", schemeNested, true)
	schemeRoot.AddField("last", FieldTypeInt, false)
	b := NewBuffer(schemeRoot)

	// fill
	bNested := NewBuffer(schemeNested)
	bNested.Set("price", 0.123)
	bNested.Set("quantity", 42)
	b.Set("name", "str")
	b.Set("last", 42)

	// error if mandatory object is not set
	bytes, err := b.ToBytes()
	require.NotNil(t, err)
	b.Set("nes", nil)
	bytes, err = b.ToBytes()
	require.NotNil(t, err)

	// set mandatory object
	b.Set("nes", bNested)
	bytes, err = b.ToBytes()
	require.Nil(t, err)

	// modify nested
	b = ReadBuffer(bytes, schemeRoot)
	bNested = b.Get("nes").(*Buffer)
	bNested.Set("quantity", 43)
	bNested.Set("price", 0.124)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, schemeRoot)
	require.Equal(t, "str", b.Get("name"))
	require.Equal(t, int32(42), b.Get("last"))
	bNested = b.Get("nes").(*Buffer)
	require.Equal(t, int32(43), bNested.Get("quantity"))
	require.Equal(t, float32(0.124), bNested.Get("price"))

	// non-*Buffer is provided -> error
	b.Set("nes", []int32{0, 1})
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
	b.Set("nes", bNested)

	// error if unset mandatory in nested
	bNested.Set("quantity", nil)
	bytes, err = b.ToBytes()
	require.NotNil(t, err)

	// error if unset nested
	b.Set("nes", nil)
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
}

func TestArraysBasic(t *testing.T) {
	s := NewScheme()
	s.AddArray("longs", FieldTypeLong, false)

	// initial
	b := NewBuffer(s)
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Nil(t, b.Get("longs"))

	// set and read
	longs := []int64{5, 6}
	b.Set("longs", longs)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, int64(5), b.GetByIndex("longs", 0))
	require.Nil(t, b.GetByIndex("longs", 3))
	require.Equal(t, int64(6), b.GetByIndex("longs", 1))
	require.Nil(t, b.GetByIndex("longs", -1))
	require.Nil(t, b.GetByIndex("unexisting", 0))

	// Non-modified array should be copied on ToBytes()
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, int64(5), b.GetByIndex("longs", 0))
	require.Equal(t, int64(6), b.GetByIndex("longs", 1))

	//test Array struct
	longsActual := b.Get("longs").([]int64)
	require.Equal(t, longs, longsActual)

	// modify
	longs = []int64{7, 8, 9}
	b.Set("longs", longs)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, int64(7), b.GetByIndex("longs", 0))
	require.Equal(t, int64(8), b.GetByIndex("longs", 1))

	// unset
	b.Set("longs", nil)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Nil(t, b.Get("longs"))

	// set to empty
	longs = []int64{}
	b.Set("longs", longs)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Nil(t, b.GetByIndex("longs", 0))
	longsActual = b.Get("longs").([]int64)
	require.True(t, len(longsActual) == 0)
}

func TestArraysAllTypesSet(t *testing.T) {
	s, err := YamlToScheme(arraysAllTypesYaml)
	require.Nil(t, err)

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
	schemeNested := s.GetNestedScheme("intsObj")

	bNestedArr := []*Buffer{}
	bNested := NewBuffer(schemeNested)
	bNested.Set("int", 55)
	bNestedArr = append(bNestedArr, bNested)
	bNested = NewBuffer(schemeNested)
	bNested.Set("int", 56)
	bNestedArr = append(bNestedArr, bNested)
	b.Set("intsObj", bNestedArr)
	bytes, err := b.ToBytes()
	require.Nil(t, err)

	b = ReadBuffer(bytes, s)
	require.Equal(t, int32(1), b.GetByIndex("ints", 0))
	require.Equal(t, int32(2), b.GetByIndex("ints", 1))
	require.Equal(t, int64(3), b.GetByIndex("longs", 0))
	require.Equal(t, int64(4), b.GetByIndex("longs", 1))
	require.Equal(t, float32(0.5), b.GetByIndex("floats", 0))
	require.Equal(t, float32(0.6), b.GetByIndex("floats", 1))
	require.Equal(t, float64(0.7), b.GetByIndex("doubles", 0))
	require.Equal(t, float64(0.8), b.GetByIndex("doubles", 1))
	require.Equal(t, true, b.GetByIndex("boolTrues", 0))
	require.Equal(t, false, b.GetByIndex("boolTrues", 1))
	require.Equal(t, false, b.GetByIndex("boolFalses", 0))
	require.Equal(t, true, b.GetByIndex("boolFalses", 1))
	require.Equal(t, byte(9), b.GetByIndex("bytes", 0))
	require.Equal(t, byte(10), b.GetByIndex("bytes", 1))
	require.Equal(t, "str1", b.GetByIndex("strings", 0))
	require.Equal(t, "str2", b.GetByIndex("strings", 1))
	require.Equal(t, int32(55), b.GetByIndex("intsObj", 0).(*Buffer).Get("int"))
	require.Equal(t, int32(56), b.GetByIndex("intsObj", 1).(*Buffer).Get("int"))

	require.Equal(t, ints, b.Get("ints"))
	require.Equal(t, longs, b.Get("longs"))
	require.Equal(t, floats, b.Get("floats"))
	require.Equal(t, doubles, b.Get("doubles"))
	require.Equal(t, trueBools, b.Get("boolTrues"))
	require.Equal(t, falseBools, b.Get("boolFalses"))
	require.Equal(t, bytesArr, b.Get("bytes"))
	require.Equal(t, strings, b.Get("strings"))

	require.Nil(t, b.Get("unknown"))

	// wrong type provided -> error
	b.Set("ints", []float32{0})
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
	b.Set("ints", ints)

	b.Set("longs", []float32{0})
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
	b.Set("longs", longs)

	b.Set("floats", []int32{0})
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
	b.Set("floats", floats)

	b.Set("doubles", []int32{0})
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
	b.Set("doubles", doubles)

	b.Set("boolTrues", []int32{0})
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
	b.Set("boolTrues", trueBools)

	b.Set("boolFalses", []int32{0})
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
	b.Set("boolFalses", falseBools)

	b.Set("bytes", []int32{0})
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
	b.Set("bytes", bytes)

	b.Set("strings", []int32{0})
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
	b.Set("strings", strings)
}

func TestArraysAllTypesAppend(t *testing.T) {
	s, err := YamlToScheme(arraysAllTypesYaml)
	require.Nil(t, err)

	ints := []int32{1, 2}
	longs := []int64{3, 4}
	floats := []float32{0.5, 0.6}
	doubles := []float64{0.7, 0.8}
	trueBools := []bool{true, false}
	falseBools := []bool{false, true}
	bytesArr := []byte{9, 10}
	strings := []string{"str1", "str2"}

	// initial, arrays are nil -> equivalent to Set()
	b := NewBuffer(s)
	b.Append("ints", ints)
	b.Append("longs", longs)
	b.Append("floats", floats)
	b.Append("doubles", doubles)
	b.Append("boolTrues", trueBools)
	b.Append("boolFalses", falseBools)
	b.Append("bytes", bytesArr)
	b.Append("strings", strings)
	schemeNested := s.GetNestedScheme("intsObj")

	bNestedArr := []*Buffer{}
	bNested := NewBuffer(schemeNested)
	bNested.Set("int", 55)
	bNestedArr = append(bNestedArr, bNested)
	bNested = NewBuffer(schemeNested)
	bNested.Set("int", 56)
	bNestedArr = append(bNestedArr, bNested)
	b.Append("intsObj", bNestedArr)
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)

	require.Equal(t, ints, b.Get("ints"))
	require.Equal(t, longs, b.Get("longs"))
	require.Equal(t, floats, b.Get("floats"))
	require.Equal(t, doubles, b.Get("doubles"))
	require.Equal(t, trueBools, b.Get("boolTrues"))
	require.Equal(t, falseBools, b.Get("boolFalses"))
	require.Equal(t, bytesArr, b.Get("bytes"))
	require.Equal(t, strings, b.Get("strings"))
	require.Equal(t, int32(55), b.GetByIndex("intsObj", 0).(*Buffer).Get("int"))
	require.Equal(t, int32(56), b.GetByIndex("intsObj", 1).(*Buffer).Get("int"))

	// append to existing
	b.Append("ints", []int32{-1, -2})
	b.Append("longs", []int64{-3, -4})
	b.Append("floats", []float32{-0.5, -0.6})
	b.Append("doubles", []float64{-0.7, -0.8})
	b.Append("boolTrues", []bool{true, true})
	b.Append("boolFalses", []bool{false, false})
	b.Append("bytes", []byte{11, 12})
	b.Append("strings", []string{"", "str4"})
	bNestedArr = []*Buffer{}
	bNested = NewBuffer(schemeNested)
	bNested.Set("int", 60)
	bNestedArr = append(bNestedArr, bNested)
	b.Append("intsObj", bNestedArr)

	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)

	require.Equal(t, []int32{1, 2, -1, -2}, b.Get("ints"))
	require.Equal(t, []int64{3, 4, -3, -4}, b.Get("longs"))
	require.Equal(t, []float32{0.5, 0.6, -0.5, -0.6}, b.Get("floats"))
	require.Equal(t, []float64{0.7, 0.8, -0.7, -0.8}, b.Get("doubles"))
	require.Equal(t, []bool{true, false, true, true}, b.Get("boolTrues"))
	require.Equal(t, []bool{false, true, false, false}, b.Get("boolFalses"))
	require.Equal(t, []byte{9, 10, 11, 12}, b.Get("bytes"))
	require.Equal(t, []string{"str1", "str2", "", "str4"}, b.Get("strings"))
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
	require.Nil(t, err)

	// GetByIndex
	b = ReadBuffer(bytes, schemeRoot)
	bNested = b.GetByIndex("nes", 0).(*Buffer)
	require.Equal(t, int32(42), bNested.Get("quantity"))
	require.Equal(t, float32(0.123), bNested.Get("price"))
	bNested = b.GetByIndex("nes", 1).(*Buffer)
	require.Equal(t, int32(43), bNested.Get("quantity"))
	require.Equal(t, float32(0.124), bNested.Get("price"))

	// iterator
	arr := b.Get("nes").(*ObjectArray)
	arr.Next()
	require.Equal(t, int32(42), arr.Buffer.Get("quantity"))
	require.Equal(t, float32(0.123), arr.Buffer.Get("price"))
	arr.Next()
	require.Equal(t, int32(43), arr.Value().(*Buffer).Get("quantity"))
	require.Equal(t, float32(0.124), arr.Value().(*Buffer).Get("price"))
	require.False(t, arr.Next())

	// non-[]*Buffer is provided -> error
	b.Set("nes", []int32{1})
	bytes, err = b.ToBytes()
	require.NotNil(t, err)

	// set one of elements to nil -> error
	buffers := []*Buffer{nil, nil, nil}
	b.Set("nes", buffers)
	bytes, err = b.ToBytes()
	require.NotNil(t, err)
}

func TestNilArrayOfBytes(t *testing.T) {
	s := NewScheme()
	s.AddArray("bytes", FieldTypeByte, false)
	b := NewBuffer(s)
	var x []byte
	x = nil
	b.Set("bytes", x)
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Nil(t, b.Get("bytes"))
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
	require.Nil(t, err)

	// make buffer unmodified
	b = ReadBuffer(bytes, s)

	// force copy existing unmodified array
	bytes, err = b.ToBytes()
	b = ReadBuffer(bytes, s)
	require.Equal(t, "str", b.Get("name"))
	require.Equal(t, int32(42), b.Get("id"))
	arr := b.Get("longs").([]int64)
	require.Equal(t, int64(45), arr[0])
	require.Equal(t, int64(46), arr[1])
	require.Equal(t, int64(45), b.GetByIndex("longs", 0))
	require.Equal(t, float32(0.123), b.Get("float"))
}

func TestObjectAsBytesInv(t *testing.T) {
	bl := flatbuffers.NewBuilder(0)
	strOffset := bl.CreateString("nes1")
	bl.StartObject(2)
	bl.PrependUOffsetTSlot(0, strOffset, 0)
	bl.PrependInt32(42)
	no := bl.EndObject()
	// bl.PrependInt32(int32(len(bl.Bytes)))
	bl.Finish(no)
	bytesNested := bl.FinishedBytes()

	bl = flatbuffers.NewBuilder(0)
	strOffsetOut := bl.CreateString("out")
	strOffsetOut2 := bl.CreateString("out2")
	nestedObj := bl.CreateByteVector(bytesNested) // bytesNested will be corrupted here if bl.Reset() is used instead of bl = flatbuffers.NewBuilder(0)
	bl.StartObject(5)
	bl.PrependInt32(43)
	bl.PrependUOffsetTSlot(1, strOffsetOut, 0)
	bl.PrependUOffsetTSlot(2, nestedObj, 0)
	bl.PrependInt32(44)
	bl.PrependUOffsetTSlot(4, strOffsetOut2, 0)
	bl.Finish(bl.EndObject())

	bytesOut := bl.FinishedBytes()

	outTab := &flatbuffers.Table{}
	outTab.Bytes = bytesOut
	outTab.Pos = flatbuffers.GetUOffsetT(bytesOut)

	nestedOffset := flatbuffers.UOffsetT(outTab.Offset(flatbuffers.VOffsetT((2+2)*2))) + outTab.Pos
	nestedTab := &flatbuffers.Table{}
	nestedTab.Bytes = outTab.ByteVector(nestedOffset)
	require.True(t, reflect.DeepEqual(bytesNested, nestedTab.Bytes))
	nestedTab.Pos = flatbuffers.GetUOffsetT(nestedTab.Bytes)

	nestedStrOffset := flatbuffers.UOffsetT(nestedTab.Offset(flatbuffers.VOffsetT((0+2)*2))) + nestedTab.Pos

	require.Equal(t, "nes1", byteSliceToString(nestedTab.ByteVector(nestedStrOffset)))
}

func TestGetBytes(t *testing.T) {
	s := NewScheme()
	s.AddField("name", FieldTypeString, false)
	s.AddField("price", FieldTypeFloat, false)
	s.AddArray("quantity", FieldTypeInt, false)
	b := NewBuffer(s)
	b.Set("price", float32(0.123))
	b.Set("quantity", nil)
	bytes, err := b.ToBytes()
	require.Nil(t, err, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, bytes, b.GetBytes())
}

func TestGetNames(t *testing.T) {
	s := NewScheme()
	s.AddField("name", FieldTypeString, false)
	s.AddField("price", FieldTypeFloat, false)
	s.AddField("quantity", FieldTypeInt, false)
	b := NewBuffer(s)

	require.Empty(t, b.GetNames())

	b.Set("price", 0.123)
	require.Empty(t, b.GetNames())

	b = NewBuffer(s)
	bytes, err := b.ToBytes()
	require.Nil(t, err, err)
	b = ReadBuffer(bytes, s)
	require.Empty(t, b.GetNames())

	b.Set("price", 0.123)
	bytes, err = b.ToBytes()
	require.Nil(t, err, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []string{"price"}, b.GetNames())

	b.Set("quantity", 42)
	bytes, err = b.ToBytes()
	require.Nil(t, err, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []string{"price", "quantity"}, b.GetNames())
}

func TestCorrectErrorOnReflectIsNilForArrayField(t *testing.T) {
	s := NewScheme()
	s.AddArray("arr", FieldTypeInt, true)
	b := NewBuffer(s)
	b.Set("arr", "non-array")
	bytes, err := b.ToBytes()
	// expecting panic on reflect.ValueOf("non-array").IsNil() at encodeBuffer() is avoided
	require.NotNil(t, err)
	require.Nil(t, bytes)
}

func BenchmarkSimpleDynobuffersArrayOfObjectsSet(b *testing.B) {
	s := NewScheme()
	sNested := NewScheme()
	sNested.AddField("int", FieldTypeInt, false)
	s.AddNestedArray("ints", sNested, false)

	bfNested := NewBuffer(sNested)
	bfNested.Set("int", 42)
	bf := NewBuffer(s)
	bufs := []*Buffer{bfNested}

	bf.Set("ints", bufs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bf.ToBytes()
	}
}

func BenchmarkSimpleDynobuffersArrayOfObjectsAppend(b *testing.B) {
	s := NewScheme()
	sNested := NewScheme()
	sNested.AddField("int", FieldTypeInt, false)
	s.AddNestedArray("ints", sNested, false)

	bfNested := NewBuffer(sNested)
	bfNested.Set("int", 42)
	bf := NewBuffer(s)
	bufs := []*Buffer{bfNested}

	bf.Set("ints", bufs)
	bytes, err := bf.ToBytes()
	if err != nil {
		b.Fatal(err)
	}
	bf = ReadBuffer(bytes, s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Append("ints", bufs)
		_, _ = bf.ToBytes()
	}
}

func BenchmarkSimpleFlatbuffersArrayOfObjectsAppend(b *testing.B) {
	bf := flatbuffers.NewBuilder(0)
	bf.StartObject(1)
	bf.PrependInt32(42)
	nestedOffset := bf.EndObject()
	bf.Finish(nestedOffset)

	bf.StartVector(4, 1, 4)
	bf.PrependUOffsetT(nestedOffset)
	vectorOffset := bf.EndVector(1)

	bf.StartObject(1)
	bf.PrependUOffsetTSlot(0, vectorOffset, 0)

	bf.Finish(bf.EndObject())
	bytes := bf.FinishedBytes()

	tab := &flatbuffers.Table{}
	tab.Bytes = bytes
	tab.Pos = flatbuffers.GetUOffsetT(bytes)

	for i := 0; i < b.N; i++ {
		bf := flatbuffers.NewBuilder(0)
		// read existing
		existingArrayOffset := flatbuffers.UOffsetT(tab.Offset(flatbuffers.VOffsetT((0+2)*2))) + tab.Pos
		_ = tab.Vector(existingArrayOffset - tab.Pos)
		// elemSize := 4
		existingArrayVectorOffset := tab.Vector(existingArrayOffset - tab.Pos)

		// read previous
		elemOffset := existingArrayVectorOffset + flatbuffers.UOffsetT(0)
		elem := &flatbuffers.Table{}
		elem.Bytes = bytes
		elem.Pos = tab.Indirect(elemOffset)
		elemNestedValueOffset := flatbuffers.UOffsetT(elem.Offset(flatbuffers.VOffsetT((0+2)*2))) + elem.Pos

		// encodeArray
		// write previous
		bf.StartObject(1)
		bf.PrependInt32(elem.GetInt32(elemNestedValueOffset))
		bf.Slot(0)
		arrayElemObjectPrev := bf.EndObject()
		bf.Finish(arrayElemObjectPrev)

		// write new
		bf.StartObject(1)
		bf.PrependInt32(43)
		bf.Slot(0)
		arrayElemObjectNew := bf.EndObject()
		bf.Finish(arrayElemObjectNew)

		// write array
		bf.StartVector(4, 2, 4)
		bf.PrependUOffsetT(arrayElemObjectPrev)
		bf.PrependUOffsetT(arrayElemObjectNew)
		arrayOffset := bf.EndVector(2)

		// write object with array
		bf.StartObject(1)
		bf.PrependUOffsetTSlot(0, arrayOffset, 0)
		bf.Finish(bf.EndObject())
		_ = bf.FinishedBytes()

	}
}

func BenchmarkSimpleFlatbuffersArrayOfObjectsSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bf := flatbuffers.NewBuilder(0)
		bf.StartObject(1)
		bf.PrependInt32(42)
		nestedOffset := bf.EndObject()
		bf.Finish(nestedOffset)

		bf.StartVector(4, 1, 4)
		bf.PrependUOffsetT(nestedOffset)
		vectorOffset := bf.EndVector(1)

		bf.StartObject(1)
		bf.PrependUOffsetTSlot(0, vectorOffset, 0)

		bf.Finish(bf.EndObject())
		_ = bf.FinishedBytes()
	}
}

func BenchmarkJSON(b *testing.B) {
	dest := map[string]interface{}{
		"ints": []int{42},
	}

	for i := 0; i < b.N; i++ {
		json.Marshal(dest)

	}
}

func TestObjectsCopy(t *testing.T) {
	s := NewScheme()
	sNested := NewScheme()
	sNested.AddField("int", FieldTypeInt, false)
	s.AddNestedArray("ints", sNested, false)

	b := NewBuffer(s)
	buffs := []*Buffer{}
	bNested := NewBuffer(sNested)
	bNested.Set("int", 42)
	buffs = append(buffs, bNested)
	bNested = NewBuffer(sNested)
	bNested.Set("int", 43)
	buffs = append(buffs, bNested)
	b.Set("ints", buffs)
	bytes, err := b.ToBytes()
	require.Nil(t, err)

	tab := &flatbuffers.Table{}
	tab.Bytes = bytes
	tab.Pos = flatbuffers.GetUOffsetT(bytes)
}
