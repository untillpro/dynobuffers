/*
 * Copyright (c) 2019-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package benchmarks

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/untillpro/dynobuffers"
)

func Benchmark_MapToBytes_Pbill_Dyno(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)
	bb := dynobuffers.NewBuffer(s)
	fillBuffer(bb)
	jsonBytes := bb.ToJSON()
	dest := map[string]interface{}{}
	require.Nil(b, json.Unmarshal(jsonBytes, &dest))
	bb.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			pb := dynobuffers.NewBuffer(s)
			if err := pb.ApplyMap(dest); err != nil {
				pb.Release()
				b.Fatal(err)
			}
			if _, err := pb.ToBytes(); err != nil {
				pb.Release()
				b.Fatal(err)
			}
			pb.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_MapToBytes_PBill_AppendArrays_Dyno(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)
	pb := dynobuffers.NewBuffer(s)
	fillBuffer(pb)
	jsonBytes := []byte(pb.ToJSON())
	pb.Release()
	pb = dynobuffers.NewBuffer(s)
	bytes, _, err := pb.ApplyJSONAndToBytes(jsonBytes)
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	pb.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		dest := map[string]interface{}{}
		require.Nil(b, json.Unmarshal(jsonBytes, &dest))
		for p.Next() {
			pb := dynobuffers.ReadBuffer(bytes, s)
			if err := pb.ApplyMap(dest); err != nil {
				b.Fatal(err)
			}
			if _, err := pb.ToBytes(); err != nil {
				b.Fatal(err)
			}
			pb.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_R_PbillItem_ByIndex_Dyno(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)
	pb := dynobuffers.NewBuffer(s)
	fillBuffer(pb)
	bytes, err := pb.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	pb.Release()

	pb = dynobuffers.ReadBuffer(bytes, s)
	oa := pb.Get("pbill_item").(*dynobuffers.ObjectArray)
	oa.Next()
	pbillItemBytes := oa.Buffer.GetBytes()
	pbillItemBytes = copyBytes(pbillItemBytes)
	pbillItem := dynobuffers.ReadBuffer(pbillItemBytes, oa.Buffer.Scheme)
	pbillItems := []*dynobuffers.Buffer{}
	for i := 0; i < 9; i++ {
		pbillItems = append(pbillItems, pbillItem)
	}
	pb.Append("pbill_item", pbillItems)
	bytes, err = pb.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	pb.Release()
	pb = dynobuffers.ReadBuffer(bytes, s)
	assert.Equal(b, 10, pb.Get("pbill_item").(*dynobuffers.ObjectArray).Len)
	pb.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		sum := float32(0)
		for p.Next() {
			pb := dynobuffers.ReadBuffer(bytes, s)
			oa := pb.Get("pbill_item").(*dynobuffers.ObjectArray)
			for oa.Next() {
				s, _ := oa.Buffer.GetFloat32("price")
				sum += s
			}
			pb.Release()
		}
		_ = sum
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_R_PBillItem_Iter_Dyno(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)
	pb := dynobuffers.NewBuffer(s)
	fillBuffer(pb)
	bytes, err := pb.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	pb.Release()
	pb = dynobuffers.ReadBuffer(bytes, s)
	oa := pb.Get("pbill_item").(*dynobuffers.ObjectArray)
	oa.Next()
	pbillItemBytes := oa.Buffer.GetBytes()
	pbillItemBytes = copyBytes(pbillItemBytes)
	pbillItem := dynobuffers.ReadBuffer(pbillItemBytes, oa.Buffer.Scheme)
	pbillItems := []*dynobuffers.Buffer{}
	for i := 0; i < 9; i++ {
		pbillItems = append(pbillItems, pbillItem)
	}
	pb.Append("pbill_item", pbillItems)
	bytes, err = pb.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	pb.Release()

	pb = dynobuffers.ReadBuffer(bytes, s)
	assert.Equal(b, 10, pb.Get("pbill_item").(*dynobuffers.ObjectArray).Len)
	pb.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		sum := float32(0)
		for p.Next() {
			pb := dynobuffers.ReadBuffer(bytes, s)
			pbillItems := pb.Get("pbill_item").(*dynobuffers.ObjectArray)
			for pbillItems.Next() {
				s, _ := pbillItems.Buffer.GetFloat32("price")
				sum += s
			}
			pb.Release()
		}
		_ = sum
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_RW_Pbill_Json(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)

	pb := dynobuffers.NewBuffer(s)

	fillBuffer(pb)

	jsonBytes := pb.ToJSON()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		dest := map[string]interface{}{}
		for p.Next() {
			if err := json.Unmarshal(jsonBytes, &dest); err != nil {
				b.Fatal(err)
			}
			for range dest {
			}
			if _, err := json.Marshal(dest); err != nil {
				b.Fatal(err)
			}
		}
	})
	pb.Release()
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_RW_Pbill_Dyno_AllFields(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)
	pb := dynobuffers.NewBuffer(s)
	fillBuffer(pb)
	bytes, err := pb.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	pb.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			pb := dynobuffers.ReadBuffer(bytes, s)
			pb.Set(s.Fields[0].Name, 1)
			if _, err := pb.ToBytes(); err != nil {
				b.Fatal(err)
			}
			pb.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func Benchmark_RW_Pbill_Dyno_NoRead(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)
	pb := dynobuffers.NewBuffer(s)
	fillBuffer(pb)
	bytes, err := pb.ToBytes()
	require.Nil(b, err)
	bytes = copyBytes(bytes)
	pb.Release()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			pb := dynobuffers.ReadBuffer(bytes, s)
			pb.Set(s.Fields[0].Name, 1)
			if _, err := pb.ToBytes(); err != nil {
				b.Fatal(err)
			}
			pb.Release()
		}
	})
	require.Zero(b, dynobuffers.GetObjectsInUse())
}

func fillBuffer(b *dynobuffers.Buffer) {
	for i, f := range b.Scheme.Fields {
		var val interface{}
		switch f.Ft {
		case dynobuffers.FieldTypeBool:
			val = true
		case dynobuffers.FieldTypeString:
			val = "str" + strconv.Itoa(i)
		case dynobuffers.FieldTypeObject:
			nested := dynobuffers.NewBuffer(f.FieldScheme)
			fillBuffer(nested)
			if f.IsArray {
				val = []*dynobuffers.Buffer{nested}
			} else {
				val = nested
			}
		case dynobuffers.FieldTypeByte:
			if f.IsArray {
				val = []byte{byte(i)}
			} else {
				val = i
			}
		default:
			val = i
		}
		b.Set(f.Name, val)
	}
}
