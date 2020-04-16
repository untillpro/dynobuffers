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

	"github.com/linkedin/goavro"
	"github.com/untillpro/dynobuffers"
)

/*
Prior optimization
BenchmarkWriteSimpleDyno-8   	  667756	      1518 ns/op	     520 B/op	      16 allocs/op
*/

func BenchmarkWriteSimple_DynoReset(b *testing.B) {
	s := getSimpleScheme()
	bf := dynobuffers.NewBuffer(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Set("name", "cola")
		bf.Set("price", float32(0.123))
		bf.Set("quantity", int32(42))
		_, err := bf.ToBytes()
		if err != nil {
			b.Fatal(err)
		}
		bf.Reset()
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
	}
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
