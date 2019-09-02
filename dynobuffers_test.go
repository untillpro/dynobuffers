package dynobuffers

import (
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

func TestBasicUsage(t *testing.T) {
	s, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}

	// create new from sratch
	b := NewBuffer(s)
	err = b.Set("name", "cola")
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("price", float32(0.123))
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("quantity", int32(42))
	if err != nil {
		t.Fatal(err)
	}
	bytes := b.ToBytes()

	// create from bytes
	b = ReadBuffer(bytes, s)

	actual, _ := b.Get("name")
	assert.Equal(t, "cola", actual.(string))
	actual, _ = b.Get("price")
	assert.Equal(t, float32(0.123), actual.(float32))
	actual, _ = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))

	// modify existing
	b.Set("price", float32(0.124))
	bytes = b.ToBytes()
	b = ReadBuffer(bytes, s)
	actual, _ = b.Get("name")
	assert.Equal(t, "cola", actual.(string))
	actual, _ = b.Get("price")
	assert.Equal(t, float32(0.124), actual.(float32))
	actual, _ = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))

	actual, isSet := b.Get("unknownField")
	assert.Nil(t, actual)
	assert.False(t, isSet)

	// errors
	// unsupported type
	err = b.Set("name", int(1))
	assert.NotNil(t, err)
	// type mismatch
	err = b.Set("name", int32(1))
	assert.NotNil(t, err)
	// unknown field
	err = b.Set("unknownField", int32(1))
	assert.NotNil(t, err)
}

func TestSetNullValue(t *testing.T) {
	s, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(s)
	err = b.Set("name", "cola")
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("price", nil)
	if err != nil {
		t.Fatal(err)
	}
	bytes := b.ToBytes()
	b = ReadBuffer(bytes, s)
	actual, _ := b.Get("name")
	assert.Equal(t, string("cola"), actual)
	actual, isSet := b.Get("price")
	assert.Nil(t, actual)
	assert.True(t, isSet)

	// test set null in existing
	b.Set("name", nil)
	bytes = b.ToBytes()
	b = ReadBuffer(bytes, s)
	actual, isSet = b.Get("name")
	assert.Nil(t, actual)
	assert.True(t, isSet)
	actual, isSet = b.Get("price")
	assert.Nil(t, actual)
	assert.True(t, isSet)
}

func TestNonSetField(t *testing.T) {
	s, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(s)
	bytes := b.ToBytes()
	b = ReadBuffer(bytes, s)
	actual, isSet := b.Get("name")
	assert.Nil(t, actual)
	assert.False(t, isSet)
	actual, isSet = b.Get("price")
	assert.Nil(t, actual)
	assert.False(t, isSet)
}

func TestUnsetField(t *testing.T) {
	s, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(s)
	err = b.Set("name", "cola")
	if err != nil {
		t.Fatal(err)
	}
	bytes := b.ToBytes()
	b = ReadBuffer(bytes, s)
	b.Unset("name")
	bytes = b.ToBytes()
	b = ReadBuffer(bytes, s)
	actual, isSet := b.Get("name")
	assert.Nil(t, actual)
	assert.False(t, isSet)
}

func TestWriteNewReadOld(t *testing.T) {
	schemaNew, err := YamlToSchema(schemaStrNew)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(schemaNew)
	err = b.Set("name", "cola")
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("price", float32(0.123))
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("quantity", int32(42))
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("newField", int64(1))
	if err != nil {
		t.Fatal(err)
	}
	bytesNew := b.ToBytes()

	schemaOld, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytesNew, schemaOld)

	actual, _ := b.Get("name")
	assert.Equal(t, "cola", actual.(string))
	actual, _ = b.Get("price")
	assert.Equal(t, float32(0.123), actual.(float32))
	actual, _ = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))

	actual, isSet := b.Get("newField")
	assert.Nil(t, actual)
	assert.False(t, isSet)
}

func TestWriteOldReadNew(t *testing.T) {
	schemaOld, err := YamlToSchema(schemaStr)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(schemaOld)
	err = b.Set("name", "cola")
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("price", float32(0.123))
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("quantity", int32(42))
	if err != nil {
		t.Fatal(err)
	}
	bytesOld := b.ToBytes()

	schemaNew, err := YamlToSchema(schemaStrNew)
	if err != nil {
		t.Fatal(err)
	}
	b = ReadBuffer(bytesOld, schemaNew)

	actual, _ := b.Get("name")
	assert.Equal(t, "cola", actual.(string))
	actual, _ = b.Get("price")
	assert.Equal(t, float32(0.123), actual.(float32))
	actual, _ = b.Get("quantity")
	assert.Equal(t, int32(42), actual.(int32))

	actual, isSet := b.Get("newField")
	assert.Nil(t, actual)
	assert.False(t, isSet)
}

func TestYamlToSchemaErrors(t *testing.T) {
	_, err := YamlToSchema("wrong yaml")
	assert.NotNil(t, err)
	_, err = YamlToSchema("name: wrongType")
	assert.NotNil(t, err)
}

func TestFieldTypes(t *testing.T) {
	s, err := YamlToSchema(`
int: int
long: long
float: float
double: double
string: string
boolTrue: bool
boolFalse: bool
byte: byte
`)
	if err != nil {
		t.Fatal(err)
	}
	b := NewBuffer(s)
	err = b.Set("int", int32(1))
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("long", int64(2))
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("float", float32(3))
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("double", float64(4))
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("string", "str")
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("boolFalse", false)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("boolTrue", true)
	if err != nil {
		t.Fatal(err)
	}
	err = b.Set("byte", byte(5))
	if err != nil {
		t.Fatal(err)
	}
	bytes := b.ToBytes()
	b = ReadBuffer(bytes, s)
	actual, _ := b.Get("int")
	assert.Equal(t, int32(1), actual)
	actual, _ = b.Get("long")
	assert.Equal(t, int64(2), actual)
	actual, _ = b.Get("float")
	assert.Equal(t, float32(3), actual)
	actual, _ = b.Get("double")
	assert.Equal(t, float64(4), actual)
	actual, _ = b.Get("string")
	assert.Equal(t, "str", actual)
	actual, _ = b.Get("byte")
	assert.Equal(t, byte(5), actual)
	actual, _ = b.Get("boolTrue")
	assert.Equal(t, true, actual)
	actual, _ = b.Get("boolFalse")
	assert.Equal(t, false, actual)
}

func Benchmark(b *testing.B) {
	b.StopTimer()
	s, _ := YamlToSchema(schemaStrNew)
	bf := NewBuffer(s)
	bf.Set("name", "cola")
	bf.Set("price", float32(0.123))
	bf.Set("quantity", int32(42))
	bytes := bf.ToBytes()

	b.StartTimer()
	bf = ReadBuffer(bytes, s)
	sum := float32(0)
	for i := 0; i < b.N; i++ {
		intf, _ := bf.Get("price")
		price := intf.(float32)
		intf, _ = bf.Get("quantity")
		quantity := intf.(int32)
		sum += price * float32(quantity)
		// p.Set("newField", 1)
		// p.ToBytes()
	}
}

// for debug purposes
func bits(b byte) string {
	res := ""
	for i := 0; i < 8; i++ {
		if hasBit([]byte{b}, uint(i)) {
			res += "1"
		} else {
			res += "0"
		}
	}
	return res
}
