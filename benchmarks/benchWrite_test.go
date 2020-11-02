/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package benchmarks

import (
	"encoding/json"
	"fmt"
	"testing"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/linkedin/goavro"
	"github.com/untillpro/dynobuffers"
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
		builder.Reset()
		err := bf.ToBytesWithBuilder(builder)
		if err != nil {
			b.Fatal(err)
		}
		bf.Reset(nil)
	}
}

func BenchmarkWriteSimple_Dyno_ToBytesBuffered(b *testing.B) {
	s := getSimpleScheme()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf := dynobuffers.NewBuffer(s)
		bf.Set("name", "cola")
		bf.Set("price", float32(0.123))
		bf.Set("quantity", int32(42))
		if bytes, err := bf.ToBytesBuffered(); err != nil {
			b.Fatal(err)
		} else {
			_ = bytes
		}
		bf.Release()
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
		err := bf.ToBytesWithBuilder(builder)

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
		bf.Reset(nil)
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

/*
{"viewMods": [{"viewType": "usertable","partitionKey":{"value": "user40"},"clusterKey":{"value": "52466453699787802"},"values":{"field0":0[u+S=3P#)&v3Uc\"/60'z&^{17>9Po%Z%1C-06d>C-&D+0Rm0^)5!d3::570;Ri7Y'5Ow*&|.%h?=l:A#<Jk ^7!$p:6v6Ww:Gm;","field1": ")\"60A?+Wg\"Ew;Tk:<v7D5$Oc*@u:-p\"T/%E/'K75_-0_%9Jc8L}\"P58Xi6N2Uo;^q(]')D%1b;W7.S;89n/; ) ~4H10[=.#b:","field2": ";P9 !>-)0!; # $(+>8Wm3Cs<#h*Q/5N18A;$9v7D-19($H/$$d*I{>T5):f?@5=Aw$ :49n68\"1&65Uc;+x3Ay+[o?])(((;2b>","field3": ";P3 ]7 >v/P1 Q55J'9Mq6O} &\"$P{/+|2R}#6v/4~6R9)Ky.Zw/:*3A/<Y},Yk/(>?V3%8|5:b.J{#3x+^+03 4Vg#Og?/p&W=<","field4": "9Hq?B{\"- <Qa4Y-(Ku$<60#d##p*?6&+x/..&5l:0v'T3-Cw:., '(( :4Zw9; 8Um 80=<>7L%9+l6R?1(z??x;.4$Bu2W5%Mw7","field5": "$Q3$22;=t:@k#U),5z%Yy'Ke(5r:S%?G!<Hg1Ii2Fi26p*E?8<j2\"8W-126.D%)&v1!v:]w , $Bq/04 Wk%+$5Ia,$$)3x4^#4","field6": "<M#)*b'U!4L?;>x#.l766,^a*@;9Qa%*+;n(C7$\\q9=3&|9I5$/f,S5(/f7T38]31Wg;:,(&:)(r%#|'=f3Sw#M1/G+0V7.$n>","field7": "35\"-H'1]};..1[!5Q7$$( Fk5C5!?.\"1>)/$%Tw\"E}$Mq\"\"r2Ia?Ni$\",<Xi#9\"5!t\"Be:U+#/f+-~51.,N{8\\{:I%5Fc*Hi1=,%","field8": ";5$<Ce4O%1:<.=:,8~:-|++()Ie65b4-t6Mw#2b\"6r9M5(\"h756,\\y8^k>\"~<M)(V1$Ig:. 5;44Rg6@50 $890<.d(+*8>t*Z+-","field9": "&Co3^g.+ ,.\"-Mc;,d!/z\"(48[u+_!=b\"Fu4#8-/z0Wg&W-%.r3+l'/r3/4'^3%(()Z%)2.%\"|#Bw6Y)5!*1'V9!J#93*9Ea7"}}]}
*/

// apply map
var testData = []byte(`
{
	"viewMods": [{
		"viewType": "usertable",
		"partitionKey": {
			"value": "user40"
		},
		"clusterKey": {
			"value": "52466453699787802"
		},
		"values": {
			"field0": "0[u+S=3P#)&v3Uc\"/60'z&^{17>9Po%Z%1C-06d>C-&D+0Rm0^)5!d3::570;Ri7Y'5Ow*&|.%h?=l:A#<Jk ^7!$p:6v6Ww:Gm;",
			"field1": ")\"60A?+Wg\"Ew;Tk:<v7D5$Oc*@u:-p\"T/%E/'K75_-0_%9Jc8L}\"P58Xi6N2Uo;^q(]')D%1b;W7.S;89n/; ) ~4H10[=.#b:",
			"field2": ";P9 !>-)0!; # $(+>8Wm3Cs<#h*Q/5N18A;$9v7D-19($H/$$d*I{>T5):f?@5=Aw$ :49n68\"1&65Uc;+x3Ay+[o?])(((;2b>",
			"field3": ";P3 ]7 >v/P1 Q55J'9Mq6O} &\"$P{/+|2R}#6v/4~6R9)Ky.Zw/:*3A/<Y},Yk/(>?V3%8|5:b.J{#3x+^+03 4Vg#Og?/p&W=<",
			"field4": "9Hq?B{\"- <Qa4Y-(Ku$<60#d##p*?6&+x/..&5l:0v'T3-Cw:., '(( :4Zw9; 8Um 80=<>7L%9+l6R?1(z??x;.4$Bu2W5%Mw7",
			"field5": "$Q3$22;=t:@k#U),5z%Yy'Ke(5r:S%?G!<Hg1Ii2Fi26p*E?8<j2\"8W-126.D%)&v1!v:]w , $Bq/04 Wk%+$5Ia,$$)3x4^#4",
			"field6": "<M#)*b'U!4L?;>x#.l766,^a*@;9Qa%*+;n(C7$\\q9=3&|9I5$/f,S5(/f7T38]31Wg;:,(&:)(r%#|'=f3Sw#M1/G+0V7.$n>",
			"field7": "35\"-H'1]};..1[!5Q7$$( Fk5C5!?.\"1>)/$%Tw\"E}$Mq\"\"r2Ia?Ni$\",<Xi#9\"5!t\"Be:U+#/f+-~51.,N{8\\{:I%5Fc*Hi1=,%",
			"field8": ";5$<Ce4O%1:<.=:,8~:-|++()Ie65b4-t6Mw#2b\"6r9M5(\"h756,\\y8^k>\"~<M)(V1$Ig:. 5;44Rg6@50 $890<.d(+*8>t*Z+-",
			"field9": "&Co3^g.+ ,.\"-Mc;,d!/z\"(48[u+_!=b\"Fu4#8-/z0Wg&W-%.r3+l'/r3/4'^3%(()Z%)2.%\"|#Bw6Y)5!*1'V9!J#93*9Ea7"
		}
	}]
}
`)

func BenchmarkWriteNestedSimple_ApplyMap_withUnmarshal_Test(b *testing.B) {
	d := map[string]interface{}{}
	s := getNestedScheme()

	bf := dynobuffers.NewBuffer(s)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := json.Unmarshal(testData, &d)

		if err != nil {
			b.Fatal(err)
		}

		if err := bf.ApplyMap(d); err != nil {
			b.Fatal(err)
		}

		bf.Reset(nil)
	}

	b.StopTimer()
}

func BenchmarkWriteNestedSimple_ApplyMapBuffer_Test(b *testing.B) {
	s := getNestedScheme()

	bf := dynobuffers.NewBuffer(s)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := bf.ApplyMapBuffer(testData); err != nil {
			b.Fatal(err)
		}

		bf.Reset(nil)
	}

	b.StopTimer()
}
