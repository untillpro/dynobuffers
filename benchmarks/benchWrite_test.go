/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package benchmarks

import (
	"encoding/json"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/linkedin/goavro"
	"github.com/stretchr/testify/require"
	"github.com/untillpro/dynobuffers"
)

func Benchmark_Fill_ToBytes_Simple_Dyno_SameBuilder(b *testing.B) {
	s := getSimpleScheme()

	b.RunParallel(func(p *testing.PB) {
		builder := flatbuffers.NewBuilder(0)
		for p.Next() {
			bf := dynobuffers.NewBuffer(s)
			bf.Set("name", "cola")
			bf.Set("price", float32(0.123))
			bf.Set("quantity", int32(42))
			if err := bf.ToBytesWithBuilder(builder); err != nil {
				b.Fatal(err)
			}
			builder.Reset()
			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_Fill_ToBytes_Simple_Dyno(b *testing.B) {
	s := getSimpleScheme()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			bf := dynobuffers.NewBuffer(s)
			bf.Set("name", "cola")
			bf.Set("price", float32(0.123))
			bf.Set("quantity", int32(42))
			if _, err := bf.ToBytes(); err != nil {
				b.Fatal(err)
			}
			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_MapToBytes_Nested_Dyno(b *testing.B) {
	s := getNestedScheme()
	data := getNestedData()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			bf := dynobuffers.NewBuffer(s)
			if err := bf.ApplyMap(data); err != nil {
				b.Fatal(err)
			}
			if _, err := bf.ToBytes(); err != nil {
				b.Fatal(err)
			}
			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_MapToBytes_Nested_Dyno_SameBuilder(b *testing.B) {
	s := getNestedScheme()
	data := getNestedData()

	b.RunParallel(func(p *testing.PB) {
		builder := flatbuffers.NewBuilder(0)
		for p.Next() {
			bf := dynobuffers.NewBuffer(s)
			bf.ApplyMap(data)
			if err := bf.ToBytesWithBuilder(builder); err != nil {
				b.Fatal(err)
			}
			builder.Reset()
			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_MapToBytes_Simple_Avro(b *testing.B) {
	codec, err := goavro.NewCodec(`
		{"namespace": "unTill",
		"type": "record",
		"name": "OrderItem",
		"fields": [
			{"name": "name", "type": "string"},
			{"name": "price", "type": "float"},
			{"name": "quantity", "type": "int", "default": 0}
		]}
	`)
	require.Nil(b, err)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		buf := make([]byte, 0)
		for p.Next() {
			data := make(map[string]interface{})
			data["name"] = "cola"
			data["quantity"] = 1
			data["price"] = float32(0.123)
			if _, err := codec.BinaryFromNative(buf[:0], data); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_JSONToBytes_Nested_Dyno(b *testing.B) {
	s := getNestedScheme()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			bf := dynobuffers.NewBuffer(s)
			if err := bf.ApplyMapBuffer(testData); err != nil {
				b.Fatal(err)
			}
			if _, err := bf.ToBytes(); err != nil {
				b.Fatal(err)
			}
			bf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_JSONToBytes_Simple_Dyno(b *testing.B) {
	s := getSimpleScheme()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			buf := dynobuffers.NewBuffer(s)
			if _, _, err := buf.ApplyJSONAndToBytes([]byte(`{"name": "cola", "price": 0.123, "quantity": 1}`)); err != nil {
				b.Fatal(err)
			}
			buf.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}
func Benchmark_JSONToBytes_Simple_Avro(b *testing.B) {
	codec, err := goavro.NewCodec(`
		{"namespace": "unTill",
		"type": "record",
		"name": "OrderItem",
		"fields": [
			{"name": "name", "type": "string"},
			{"name": "price", "type": "float"},
			{"name": "quantity", "type": "int"}
		]}
	`)
	require.Nil(b, err)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			native, _, err := codec.NativeFromTextual([]byte(`{"name": "cola", "price": 0.123, "quantity": 1}`))
			if err != nil {
				b.Fatal(err)
			}

			if _, err := codec.BinaryFromNative(nil, native); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_Fill_ToBytes_Read_Simple_Dyno(bench *testing.B) {
	s := getSimpleScheme()

	bench.ResetTimer()
	bench.RunParallel(func(p *testing.PB) {
		for p.Next() {
			b := dynobuffers.NewBuffer(s)
			// str := "cola"
			// b.Set("name", str) // +1 allocation
			b.Set("name", "cola") // 0 allocations
			b.Set("price", float32(0.123))
			b.Set("quantity", int32(42))
			b.Set("unknownField", "some value") // Nothing happens here, nothing will be written to buffer
			bytes, err := b.ToBytes()
			if err != nil {
				bench.Fatal(err)
			}
			b.Reset(bytes)

			// Now we can Get fields
			str, _ := b.GetString("name")
			if str != "cola" {
				bench.Fatal()
			}
			float, _ := b.GetFloat("price")
			if float != 0.123 {
				bench.Fatal()
			}
			q, _ := b.GetInt("quantity")
			if q != 42 {
				bench.Fatal()
			}
			b.Release()
		}
	})
	require.Zero(bench, dynobuffers.GetObjectsInUse())
}

func Benchmark_ToJSON_Simple_Dyno(b *testing.B) {
	scheme := getSimpleScheme()

	buf := dynobuffers.NewBuffer(scheme)
	actual := map[string]interface{}{}
	jsonBytes := buf.ToJSON()
	json.Unmarshal(jsonBytes, &actual)
	require.True(b, len(actual) == 0)
	buf.Set("name", "cola")
	buf.Set("price", float32(0.123))
	buf.Set("quantity", int32(42))
	bytes, err := buf.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	buf.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		buf := dynobuffers.ReadBuffer(bytes, scheme)
		for p.Next() {
			buf.ToJSON()
		}
		buf.Release()
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_ToJSONMap_Simple_Dyno(b *testing.B) {
	scheme := getSimpleScheme()

	buf := dynobuffers.NewBuffer(scheme)
	actual := map[string]interface{}{}
	jsonBytes := buf.ToJSON()
	json.Unmarshal(jsonBytes, &actual)
	require.True(b, len(actual) == 0)
	buf.Set("name", "cola")
	buf.Set("price", float32(0.123))
	buf.Set("quantity", int32(42))
	bytes, err := buf.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	buf.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		buf := dynobuffers.ReadBuffer(bytes, scheme)
		for p.Next() {
			buf.ToJSONMap()
		}
		buf.Release()
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func copyBytes(src []byte) []byte {
	res := make([]byte, len(src))
	copy(res, src)
	return res
}
