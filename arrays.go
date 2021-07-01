/*
 * Copyright (c) 2021-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"fmt"

	flatbuffers "github.com/google/flatbuffers/go"
)

type IInt32Array interface {
	Len() int
	At(idx int) int32
}

type IInt64Array interface {
	Len() int
	At(idx int) int64
}

type IFloat32Array interface {
	Len() int
	At(idx int) float32
}

type IFloat64Array interface {
	Len() int
	At(idx int) float64
}

type IStringArray interface {
	Len() int
	At(idx int) string
}

type IByteArray interface {
	Bytes() []byte
}

type IBoolArray interface {
	Len() int
	At(idx int) bool
}

type abstractArray struct {
	len     int
	uOffset flatbuffers.UOffsetT
	tab     flatbuffers.Table
}

type implIInt32Array struct {
	abstractArray
}

type implIInt64Array struct {
	abstractArray
}

type implIFloat32Array struct {
	abstractArray
}

type implIFloat64Array struct {
	abstractArray
}

type implIByteArray struct {
	abstractArray
}

type implIStringArray struct {
	abstractArray
}

type implIBoolArray struct {
	abstractArray
}

func (a abstractArray) Len() int {
	return a.len
}

func (a abstractArray) check(idx int) {
	if idx < 0 || idx >= a.len {
		panic(fmt.Sprintf("index out of range: %d of %d", idx, a.len))
	}
}

func (i implIInt32Array) At(idx int) int32 {
	i.check(idx)
	return i.tab.GetInt32(i.uOffset + flatbuffers.UOffsetT((i.len-idx-1)*flatbuffers.SizeInt32))
}

func (i implIInt64Array) At(idx int) int64 {
	i.check(idx)
	return i.tab.GetInt64(i.uOffset + flatbuffers.UOffsetT((i.len-idx-1)*flatbuffers.SizeInt64))
}

func (i implIFloat32Array) At(idx int) float32 {
	i.check(idx)
	return i.tab.GetFloat32(i.uOffset + flatbuffers.UOffsetT((i.len-idx-1)*flatbuffers.SizeFloat32))
}

func (i implIFloat64Array) At(idx int) float64 {
	i.check(idx)
	return i.tab.GetFloat64(i.uOffset + flatbuffers.UOffsetT((i.len-idx-1)*flatbuffers.SizeFloat64))
}

func (i implIBoolArray) At(idx int) bool {
	i.check(idx)
	return i.tab.GetBool(i.uOffset + flatbuffers.UOffsetT((i.len-idx-1)*flatbuffers.SizeBool))
}

func (i implIStringArray) At(idx int) string {
	i.check(idx)
	elementUOffsetT := i.uOffset + flatbuffers.UOffsetT((i.len-idx-1)*flatbuffers.SizeUOffsetT)
	return byteSliceToString(i.tab.ByteVector(elementUOffsetT))
}

func (i implIByteArray) Bytes() []byte {
	bytes := i.tab.Bytes[i.uOffset : i.len+int(i.uOffset)]
	res := make([]byte, len(bytes))
	copy(res, bytes)
	return res
}
