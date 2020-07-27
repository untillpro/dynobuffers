/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"sync"

	flatbuffers "github.com/google/flatbuffers/go"
)

const DefaultBufferSize = 10

var BuilderPool = sync.Pool{
	New: func() interface{} { return flatbuffers.NewBuilder(0) },
}

var BufferPool = sync.Pool{
	New: func() interface{} {
		return &Buffer{
			builder:        flatbuffers.NewBuilder(0),
			modifiedFields: make([]*modifiedField, DefaultBufferSize),
		}
	},
}

func getBuffer() (b *Buffer) {
	b = BufferPool.Get().(*Buffer)
	return
}

//getUOffsetSlice

type UOffsetSliceC struct {
	Slice []flatbuffers.UOffsetT
}

var UOffsetPool = sync.Pool{
	New: func() interface{} {
		return &UOffsetSliceC{
			Slice: make([]flatbuffers.UOffsetT, DefaultBufferSize),
		}
	},
}

func getUOffsetSlice(l int) (cont *UOffsetSliceC) {
	cont = UOffsetPool.Get().(*UOffsetSliceC)

	if cap(cont.Slice) < l {
		cont.Slice = make([]flatbuffers.UOffsetT, l)
	} else {
		cont.Slice = cont.Slice[:l]

		for k := range cont.Slice {
			cont.Slice[k] = 0
		}
	}

	return
}

func putUOffsetSlice(cont *UOffsetSliceC) {
	UOffsetPool.Put(cont)
}

//getUOffsetSlice

type OffsetSliceC struct {
	Slice []offset
}

var OffsetPool = sync.Pool{
	New: func() interface{} {
		return &OffsetSliceC{
			Slice: make([]offset, DefaultBufferSize),
		}
	},
}

func getOffsetSlice(l int) (cont *OffsetSliceC) {
	cont = OffsetPool.Get().(*OffsetSliceC)

	if cap(cont.Slice) < l {
		cont.Slice = make([]offset, l)
	} else {
		cont.Slice = cont.Slice[:l]

		for k := range cont.Slice {
			cont.Slice[k] = offset{}
		}
	}

	return
}

func putOffsetSlice(cont *OffsetSliceC) {
	OffsetPool.Put(cont)
}

//getUOffsetSlice

type BuffersSliceC struct {
	Slice []*Buffer
}

var BuffersContPool = sync.Pool{
	New: func() interface{} {
		return &BuffersSliceC{}
	},
}

var BuffersPool = sync.Pool{
	New: func() interface{} {
		return &BuffersSliceC{
			Slice: make([]*Buffer, DefaultBufferSize),
		}
	},
}

func getBufferSlice(l int) (b []*Buffer) {
	c := BuffersPool.Get().(*BuffersSliceC)
	b = c.Slice
	c.Slice = nil
	BuffersContPool.Put(c)

	if cap(b) < l {
		b = make([]*Buffer, l)
	} else {
		b = b[:l]

		for k := range b {
			b[k] = nil
		}
	}

	return
}

func putBufferSlice(b []*Buffer) {
	c := BuffersContPool.Get().(*BuffersSliceC)
	c.Slice = b

	BuffersPool.Put(c)
}

//getStringSlice

type StringsSlice struct {
	Slice []string
}

var StringsContPool = sync.Pool{
	New: func() interface{} {
		return &StringsSlice{}
	},
}

var StringsPool = sync.Pool{
	New: func() interface{} {
		return &StringsSlice{
			Slice: make([]string, DefaultBufferSize),
		}
	},
}

func getStringSlice(l int) (b *StringsSlice) {

	b = StringsPool.Get().(*StringsSlice)

	if cap(b.Slice) < l {
		b.Slice = make([]string, l)
	} else {
		b.Slice = b.Slice[:l]

		for k := range b.Slice {
			b.Slice[k] = ""
		}
	}

	return
}

func putStringSlice(b *StringsSlice) {
	StringsPool.Put(b)
}

/*



var StringSlice = make([]string, MaxBufLen)
var StringsCursor = 0
var NextStringsCursor = 0

func getStringSlice(l int) (buf []string) {
	NextStringsCursor = StringsCursor + l

	if NextStringsCursor >= MaxBufLen {
		StringsCursor = 0
		NextStringsCursor = l
	}

	buf = StringSlice[StringsCursor:NextStringsCursor]
	StringsCursor = NextStringsCursor

	for k := range buf {
		buf[k] = ""
	}

	return
}

*/
