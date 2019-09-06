package dynobuffers

import (
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/stretchr/testify/assert"
)

type byVal struct {
	_t flatbuffers.Table
}

func Benchmark0Allocs(b *testing.B) {
	res := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := getX()
		res += int(x._t.Pos)
	}
	assert.True(b, res > 0)
}

func Benchmark1Alloc(b *testing.B) {
	res := 0
	var x *byVal

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = getX()
		res += int(x._t.Pos)
	}
	assert.True(b, res > 0)
	// assert.True(b, res != 0)
}

func getX() *byVal {
	res := &byVal{}
	res._t.Pos = 1
	return res
}

func BenchmarkReadTyped(b *testing.B) {
	s, _ := YamlToSchema(schemaStrNew)
	bf := NewBuffer(s)
	bf.Set("name", "cola")
	bf.Set("price", float32(0.123))
	bf.Set("quantity", int32(42))
	bytes := bf.ToBytes()

	b.ResetTimer()
	sum := float32(0)
	for i := 0; i < b.N; i++ {
		bf := ReadBuffer(bytes, s)
		price, _ := bf.GetFloat("price")
		quantity, _ := bf.GetInt("quantity")
		sum += price * float32(quantity)
	}
}

func BenchmarkReadUntyped(b *testing.B) {
	s, _ := YamlToSchema(schemaStrNew)
	bf := NewBuffer(s)
	bf.Set("name", "cola")
	bf.Set("price", float32(0.123))
	bf.Set("quantity", int32(42))
	bytes := bf.ToBytes()

	b.ResetTimer()
	sum := float32(0)
	for i := 0; i < b.N; i++ {
		bf := ReadBuffer(bytes, s)
		intf := bf.Get("price")
		price := intf.(float32)
		intf = bf.Get("quantity")
		quantity := intf.(int32)
		sum += price * float32(quantity)
	}
}
