/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package benchmarks

import (
	"fmt"
	"testing"

	"github.com/Yohanson555/dynobuffers"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/linkedin/goavro"
)

func BenchmarkWriteSimple_Dyno_SameBuilder(b *testing.B) {
	s := getSimpleScheme()
	bf := dynobuffers.NewBuffer(s)
	builder := flatbuffers.NewBuilder(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Set("name", "cola")
		bf.Set("price", float32(0.123))
		bf.Set("quantity", int32(42))
		_, err := bf.ToBytesWithBuilder(builder)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteSimple_Dyno(b *testing.B) {
	s := getSimpleScheme()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := dynobuffers.NewBuffer(s)
		bf.Set("name", "cola")
		bf.Set("price", float32(0.123))
		bf.Set("quantity", int32(42))
		_, err := bf.ToBytes()
		if err != nil {
			b.Fatal(err)
		}
		bf.Release()
	}
}

func BenchmarkWriteNestedSimple_Dyno(b *testing.B) {
	s := getNestedScheme()
	data := getNestedData()
	bf := dynobuffers.NewBuffer(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Reset(nil)
		bf.ApplyMap(data)
		_, err := bf.ToBytes()

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteNestedSimple_Dyno_SameBuilder(b *testing.B) {
	s := getNestedScheme()
	data := getNestedData()
	builder := flatbuffers.NewBuilder(0)

	bf := dynobuffers.NewBuffer(s)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bf.Reset(nil)
		bf.ApplyMap(data)
		_, err := bf.ToBytesWithBuilder(builder)

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteNestedSimple_ApplyMap_Test(b *testing.B) {
	s := getNestedScheme()
	data := getNestedData()

	bf := dynobuffers.NewBuffer(s)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bf.Reset(nil)
		bf.ApplyMap(data)
	}

	b.StopTimer()
}

func BenchmarkWriteNested_ToBytes_Test(b *testing.B) {
	s := getNestedScheme()
	data := getNestedData()

	bf := dynobuffers.NewBuffer(s)

	bf.ApplyMap(data)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := bf.ToBytes()

		if err != nil {
			b.Fatal(err)
		}
	}

	b.StopTimer()
}

func BenchmarkWriteNested_ToBytes_Parallel(b *testing.B) {
	s := getNestedScheme()
	data := getNestedData()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		bf := dynobuffers.NewBuffer(s)

		bf.ApplyMap(data)

		for pb.Next() {

			_, err := bf.ToBytes()

			if err != nil {
				b.Fatal(err)
			}

		}

		bf.Release()
	})

	b.StopTimer()
}

func BenchmarkWriteSimple_Avro(b *testing.B) {
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
	if err != nil {
		fmt.Println(err)
	}

	b.ResetTimer()

	buf := make([]byte, 0)
	for i := 0; i < b.N; i++ {
		data := make(map[string]interface{})
		data["name"] = "cola"
		data["quantity"] = 1
		data["price"] = float32(0.123)
		buf, err = codec.BinaryFromNative(buf[:0], data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestWriteNestedSimple_ApplyMap_Test(b *testing.T) {
	s := getNestedScheme()
	data := getNestedData()

	bf := dynobuffers.NewBuffer(s)

	for i := 0; i < 1000; i++ {
		bf.ApplyMap(data)
		bf.Reset(nil)
	}
	bf.Release()
}
