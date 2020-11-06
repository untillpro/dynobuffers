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

func BenchmarkPBill_ApplyMap(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)
	bb := dynobuffers.NewBuffer(s)
	fillBuffer(bb)
	jsonBytes := bb.ToJSON()
	dest := map[string]interface{}{}
	require.Nil(b, json.Unmarshal(jsonBytes, &dest))

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			pb := dynobuffers.NewBuffer(s)
			if err = pb.ApplyMap(dest); err != nil {
				b.Fatal(err)
			}
			if _, err = pb.ToBytes(); err != nil {
				b.Fatal(err)
			}
			pb.Release()
		}
	})
}

func BenchmarkPBill_ApplyMap_Append(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)
	pb := dynobuffers.NewBuffer(s)
	fillBuffer(pb)
	jsonBytes := []byte(pb.ToJSON())
	pb = dynobuffers.NewBuffer(s)

	bytes, err := pb.ApplyJSONAndToBytes(jsonBytes)
	require.Nil(b, err)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		dest := map[string]interface{}{}
		require.Nil(b, json.Unmarshal(jsonBytes, &dest))
		for p.Next() {
			pb := dynobuffers.ReadBuffer(bytes, s)
			if err = pb.ApplyMap(dest); err != nil {
				b.Fatal(err)
			}
			if _, err = pb.ToBytes(); err != nil {
				b.Fatal(err)
			}
			pb.Release()
		}
	})
}

func BenchmarkPBill_ItemRead_ByIndex(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)
	pb := dynobuffers.NewBuffer(s)
	fillBuffer(pb)
	bytes, err := pb.ToBytes()
	require.Nil(b, err)

	pb = dynobuffers.ReadBuffer(bytes, s)
	pbillItem := pb.GetByIndex("pbill_item", 0).(*dynobuffers.Buffer)
	pbillItems := []*dynobuffers.Buffer{}
	for i := 0; i < 9; i++ {
		pbillItems = append(pbillItems, pbillItem)
	}
	pb.Append("pbill_item", pbillItems)
	bytes, err = pb.ToBytes()
	require.Nil(b, err)
	pb = dynobuffers.ReadBuffer(bytes, s)
	assert.Equal(b, 10, pb.Get("pbill_item").(*dynobuffers.ObjectArray).Len)
	sum := float32(0)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			pb := dynobuffers.ReadBuffer(bytes, s)
			for i := 0; i < 10; i++ {
				pbillItem := pb.GetByIndex("pbill_item", i).(*dynobuffers.Buffer)
				s, _ := pbillItem.GetFloat("price")
				sum += s
				pbillItem.Release()
			}
			pb.Release()
		}
	})
	_ = sum
}

func BenchmarkPBill_ItemRead_Iter(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)
	pb := dynobuffers.NewBuffer(s)
	fillBuffer(pb)
	bytes, err := pb.ToBytes()
	require.Nil(b, err)
	pb = dynobuffers.ReadBuffer(bytes, s)
	pbillItem := pb.GetByIndex("pbill_item", 0).(*dynobuffers.Buffer)
	pbillItems := []*dynobuffers.Buffer{}
	for i := 0; i < 9; i++ {
		pbillItems = append(pbillItems, pbillItem)
	}
	pb.Append("pbill_item", pbillItems)
	bytes, err = pb.ToBytes()
	require.Nil(b, err)

	pb = dynobuffers.ReadBuffer(bytes, s)
	assert.Equal(b, 10, pb.Get("pbill_item").(*dynobuffers.ObjectArray).Len)

	sum := float32(0)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			pb := dynobuffers.ReadBuffer(bytes, s)
			pbillItems := pb.Get("pbill_item").(*dynobuffers.ObjectArray)
			for pbillItems.Next() {
				s, _ := pbillItems.Buffer.GetFloat("price")
				sum += s
			}
			pbillItems.Release()
			pb.Release()
		}
	})
	_ = sum
}

func BenchmarkPBillItem_Read_NoAlloc(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)

	pb := dynobuffers.NewBuffer(s)

	fillBuffer(pb)
	bytes, err := pb.ToBytes()
	require.Nil(b, err)

	pb = dynobuffers.ReadBuffer(bytes, s)
	pbillItem := pb.GetByIndex("pbill_item", 0).(*dynobuffers.Buffer)
	pbillItems := []*dynobuffers.Buffer{}
	for i := 0; i < 9; i++ {
		pbillItems = append(pbillItems, pbillItem)
	}
	pb.Append("pbill_item", pbillItems)

	bytes, err = pb.ToBytes()
	require.Nil(b, err)

	pb = dynobuffers.ReadBuffer(bytes, s)
	assert.Equal(b, 10, pb.Get("pbill_item").(*dynobuffers.ObjectArray).Len)

	sum := float32(0)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			pb := dynobuffers.ReadBuffer(bytes, s)
			arr := pb.Get("pbill_item").(*dynobuffers.ObjectArray)
			for arr.Next() {
				s, _ := arr.Buffer.GetFloat("price")
				sum += s
			}
			arr.Release()
			pb.Release()
		}
	})
	_ = sum
}

func BenchmarkPbill_Json_ReadWrite(b *testing.B) {
	s, err := dynobuffers.YamlToScheme(pbillYaml)
	require.Nil(b, err)

	pb := dynobuffers.NewBuffer(s)

	fillBuffer(pb)

	jsonBytes := pb.ToJSON()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		dest := map[string]interface{}{}
		for p.Next() {
			if err = json.Unmarshal(jsonBytes, &dest); err != nil {
				b.Fatal(err)
			}
			if _, err = json.Marshal(dest); err != nil {
				b.Fatal(err)
			}
		}
	})
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
