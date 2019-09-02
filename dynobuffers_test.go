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
	err = b.Set("id", int64(1))
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
		// p.Set("id", 1)
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
