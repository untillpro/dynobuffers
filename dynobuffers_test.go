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

	var s *Scheme
	var err error

	// Creating Scheme from yaml representation
	{
		s, err = YamlToScheme(schemeYaml)
		require.Nil(t, err)
	}

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

	// Check HasValue
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

	// Set string field to non-string -> error
	{
		b.Set("name", 123)
		bytes, err = b.ToBytes()
		require.NotNil(t, err)
	}

	// nil Scheme provided -> panic
	{
		require.Panics(t, func() { NewBuffer(nil) })
		require.Panics(t, func() { ReadBuffer([]byte{}, nil) })
	}
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
bytesBase64..: byte
intsObj..:
  Int: int
`

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

	require.Equal(t, "cola", b.Get("name"))
	require.Equal(t, float32(0.123), b.Get("price"))
	require.Equal(t, int32(42), b.Get("quantity"))
	require.Nil(t, b.Get("newField"))
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

	require.Equal(t, "cola", b.Get("name"))
	require.Equal(t, float32(0.123), b.Get("price"))
	require.Equal(t, int32(42), b.Get("quantity"))
	require.Nil(t, b.Get("newField"))
}

func stringArrayContains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

func testFieldValues(t *testing.T, b *Buffer, values ...interface{}) {
	names := b.GetNames()
	for i, f := range b.Scheme.Fields {
		if f.Ft != FieldTypeObject {
			require.Equal(t, values[i], b.Get(f.Name), f.Name)
		}
		if values[i] != nil {
			require.True(t, stringArrayContains(names, f.Name), f.Name)
			require.True(t, b.HasValue(f.Name), f.Name)
		} else {
			require.False(t, stringArrayContains(names, f.Name), f.Name)
			require.False(t, b.HasValue(f.Name), f.Name)
		}
		if f.IsArray {
			if f.Ft == FieldTypeObject {
				nestedArr := b.Get(f.Name).(*ObjectArray)
				valuesNesteds := values[i].([]interface{})
				require.Equal(t, nestedArr.Len, len(valuesNesteds))
				require.Equal(t, nestedArr.Buffer, nestedArr.Value()) // cover
				elementsAmount := 0
				// test using iterator
				for nestedArr.Next() {
					valuesNested := valuesNesteds[elementsAmount].([]interface{})
					testFieldValues(t, nestedArr.Buffer, valuesNested...)
					elementsAmount++
				}
				require.Equal(t, nestedArr.Len, elementsAmount)
				// test by GetByIndex
				for i := 0; i < nestedArr.Len; i++ {
					valuesNested := valuesNesteds[i].([]interface{})
					nestedBuf := b.GetByIndex(f.Name, i).(*Buffer)
					testFieldValues(t, nestedBuf, valuesNested...)
				}
				require.Nil(t, b.GetByIndex(f.Name, nestedArr.Len))
				nestedArr.Release()
			} else {
				require.Equal(t, values[i], b.Get(f.Name))
			}
		} else {
			// not array
			var okGlobal bool
			switch f.Ft {
			case FieldTypeBool:
				actual, ok := b.GetBool(f.Name)
				okGlobal = ok
				if values[i] != nil {
					require.Equal(t, values[i], actual, f.Name)
				}
			case FieldTypeByte:
				actual, ok := b.GetByte(f.Name)
				okGlobal = ok
				if values[i] != nil {
					require.Equal(t, values[i], actual, f.Name)
				}
			case FieldTypeDouble:
				actual, ok := b.GetDouble(f.Name)
				okGlobal = ok
				if values[i] != nil {
					require.Equal(t, values[i], actual, f.Name)
				}
			case FieldTypeFloat:
				actual, ok := b.GetFloat(f.Name)
				okGlobal = ok
				if values[i] != nil {
					require.Equal(t, values[i], actual, f.Name)
				}
			case FieldTypeInt:
				actual, ok := b.GetInt(f.Name)
				okGlobal = ok
				if values[i] != nil {
					require.Equal(t, values[i], actual, f.Name)
				}
			case FieldTypeLong:
				actual, ok := b.GetLong(f.Name)
				okGlobal = ok
				if values[i] != nil {
					require.Equal(t, values[i], actual, f.Name)
				}
			case FieldTypeString:
				actual, ok := b.GetString(f.Name)
				okGlobal = ok
				if values[i] != nil {
					require.Equal(t, values[i], actual, f.Name)
				}
			case FieldTypeObject:
				nested := b.Get(f.Name)
				okGlobal = values[i] != nil
				if values[i] != nil {
					testFieldValues(t, nested.(*Buffer), values[i].([]interface{})...)
				}
			}
			if values[i] == nil {
				require.False(t, okGlobal, f.Name)
			} else {
				require.True(t, okGlobal, f.Name)
			}
		}
	}
}

func TestToJSONBasic(t *testing.T) {
	scheme, err := YamlToScheme(schemeStr)
	require.Nil(t, err)

	b := NewBuffer(scheme)
	actual := map[string]interface{}{}
	jsonBytes := b.ToJSON()
	json.Unmarshal(jsonBytes, &actual)
	require.True(t, len(actual) == 0)

	// basic test
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	jsonBytes = b.ToJSON()
	json.Unmarshal(jsonBytes, &actual)
	require.True(t, len(actual) == 3)
	require.Equal(t, "cola", actual["name"])
	require.Equal(t, float64(0.123), actual["price"])
	require.Equal(t, float64(42), actual["quantity"])

	// unmodified
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)
	jsonBytes = b.ToJSON()
	json.Unmarshal(jsonBytes, &actual)
	require.True(t, len(actual) == 3)
	require.Equal(t, "cola", actual["name"])
	require.Equal(t, float64(0.123), actual["price"])
	require.Equal(t, float64(42), actual["quantity"])

	// test field initially not set
	b = NewBuffer(scheme)
	b.Set("name", "cola")
	b.Set("quantity", int32(42))
	jsonBytes = b.ToJSON()
	actual = map[string]interface{}{}
	json.Unmarshal(jsonBytes, &actual)
	require.True(t, len(actual) == 2)
	require.Equal(t, "cola", actual["name"])
	require.Equal(t, float64(42), actual["quantity"])

	// test field not set after ReadBuffer
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)
	jsonBytes = b.ToJSON()
	actual = map[string]interface{}{}
	json.Unmarshal(jsonBytes, &actual)
	require.True(t, len(actual) == 2)
	require.Equal(t, "cola", actual["name"])
	require.Equal(t, float64(42), actual["quantity"])

	// test unset field
	b.Set("quantity", nil)
	jsonBytes = b.ToJSON()
	actual = map[string]interface{}{}
	json.Unmarshal(jsonBytes, &actual)
	require.True(t, len(actual) == 1)
	require.Equal(t, "cola", actual["name"])

	// test read bytes with an unset field
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)
	jsonBytes = b.ToJSON()
	actual = map[string]interface{}{}
	json.Unmarshal(jsonBytes, &actual)
	require.True(t, len(actual) == 1)
	require.Equal(t, "cola", actual["name"])
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

	// test read bytes with an unset field
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, scheme)
	dest = b.ToJSONMap()
	require.True(t, len(dest) == 1)
	require.Equal(t, "cola", dest["name"])
}

func TestApplyJSONArrays(t *testing.T) {
	s, err := YamlToScheme(arraysAllTypesYaml)
	require.Nil(t, err)
	b := NewBuffer(s)

	// errors
	wrongs := []struct {
		json        string
		shouldBeNil bool
	}{
		// wrong types -> error (arrays expected)
		{json: `{"strings": 42}`},
		{json: `{"longs": "str"}`},
		{json: `{"ints": "str"}`},
		{json: `{"floats": "str"}`},
		{json: `{"doubles": "str"}`},
		{json: `{"bytes": "str"}`, shouldBeNil: true}, // failed to decode as base64
		{json: `{"boolTrues": "str"}`},
		{json: `{"intsObj": 42}`},
		{json: `{"unknown": 42}`, shouldBeNil: true},
		{json: `{"strings": wrong}`, shouldBeNil: true},
		{json: `{"longs": wrong}`, shouldBeNil: true},
		{json: `{"ints": wrong}`, shouldBeNil: true},
		{json: `{"floats": wrong}`, shouldBeNil: true},
		{json: `{"doubles": wrong}`, shouldBeNil: true},
		{json: `{"bytes": wrong}`, shouldBeNil: true},
		{json: `{"boolTrues": wrong}`, shouldBeNil: true},
		{json: `{"intsObj": wrong}`, shouldBeNil: true},
		{json: `{"ints": [wrong]}`, shouldBeNil: true},
		{json: `{"longs": [wrong]}`, shouldBeNil: true},
		{json: `{"floats": [wrong]}`, shouldBeNil: true},
		{json: `{"doubles": [wrong]}`, shouldBeNil: true},
		{json: `{"strings": [wrong]}`, shouldBeNil: true},
		{json: `{"boolTrues": [wrong]}`, shouldBeNil: true},
		{json: `{"intObjs": [wrong]}`, shouldBeNil: true},
		// non-base64 is provided for byte array -> error
		{json: `{"bytes": "wrong base64"}`, shouldBeNil: true},
		{json: `{"bytes": [1]}`, shouldBeNil: true},
		// failed to decode nested array -> error
		{json: `{"intObjs": [{"int":wrong}]}`, shouldBeNil: true},
		{json: `{"intObjs": [{"int":"str"}]}`, shouldBeNil: true},
		// null element is met -> error
		{json: `{"ints": [44, null]}`, shouldBeNil: true},
		{json: `{"longs": [44, null]}`, shouldBeNil: true},
		{json: `{"floats": [44, null]}`, shouldBeNil: true},
		{json: `{"doubles": [44, null]}`, shouldBeNil: true},
		{json: `{"strings": ["str", null]}`, shouldBeNil: true},
		{json: `{"boolTrues": [true, null]}`, shouldBeNil: true},
		{json: `{"intObjs": [{"int":44}, null]}`, shouldBeNil: true},
		// failed to encode nested object (mandatory field is not set) -> error
		{json: `{"intObjs": [{"int":null}]}`, shouldBeNil: true},
	}
	for _, wrong := range wrongs {
		b.Reset(nil)
		bytes, err := b.ApplyJSONAndToBytes([]byte(wrong.json))
		require.Nil(t, bytes, wrong)
		require.NotNil(t, err, wrong)
		if wrong.shouldBeNil {
			require.True(t, b.IsNil(), wrong.json)
		} else {
			require.False(t, b.IsNil(), wrong.json)
		}
	}

	// apply all values
	bytes, err := b.ApplyJSONAndToBytes([]byte(`{"ints": [44, 45], "longs": [42, 43], "floats": [0.124, 0.125],
		"doubles": [0.126, 0.127], "strings": ["str1", "str2"], "boolTrues": [true, true], "boolFalses": [false,false],
		"bytes": "BQY=", "bytesBase64":"BQY=", "intsObj":[{"int":42},{"int":43}]}`))
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{44, 45}, []int64{42, 43}, []float32{0.124, 0.125}, []float64{0.126, 0.127},
		[]string{"str1", "str2"}, []bool{true, true}, []bool{false, false}, []byte{5, 6}, []byte{5, 6},
		[]interface{}{[]interface{}{int32(42)}, []interface{}{int32(43)}})

	// append arrays
	bytes, err = b.ApplyJSONAndToBytes([]byte(`{"ints": [46, 47], "longs": [48, 49], "floats": [0.128, 0.129],
		"doubles": [0.130, 0.131], "strings": ["str3", "str4"], "boolTrues": [false, false], "boolFalses": [true, true],
		"bytes": "BQY=", "bytesBase64":"BQY=", "intsObj":[{"int":50},{"int":51}]}`))
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{44, 45, 46, 47}, []int64{42, 43, 48, 49}, []float32{0.124, 0.125, 0.128, 0.129}, []float64{0.126, 0.127, 0.130, 0.131},
		[]string{"str1", "str2", "str3", "str4"}, []bool{true, true, false, false}, []bool{false, false, true, true}, []byte{5, 6, 5, 6}, []byte{5, 6, 5, 6},
		[]interface{}{[]interface{}{int32(42)}, []interface{}{int32(43)}, []interface{}{int32(50)}, []interface{}{int32(51)}})

	// unset all using nulls and empty arrays
	jsons := [][]byte{
		[]byte(`{"ints": null, "longs": null, "floats": null, "doubles": null, "strings": null, "boolTrues": null, "boolFalses": null,"bytes": null, "bytesBase64":null, "intsObj":null}`),
		[]byte(`{"ints": [], "longs": [], "floats": [],"doubles": [], "strings": [], "boolTrues": [], "boolFalses": [],"bytes": "", "bytesBase64":"", "intsObj":[]}`),
	}
	for _, json := range jsons {
		// unset all on existing -> nothing to store
		bytes, err = b.ApplyJSONAndToBytes(json)
		require.Nil(t, err)
		require.Nil(t, bytes)

		// initially not set -> nothing to store
		b1 := NewBuffer(s)
		bytes, err := b1.ApplyJSONAndToBytes(json)
		require.Nil(t, err)
		require.Nil(t, bytes)
	}
}

func TestApplyJSON(t *testing.T) {
	schemeRoot, err := YamlToScheme(allTypesYaml)
	require.Nil(t, err)
	schemeNested := NewScheme()
	schemeNested.AddField("price", FieldTypeFloat, false)
	schemeNested.AddField("quantity", FieldTypeInt, true)
	schemeRoot.AddNested("nested1", schemeNested, false)
	schemeRoot.AddNested("nested2", schemeNested, false)

	// apply empty
	b := NewBuffer(schemeRoot)
	jsons := map[string][]byte{
		"nil":       nil,
		"empty str": []byte(""),
		"empty obj": []byte("{}"),
		"null json": []byte("null"),
		"all nulls": []byte(`{"string": null, "long": null, "int": null, "float": null,
			"double": null, "byte": null, "boolTrue": null, "boolFalse": null, "nested1": null, "nested2":{}}`),
	}
	for desc, js := range jsons {
		bytes, err := b.ApplyJSONAndToBytes(js)
		require.Nil(t, err, desc)
		require.Nil(t, bytes, desc)
	}

	// errors on wrong type or JSON
	wrongs := []struct {
		json        string
		shouldBeNil bool
	}{
		{json: `{"string": 42}`, shouldBeNil: true},
		{json: `{"long": "str"}`},
		{json: `{"int": "str"}`},
		{json: `{"float": "str"}`},
		{json: `{"double": "str"}`},
		{json: `{"byte": "str"}`},
		{json: `{"boolTrue": "str"}`},
		{json: `{"nested1": 42}`},
		{json: `{"nested1": []}`},
		{json: `{"string": wrong}`, shouldBeNil: true},
		{json: `{"long": wrong}`, shouldBeNil: true},
		{json: `{"int": wrong}`, shouldBeNil: true},
		{json: `{"float": wrong}`, shouldBeNil: true},
		{json: `{"double": wrong}`, shouldBeNil: true},
		{json: `{"byte": wrong}`, shouldBeNil: true},
		{json: `{"boolTrue": wrong}`, shouldBeNil: true},
		{json: `{"nested1": wrong}`, shouldBeNil: true},
		{json: `{"unknown": 42}`, shouldBeNil: true},
		{json: `{wrong}`, shouldBeNil: true},
		{json: `wrong`, shouldBeNil: true},
		// mandatory field in nested object is null -> error
		{json: `{"nested1": {"price": 42}}`},
		{json: `{"nested1": {"quantity": null}}`},
	}
	for _, wrong := range wrongs {
		b.Reset(nil)
		bytes, err := b.ApplyJSONAndToBytes([]byte(wrong.json))
		require.Nil(t, bytes, wrong)
		require.NotNil(t, err, wrong)
		if wrong.shouldBeNil {
			require.True(t, b.IsNil(), wrong.json)
		} else {
			require.False(t, b.IsNil(), wrong.json)
		}
	}

	// apply all values
	bytes, err := b.ApplyJSONAndToBytes([]byte(`{"string": "str", "long": 42, "int": 43, "float": 0.124,
		"double": 0.125, "byte": 6, "boolTrue": true, "boolFalse": false,
		"nested1": {"price": 0.126,"quantity":44}, "nested2": {"price": 0.127,"quantity":45}}`))
	require.Nil(t, err)
	b = ReadBuffer(bytes, schemeRoot)
	testFieldValues(t, b, int32(43), int64(42), float32(0.124), float64(0.125), "str", true, false, byte(6),
		[]interface{}{float32(0.126), int32(44)}, []interface{}{float32(0.127), int32(45)})

	// unset all
	// note: nested2:{} - mandatory field is not set but ok because empty object means no object
	b = ReadBuffer(bytes, schemeRoot)
	bytes, err = b.ApplyJSONAndToBytes([]byte(`{"string": null, "long": null, "int": null, "float": null,
		"double": null, "byte": null, "boolTrue": null, "boolFalse": null, "nested1": null, "nested2":{}}`))
	require.Nil(t, err)
	require.Nil(t, bytes)

}

func TestAllValues(t *testing.T) {
	s, err := YamlToScheme(allTypesYaml)
	require.Nil(t, err)
	sNested := NewScheme()
	sNested.AddField("int", FieldTypeInt, false)
	s.AddNested("nes", sNested, false)

	b := NewBuffer(s)
	testEmpty(t, b)

	// no data -> nothing
	bytes, err := b.ToBytes()
	require.Nil(t, bytes)
	require.Nil(t, err)

	// all is nil -> nothing
	for _, f := range s.Fields {
		b.Set(f.Name, nil)
	}
	bytes, err = b.ToBytes()
	require.Nil(t, bytes)
	require.Nil(t, err)

	// wrong types -> error
	wrongs := map[string]interface{}{
		"int":       "str",
		"long":      "str",
		"float":     "str",
		"double":    "str",
		"string":    1,
		"boolTrue":  "str",
		"boolFalse": "str",
		"byte":      "str",
		"nes":       "str",
	}
	for fn, wrong := range wrongs {
		b = NewBuffer(s)
		b.Set(fn, wrong)
		bytes, err = b.ToBytes()
		require.Nil(t, bytes)
		require.NotNil(t, err)
	}

	// fill values
	b.Set("int", int32(1))
	b.Set("long", int64(2))
	b.Set("float", float32(0.1))
	b.Set("double", float64(0.2))
	b.Set("string", "str")
	b.Set("boolTrue", true)
	b.Set("boolFalse", false)
	b.Set("byte", byte(3))
	bNested := NewBuffer(sNested)
	bNested.Set("int", 4)
	b.Set("nes", bNested)
	bytesFilled, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytesFilled, s)
	expectedValues := []interface{}{int32(1), int64(2), float32(0.1), float64(0.2), "str", true, false, byte(3), []interface{}{int32(4)}}
	testFieldValues(t, b, expectedValues...)

	// ToBytesWithBuilder on unmodified Buffer -> return underlying byte array
	b.Reset(bytesFilled)
	bl := flatbuffers.NewBuilder(0)
	require.Nil(t, b.ToBytesWithBuilder(bl))
	b = ReadBuffer(bl.Bytes, s)
	testFieldValues(t, b, expectedValues...)

	// GetBytes will return ToBytes() which will return the underlying byte array here
	b.Reset(b.GetBytes())
	testFieldValues(t, b, expectedValues...)

	// ToBytes() on modified Buffer -> re-encode existing + modifications
	// modify the buffer to force re-encode, otherwise underlying byte array will be returned
	b.Set("int", b.Get("int"))
	bytesFilled, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytesFilled, s)
	testFieldValues(t, b, expectedValues...)

	// set nested object to an empty object -> no nested object
	bNested = NewBuffer(sNested)
	b.Set("nes", bNested)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Nil(t, b.Get("nes"))

	// unset values by one (check false is returned by typed Get*())
	expectedValuesCopy := make([]interface{}, len(expectedValues))
	copy(expectedValuesCopy, expectedValues)
	for i, f := range s.Fields {
		b = ReadBuffer(bytesFilled, s)
		expectedValuesCopy[i] = nil
		b.Set(f.Name, nil)
		bytes, err := b.ToBytes()
		require.Nil(t, err)
		b = ReadBuffer(bytes, s)
		testFieldValues(t, b, expectedValuesCopy...)
		expectedValuesCopy[i] = expectedValues[i]
	}

	// unset values
	for _, f := range s.Fields {
		b.Set(f.Name, nil)
	}
	bytes, err = b.ToBytes()
	require.Nil(t, bytes)
	require.Nil(t, err)
}

func testEmpty(t *testing.T, b *Buffer) {
	for _, f := range b.Scheme.Fields {
		require.Nil(t, b.Get(f.Name), f.Name)
		_, ok := b.GetString(f.Name)
		require.False(t, ok, f.Name)
		_, ok = b.GetBool(f.Name)
		require.False(t, ok, f.Name)
		_, ok = b.GetByte(f.Name)
		require.False(t, ok, f.Name)
		_, ok = b.GetInt(f.Name)
		require.False(t, ok, f.Name)
		_, ok = b.GetLong(f.Name)
		require.False(t, ok, f.Name)
		_, ok = b.GetFloat(f.Name)
		require.False(t, ok, f.Name)
		_, ok = b.GetDouble(f.Name)
		require.False(t, ok, f.Name)
	}
}

func TestApplyMap(t *testing.T) {
	s, err := YamlToScheme(allTypesYaml)
	require.Nil(t, err)
	sNested := NewScheme()
	sNested.AddField("int", FieldTypeInt, false)
	s.AddNested("nes", sNested, false)

	// applied nil -> nothing to store
	b := NewBuffer(s)
	require.Nil(t, b.ApplyMap(nil))
	bytes, err := b.ToBytes()
	require.Nil(t, bytes)
	require.Nil(t, err)

	// applied empty -> nothing to store
	require.Nil(t, b.ApplyMap(map[string]interface{}{}))
	bytes, err = b.ToBytes()
	require.Nil(t, bytes)
	require.Nil(t, err)

	// applied nil fields -> nothing to store
	require.Nil(t, b.ApplyMap(map[string]interface{}{
		"int":       nil,
		"long":      nil,
		"float":     nil,
		"double":    nil,
		"string":    nil,
		"boolTrue":  nil,
		"boolFalse": nil,
		"byte":      nil,
		"nes":       nil,
	}))
	bytes, err = b.ToBytes()
	require.Nil(t, bytes)
	require.Nil(t, err)

	// wrong types -> error
	wrongs := []struct {
		m            map[string]interface{}
		errorOnApply bool
	}{
		{m: map[string]interface{}{"int": "str"}},
		{m: map[string]interface{}{"long": "str"}},
		{m: map[string]interface{}{"float": "str"}},
		{m: map[string]interface{}{"double": "str"}},
		{m: map[string]interface{}{"string": 1}},
		{m: map[string]interface{}{"boolTrue": "str"}},
		{m: map[string]interface{}{"byte": "str"}},
		{m: map[string]interface{}{"nes": "str"}, errorOnApply: true},
		{m: map[string]interface{}{"unknown": 42}, errorOnApply: true},
	}
	for _, wrong := range wrongs {
		b = NewBuffer(s)
		if wrong.errorOnApply {
			require.NotNil(t, b.ApplyMap(wrong.m))
		} else {
			require.Nil(t, b.ApplyMap(wrong.m))
		}
		bytes, err = b.ToBytes()
		require.Nil(t, bytes)
		if wrong.errorOnApply {
			require.Nil(t, err)
		} else {
			require.NotNil(t, err)
		}
	}

	// apply values
	require.Nil(t, b.ApplyMap(map[string]interface{}{
		"int":       int32(1),
		"long":      int64(2),
		"float":     float32(0.1),
		"double":    float64(0.2),
		"string":    "str",
		"boolTrue":  true,
		"boolFalse": false,
		"byte":      byte(3),
		"nes": map[string]interface{}{
			"int": int32(4),
		},
	}))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, int32(1), int64(2), float32(0.1), float64(0.2), "str", true, false, byte(3), []interface{}{int32(4)})

	// unset values
	require.Nil(t, b.ApplyMap(map[string]interface{}{
		"int":       nil,
		"long":      nil,
		"float":     nil,
		"double":    nil,
		"string":    nil,
		"boolTrue":  nil,
		"boolFalse": nil,
		"byte":      nil,
		"nes":       nil,
	}))
	bytes, err = b.ToBytes()
	require.Nil(t, bytes)
	require.Nil(t, err)

	// apply json map
	b = NewBuffer(s)
	jsonStr := []byte(`{"int": 1, "long": 2, "float": 0.1, "double": 0.2, "string": "str", "boolTrue": true, "boolFalse": false, "byte": 3, "nes": {"int": 4}}`)
	m := map[string]interface{}{}
	require.Nil(t, json.Unmarshal(jsonStr, &m))
	require.Nil(t, b.ApplyMap(m))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, int32(1), int64(2), float32(0.1), float64(0.2), "str", true, false, byte(3), []interface{}{int32(4)})

	// unset from json
	jsonStr = []byte(`{"string": null, "long": null, "int": null, "float": null, "double": null, "byte": null, "boolTrue": null, "boolFalse": null, "nes": null}`)
	m = map[string]interface{}{}
	require.Nil(t, json.Unmarshal(jsonStr, &m))
	require.Nil(t, b.ApplyMap(m))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	require.Nil(t, bytes)
}

func TestApplyMapArrays(t *testing.T) {
	s, err := YamlToScheme(arraysAllTypesYaml)
	require.Nil(t, err)
	b := NewBuffer(s)

	// empty
	ms := []map[string]interface{}{
		nil,
		{},
		{
			"ints":        nil,
			"longs":       nil,
			"floats":      nil,
			"doubles":     nil,
			"strings":     nil,
			"boolTrues":   nil,
			"boolFalses":  nil,
			"bytes":       nil,
			"bytesBase64": nil,
			"intsObj":     nil,
		},
		{
			"ints":      []int32{},
			"longs":     []int64{},
			"floats":    []float32{},
			"doubles":   []float64{},
			"strings":   []string{},
			"boolTrues": []bool{},
			"bytes":     []byte{},
			"intsObj":   []interface{}{},
		},
	}

	for _, m := range ms {
		require.Nil(t, b.ApplyMap(m))
		bytes, err := b.ToBytes()
		require.Nil(t, bytes)
		require.Nil(t, err)
	}

	// errors
	wrongs := []struct {
		m            map[string]interface{}
		errorOnApply bool
	}{
		// wrong types -> error: non-array provided
		{m: map[string]interface{}{"longs": "str"}},
		{m: map[string]interface{}{"ints": "str"}},
		{m: map[string]interface{}{"floats": "str"}},
		{m: map[string]interface{}{"doubles": "str"}},
		{m: map[string]interface{}{"bytes": "str"}}, // failed to decode as base64
		{m: map[string]interface{}{"boolTrues": "str"}},
		{m: map[string]interface{}{"intsObj": 42}, errorOnApply: true},
		{m: map[string]interface{}{"unknown": 42}, errorOnApply: true},
		// wrong types -> error: array of wrong type provided
		{m: map[string]interface{}{"strings": []int16{42}}},
		{m: map[string]interface{}{"longs": []int16{42}}},
		{m: map[string]interface{}{"ints": []int16{42}}},
		{m: map[string]interface{}{"floats": []int16{42}}},
		{m: map[string]interface{}{"doubles": []int16{42}}},
		{m: map[string]interface{}{"bytes": []int16{42}}},
		{m: map[string]interface{}{"boolTrues": []int16{42}}},
		{m: map[string]interface{}{"intsObj": []int16{42}}, errorOnApply: true},
		// non-base64 is provided for byte array -> error
		{m: map[string]interface{}{"bytes": "wrong base64"}},
		{m: map[string]interface{}{"bytes": []int32{1}}},
		// failed to decode nested array -> error
		{m: map[string]interface{}{"intsObj": []interface{}{
			map[string]interface{}{"int": "str"},
		}}},
		// nil element is met -> error
		{m: map[string]interface{}{"ints": []interface{}{44, nil}}},
		{m: map[string]interface{}{"longs": []interface{}{44, nil}}},
		{m: map[string]interface{}{"floats": []interface{}{44, nil}}},
		{m: map[string]interface{}{"doubles": []interface{}{44, nil}}},
		{m: map[string]interface{}{"strings": []interface{}{44, nil}}},
		{m: map[string]interface{}{"boolTrues": []interface{}{44, nil}}},
		{m: map[string]interface{}{"intsObj": []interface{}{map[string]interface{}{"int": 44}, nil}}, errorOnApply: true},
		// failed to encode an array element (required field notset) -> error
		{m: map[string]interface{}{"intsObj": []interface{}{map[string]interface{}{"int": nil}}}},
	}
	for _, wrong := range wrongs {
		b.Reset(nil)
		if wrong.errorOnApply {
			require.NotNil(t, b.ApplyMap(wrong.m), wrong)
		} else {
			require.Nil(t, b.ApplyMap(wrong.m), wrong)
		}
		bytes, err := b.ToBytes()
		require.Nil(t, bytes, wrong)
		if wrong.errorOnApply {
			require.Nil(t, err, wrong)
		} else {
			require.NotNil(t, err, wrong)
		}
	}

	// set values
	m := map[string]interface{}{
		"ints":        []int32{1, 2},
		"longs":       []int64{3, 4},
		"floats":      []float32{0.1, 0.2},
		"doubles":     []float64{0.3, 0.4},
		"strings":     []string{"str1", "str2"},
		"boolTrues":   []bool{true, true},
		"boolFalses":  []bool{false, false},
		"bytes":       []byte{7, 8},
		"bytesBase64": "BQY=",
		"intsObj": []interface{}{
			map[string]interface{}{
				"int": int32(7),
			},
		},
	}
	b = NewBuffer(s)
	require.Nil(t, b.ApplyMap(m))
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true},
		[]bool{false, false}, []byte{7, 8}, []byte{5, 6}, []interface{}{[]interface{}{int32(7)}})

	// append values. Types of all numerics are matched to the scheme
	m = map[string]interface{}{
		"ints":        []int32{9, 10},
		"longs":       []int64{11, 12},
		"floats":      []float32{0.5, 0.6},
		"doubles":     []float64{0.7, 0.8},
		"strings":     []string{"str3", "str4"},
		"boolTrues":   []bool{false, false},
		"boolFalses":  []bool{true, true},
		"bytes":       []byte{13, 14},
		"bytesBase64": "BQY=",
		"intsObj": []interface{}{
			map[string]interface{}{
				"int": int32(8),
			},
		},
	}
	require.Nil(t, b.ApplyMap(m))
	bytesFilled, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytesFilled, s)
	testFieldValues(t, b, []int32{1, 2, 9, 10}, []int64{3, 4, 11, 12}, []float32{0.1, 0.2, 0.5, 0.6}, []float64{0.3, 0.4, 0.7, 0.8}, []string{"str1", "str2", "str3", "str4"}, []bool{true, true, false, false},
		[]bool{false, false, true, true}, []byte{7, 8, 13, 14}, []byte{5, 6, 5, 6}, []interface{}{[]interface{}{int32(7)}, []interface{}{int32(8)}})

	// unset all by nils
	m = map[string]interface{}{
		"ints":        nil,
		"longs":       nil,
		"floats":      nil,
		"doubles":     nil,
		"strings":     nil,
		"boolTrues":   nil,
		"boolFalses":  nil,
		"bytes":       nil,
		"bytesBase64": nil,
		"intsObj":     nil,
	}
	require.Nil(t, b.ApplyMap(m))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	require.Nil(t, bytes)

	// unset all by empty arrays
	b = ReadBuffer(bytesFilled, s)
	m = map[string]interface{}{
		"ints":        []int32{},
		"longs":       []int64{},
		"floats":      []float32{},
		"doubles":     []float64{},
		"strings":     []string{},
		"boolTrues":   []bool{},
		"boolFalses":  []bool{},
		"bytes":       []byte{},
		"bytesBase64": []byte{},
		"intsObj":     []interface{}{},
	}
	require.Nil(t, b.ApplyMap(m))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	require.Nil(t, bytes)

	// unset all by empty arrays from json (check []float64 for numerics)
	// note: `bytes` will be unmarshaled to []interface{}{}. Should be []byte or base64 string
	jsonStr := []byte(`{"ints":[],"longs":[],"floats":[],"doubles":[],"strings":[],"boolTrues":[],"boolFalses":[],"bytes":null,"bytesBase64": "", "intsObj":[]}`)
	m = map[string]interface{}{}
	require.Nil(t, json.Unmarshal(jsonStr, &m))
	b = ReadBuffer(bytesFilled, s)
	require.Nil(t, b.ApplyMap(m))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	require.Nil(t, bytes)

	// load from json. All numerics are float64. No errors expected despite type are not matched to the scheme
	jsonStr = []byte(`{"ints":[1, 2],"longs":[3, 4],"floats":[0.1, 0.2],"doubles":[0.3, 0.4],"strings":["str1", "str2"],"boolTrues":[true, true],"boolFalses":[false, false],
		"bytes":"BQY=","bytesBase64": "BQY=", "intsObj":[{"int": 5}]}`)
	m = map[string]interface{}{}
	require.Nil(t, json.Unmarshal(jsonStr, &m))
	b = NewBuffer(s)
	require.Nil(t, b.ApplyMap(m))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true},
		[]bool{false, false}, []byte{5, 6}, []byte{5, 6}, []interface{}{[]interface{}{int32(5)}})

	// unset all by nulls from json
	// note: `bytes` will be unmarshaled to []interface{}{}. Should be []byte or base64 string
	jsonStr = []byte(`{"ints":null,"longs":null,"floats":null,"doubles":null,"strings":null,"boolTrues":null,"boolFalses":null,"bytes":null,"bytesBase64": null, "intsObj":null}`)
	m = map[string]interface{}{}
	require.Nil(t, json.Unmarshal(jsonStr, &m))
	b = ReadBuffer(bytesFilled, s)
	require.Nil(t, b.ApplyMap(m))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	require.Nil(t, bytes)
}

func TestToJSONAndToJSONMap(t *testing.T) {
	s, err := YamlToScheme(arraysAllTypesYaml)
	s.AddNested("intObj", s.GetNestedScheme("intsObj"), false)
	require.Nil(t, err)
	b := NewBuffer(s)

	// empty -> empty json and map
	empties := []map[string]interface{}{
		nil,
		{
			"ints":        []int32{},
			"longs":       []int64{},
			"floats":      []float32{},
			"doubles":     []float64{},
			"strings":     []string{},
			"boolTrues":   []bool{},
			"boolFalses":  []bool{},
			"bytes":       []byte{},
			"bytesBase64": "",
			"intsObj":     []interface{}{},
			"intObj":      map[string]interface{}{},
		},
		{},
		{
			"ints":        nil,
			"longs":       nil,
			"floats":      nil,
			"doubles":     nil,
			"strings":     nil,
			"boolTrues":   nil,
			"boolFalses":  nil,
			"bytes":       nil,
			"bytesBase64": nil,
			"intsObj":     nil,
			"intObj":      nil,
		},
	}

	for _, empty := range empties {
		b = NewBuffer(s)
		b.ApplyMap(empty)
		require.Equal(t, []byte("{}"), b.ToJSON())
		require.Empty(t, b.ToJSONMap())
	}

	// set values
	m := map[string]interface{}{
		"ints":        []int32{1, 2},
		"longs":       []int64{3, 4},
		"floats":      []float32{0.1, 0.2},
		"doubles":     []float64{0.3, 0.4},
		"strings":     []string{"str1", "str2"},
		"boolTrues":   []bool{true, true},
		"boolFalses":  []bool{false, false},
		"bytes":       []byte{7, 8},
		"bytesBase64": []byte{5, 6},
		"intsObj": []interface{}{
			map[string]interface{}{
				"int": int32(7),
			},
		},
		"intObj": map[string]interface{}{
			"int": int32(-7),
		},
	}
	b = NewBuffer(s)
	require.Nil(t, b.ApplyMap(m))
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)

	// test ToJSON
	jsonBytes := b.ToJSON()
	b = NewBuffer(s)
	bytes, err = b.ApplyJSONAndToBytes(jsonBytes)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true},
		[]bool{false, false}, []byte{7, 8}, []byte{5, 6}, []interface{}{[]interface{}{int32(7)}}, []interface{}{int32(-7)})

	// case when intsObj is []*Buffer: Set(name, []*Buffer) is called
	intsObjs := []*Buffer{}
	intsObj := NewBuffer(s.GetNestedScheme("intsObj"))
	intsObj.Set("int", 8)
	intsObjs = append(intsObjs, intsObj)
	b.Set("intsObj", intsObjs)
	jsonBytes = b.ToJSON()
	b = NewBuffer(s)
	bytes, err = b.ApplyJSONAndToBytes(jsonBytes)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true},
		[]bool{false, false}, []byte{7, 8}, []byte{5, 6}, []interface{}{[]interface{}{int32(8)}}, []interface{}{int32(-7)})

	// case when intsObj is *buffersSlice: ApplyMap() called
	b = NewBuffer(s)
	require.Nil(t, b.ApplyMap(m))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	m = b.ToJSONMap()
	b = NewBuffer(s)
	b.ApplyMap(m) // here array of objects is *bufferSlice
	b.ToJSON()
	b = NewBuffer(s)
	bytes, err = b.ApplyJSONAndToBytes(jsonBytes)
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true},
		[]bool{false, false}, []byte{7, 8}, []byte{5, 6}, []interface{}{[]interface{}{int32(8)}}, []interface{}{int32(-7)})

	// test ToJSONMap
	b = NewBuffer(s)
	require.Nil(t, b.ApplyMap(m))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	m = b.ToJSONMap()
	b = NewBuffer(s)
	b.ApplyMap(m)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true},
		[]bool{false, false}, []byte{7, 8}, []byte{5, 6}, []interface{}{[]interface{}{int32(7)}}, []interface{}{int32(-7)})

	// case when intsObj is []*Buffer: Set(name, []*Buffer) called
	intsObjs = []*Buffer{}
	intsObj = NewBuffer(s.GetNestedScheme("intsObj"))
	intsObj.Set("int", 9)
	intsObjs = append(intsObjs, intsObj)
	b.Set("intsObj", intsObjs)
	m = b.ToJSONMap()
	b = NewBuffer(s)
	b.ApplyMap(m)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true},
		[]bool{false, false}, []byte{7, 8}, []byte{5, 6}, []interface{}{[]interface{}{int32(9)}}, []interface{}{int32(-7)})

	// case when intsObj is *buffersSlice: ApplyMap() called
	b = NewBuffer(s)
	require.Nil(t, b.ApplyMap(m))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	m = b.ToJSONMap()
	b = NewBuffer(s)
	b.ApplyMap(m)     // *bufferSlice is used
	m = b.ToJSONMap() // *bufferSlice is used on map filling
	b = NewBuffer(s)
	b.ApplyMap(m)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true},
		[]bool{false, false}, []byte{7, 8}, []byte{5, 6}, []interface{}{[]interface{}{int32(9)}}, []interface{}{int32(-7)})
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

	require.Equal(t, schemeRoot.Fields, schemeNew.Fields)
	require.Equal(t, schemeRoot.FieldsMap, schemeNew.FieldsMap)

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

	// set to empty -> nil
	longs = []int64{}
	b.Set("longs", longs)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	require.Nil(t, b.Get("longs"))
}

func TestArrays(t *testing.T) {
	s, err := YamlToScheme(arraysAllTypesYaml)
	require.Nil(t, err)
	b := NewBuffer(s)

	// empty and nil arrays -> nothing
	tests := map[string][]interface{}{
		"ints":      {nil, []int32{}},
		"longs":     {nil, []int64{}},
		"floats":    {nil, []float32{}},
		"doubles":   {nil, []float64{}},
		"boolTrues": {nil, []bool{}},
		"bytes":     {nil, "", []byte{}},
		"strings":   {nil, []string{}, [][]byte{}},
		"intsObj":   {nil, []*Buffer{}, &buffersSlice{}},
	}
	for fn, values := range tests {
		for _, value := range values {
			// set
			b = NewBuffer(s)
			b.Set(fn, value)
			bytes, err := b.ToBytes()
			require.Nil(t, err, fn, value)
			require.Nil(t, bytes, fn, value)
			// append
			b = NewBuffer(s)
			b.Append(fn, value)
			bytes, err = b.ToBytes()
			require.Nil(t, err, fn, value)
			require.Nil(t, bytes, fn, value)
		}
	}

	testsErrors := map[string][]func(b *Buffer){
		"ints": {
			func(b *Buffer) { b.Set("ints", 42) },
			func(b *Buffer) { b.Set("ints", []int16{}) },
		},
		"longs": {
			func(b *Buffer) { b.Set("longs", 42) },
			func(b *Buffer) { b.Set("longs", []int16{}) },
		},
		"floats": {
			func(b *Buffer) { b.Set("floats", 42) },
			func(b *Buffer) { b.Set("floats", []int16{}) },
		},
		"doubles": {
			func(b *Buffer) { b.Set("doubles", 42) },
			func(b *Buffer) { b.Set("doubles", []int16{}) },
		},
		"boolTrues": {
			func(b *Buffer) { b.Set("boolTrues", 42) },
			func(b *Buffer) { b.Set("boolTrues", []int16{}) },
		},
		"bytes": {
			func(b *Buffer) { b.Set("bytes", 42) },
			func(b *Buffer) { b.Set("bytes", []int16{}) },
			func(b *Buffer) { b.Set("bytes", "wrong base64") },
		},
		"strings": {
			func(b *Buffer) { b.Set("strings", 42) },
			func(b *Buffer) { b.Set("strings", []int16{}) },
		},
		"intsObj": {
			func(b *Buffer) { b.Set("intsObj", 42) },
			func(b *Buffer) { b.Set("intsObj", []int16{}) },
			func(b *Buffer) {
				// nil element is met in []*Buffer -> error
				nested := NewBuffer(s.GetNestedScheme("intsObj"))
				nested.Set("int", 42)
				b.Set("intsObj", []*Buffer{nested, nil})
			},

			func(b *Buffer) {
				// nil element is met in *bufferSlice -> error. Impossible to test this on ToBytes() after ApplyMap()
				// because ApplyMap() itselffails if nil is met
				nested := NewBuffer(s.GetNestedScheme("intsObj"))
				nested.Set("int", 42)
				bs := getBufferSlice(0)
				bs.Scheme = nested.Scheme
				bs.Owner = b
				bs.Slice = append(bs.Slice, nested)
				bs.Slice = append(bs.Slice, nil)
				b.Set("intsObj", bs)
			},
			func(b *Buffer) {
				// failed to encode an element (required field is not set) -> error
				nested := NewBuffer(s.GetNestedScheme("intsObj"))
				b.Set("intsObj", []*Buffer{nested, nil})
			},
		},
	}

	for fn, tests := range testsErrors {
		for _, test := range tests {
			b = NewBuffer(s)
			test(b)
			bytes, err := b.ToBytes()
			require.NotNil(t, err, fn)
			require.Nil(t, bytes, fn)
		}
	}

	// append to nothing is equal to Set()
	b = NewBuffer(s)
	b.Append("unknown", 42) // nothing happens, no error
	b.Append("ints", []int32{1, 2})
	b.Append("longs", []int64{3, 4})
	b.Append("floats", []float32{0.1, 0.2})
	b.Append("doubles", []float64{0.3, 0.4})
	b.Append("strings", []string{"str1", "str2"})
	b.Append("boolTrues", []bool{true, true})
	b.Append("boolFalses", []bool{false, false})
	b.Append("bytes", []byte{1, 2})
	b.Append("bytesBase64", "BQY=")
	bNestedArr := []*Buffer{}
	bNested := NewBuffer(s.GetNestedScheme("intsObj"))
	bNested.Set("int", 5)
	bNestedArr = append(bNestedArr, bNested)
	bNested = NewBuffer(s.GetNestedScheme("intsObj"))
	bNested.Set("int", 6)
	bNestedArr = append(bNestedArr, bNested)
	b.Append("intsObj", bNestedArr)
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true}, []bool{false, false},
		[]byte{1, 2}, []byte{5, 6}, []interface{}{[]interface{}{int32(5)}, []interface{}{int32(6)}})

	// set values
	b = NewBuffer(s)
	b.Set("ints", []int32{1, 2})
	b.Set("longs", []int64{3, 4})
	b.Set("floats", []float32{0.1, 0.2})
	b.Set("doubles", []float64{0.3, 0.4})
	b.Set("strings", []string{"str1", "str2"})
	b.Set("boolTrues", []bool{true, true})
	b.Set("boolFalses", []bool{false, false})
	b.Set("bytes", []byte{1, 2})
	b.Set("bytesBase64", "BQY=")
	bNestedArr = []*Buffer{}
	bNested = NewBuffer(s.GetNestedScheme("intsObj"))
	bNested.Set("int", 5)
	bNestedArr = append(bNestedArr, bNested)
	bNested = NewBuffer(s.GetNestedScheme("intsObj"))
	bNested.Set("int", 6)
	bNestedArr = append(bNestedArr, bNested)
	b.Set("intsObj", bNestedArr)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true}, []bool{false, false},
		[]byte{1, 2}, []byte{5, 6}, []interface{}{[]interface{}{int32(5)}, []interface{}{int32(6)}})

	// non-modified arrays should be copied
	// set buffer modified to force re-encode, otherwise underlying byte array wil be returned
	b.Set("ints", b.Get("ints"))
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2}, []int64{3, 4}, []float32{0.1, 0.2}, []float64{0.3, 0.4}, []string{"str1", "str2"}, []bool{true, true}, []bool{false, false},
		[]byte{1, 2}, []byte{5, 6}, []interface{}{[]interface{}{int32(5)}, []interface{}{int32(6)}})

	// append existing
	b.Append("ints", []int32{7, 8})
	b.Append("longs", []int64{9, 10})
	b.Append("floats", []float32{0.5, 0.6})
	b.Append("doubles", []float64{0.7, 0.8})
	b.Append("strings", []string{"str3", "str4"})
	b.Append("boolTrues", []bool{false, false})
	b.Append("boolFalses", []bool{true, true})
	b.Append("bytes", []byte{11, 12})
	b.Append("bytesBase64", "BQY=")
	bNestedArr = []*Buffer{}
	bNested = NewBuffer(s.GetNestedScheme("intsObj"))
	bNested.Set("int", 11)
	bNestedArr = append(bNestedArr, bNested)
	bNested = NewBuffer(s.GetNestedScheme("intsObj"))
	bNested.Set("int", 12)
	bNestedArr = append(bNestedArr, bNested)
	b.Append("intsObj", bNestedArr)
	bytes, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes, s)
	testFieldValues(t, b, []int32{1, 2, 7, 8}, []int64{3, 4, 9, 10}, []float32{0.1, 0.2, 0.5, 0.6}, []float64{0.3, 0.4, .7, 0.8}, []string{"str1", "str2", "str3", "str4"},
		[]bool{true, true, false, false}, []bool{false, false, true, true}, []byte{1, 2, 11, 12}, []byte{5, 6, 5, 6},
		[]interface{}{[]interface{}{int32(5)}, []interface{}{int32(6)}, []interface{}{int32(11)}, []interface{}{int32(12)}})

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

	// test b.names reuse
	b.Release()
	b = NewBuffer(s)
	b.Set("price", 0.123)
	b.Set("quantity", 42)
	bytes, err = b.ToBytes()
	require.Nil(t, err, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []string{"price", "quantity"}, b.GetNames())

	// Reset
	b.Reset(nil)
	b.Set("price", 0.123)
	b.Set("quantity", 42)
	bytes, err = b.ToBytes()
	require.Nil(t, err, err)
	b = ReadBuffer(bytes, s)
	require.Equal(t, []string{"price", "quantity"}, b.GetNames())
}

func TestGetNestedScheme(t *testing.T) {
	bNested := NewScheme()
	bRoot := NewScheme()
	bRoot.AddNested("nes", bNested, false)

	require.Equal(t, bNested, bRoot.GetNestedScheme("nes"))
	require.Nil(t, bRoot.GetNestedScheme("unknown"))
}

func TestPreviousResultDamageOnReuse(t *testing.T) {
	t.Skip("under development")
	s, err := YamlToScheme(schemeStr)
	require.Nil(t, err)
	b := NewBuffer(s)
	b.Set("name", "str")
	b.Set("quantity", 42)
	bytes1, err := b.ToBytes()
	bytes1Copy := make([]byte, len(bytes1))
	copy(bytes1Copy, bytes1)
	b.Release()
	b = NewBuffer(s)
	b.Set("name", "str")
	b.Set("quantity", 43)
	_, err = b.ToBytes() // bytes1 damages here
	require.Nil(t, err)
	require.NotEqual(t, bytes1, bytes1Copy)
}

func TestReuse(t *testing.T) {
	s, err := YamlToScheme(schemeStr)
	require.Nil(t, err)

	// borrow a Buffer and fill it
	b := NewBuffer(s)
	b.Set("name", "cola")
	b.Set("price", float32(0.123))
	b.Set("quantity", int32(42))
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	bytesNew := make([]byte, len(bytes))
	copy(bytesNew, bytes)

	b.Release() // `bytes` becomes obsolete here

	// borrow one more
	b = ReadBuffer(bytesNew, s)

	require.Equal(t, "cola", b.Get("name"))
	require.Equal(t, float32(0.123), b.Get("price"))
	require.Equal(t, int32(42), b.Get("quantity"))
	b.Release()
}
func Benchmark_ArrayOfObjectsSet_Dyno(b *testing.B) {
	s := NewScheme()
	sNested := NewScheme()
	sNested.AddField("int", FieldTypeInt, false)
	s.AddNestedArray("ints", sNested, false)

	bfNested := NewBuffer(sNested)
	bfNested.Set("int", 42)
	bufs := []*Buffer{bfNested}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			bf := NewBuffer(s)
			bf.Set("ints", bufs)
			if _, err := bf.ToBytes(); err != nil {
				b.Fatal(err)
			}
			bf.Release()
		}
	})
}

func BenchmarkS_ArrayOfObjectsAppend_ToBytes_Dyno(b *testing.B) {
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

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			bf := ReadBuffer(bytes, s)
			bf.Append("ints", bufs)
			if _, err = bf.ToBytes(); err != nil {
				b.Fatal(err)
			}
			bf.Release()
		}
	})
}

func Benchmark_ArrayOfObjectsAppend_ToBytes_Flat(b *testing.B) {
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

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			bf := flatbuffers.NewBuilder(0)

			// read existing
			existingArrayOffset := flatbuffers.UOffsetT(tab.Offset(flatbuffers.VOffsetT((0+2)*2))) + tab.Pos
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
	})
}

func Benchmark_Fill_ToBytes_NestedArray_Flat(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
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
	})
}

func TestReset(t *testing.T) {
	s := NewScheme()
	s.AddField("name", FieldTypeString, false)
	s.AddField("price", FieldTypeFloat, false)
	s.AddField("quantity", FieldTypeInt, false)

	// create new empty, fill and to bytes1
	b := NewBuffer(s)
	b.Set("price", float32(0.123))
	b.Set("quantity", 1)
	bytes1, err := b.ToBytes()
	require.Nil(t, err, err)

	// create new from bytes1, modify and to bytes2
	b = ReadBuffer(bytes1, s)
	b.Set("price", float32(0.124))
	b.Set("quantity", 2)
	bytes2, err := b.ToBytes()
	require.Nil(t, err, err)

	// create new from bytes2
	b = ReadBuffer(bytes2, s)
	require.Equal(t, int32(2), b.Get("quantity"))
	require.Equal(t, float32(0.124), b.Get("price"))
	b.Set("quantity", 3) // to check modified fields clear later

	// reset to empty
	b.Reset(nil)
	bytes, err := b.ToBytes()
	require.Nil(t, err)
	// empty object -> nothing to store
	require.Nil(t, bytes)
	require.Empty(t, b.GetNames())

	// reset to bytes1
	b.Reset(bytes1)
	require.Equal(t, int32(1), b.Get("quantity"))
	require.Equal(t, float32(0.123), b.Get("price"))

	//check modified fields are cleared
	bytes1, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes1, s)
	require.Equal(t, int32(1), b.Get("quantity"))

	// check modifiedFields is not broken
	b.Set("quantity", 5)
	bytes1, err = b.ToBytes()
	require.Nil(t, err)
	b = ReadBuffer(bytes1, s)
	require.Equal(t, int32(5), b.Get("quantity"))
}
