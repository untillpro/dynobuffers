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

var BuilderPool = sync.Pool{
	New: func() interface{} { return flatbuffers.NewBuilder(0) },
}

var BufferPool = sync.Pool{
	New: func() interface{} {
		return &Buffer{
			builder:        flatbuffers.NewBuilder(0),
			modifiedFields: make([]*modifiedField, 10),
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

var UOffsetSliceMax = 0
var UOffsetPool = sync.Pool{
	New: func() interface{} {
		return &UOffsetSliceC{
			Slice: make([]flatbuffers.UOffsetT, UOffsetSliceMax),
		}
	},
}

func getUOffsetSlice(l int) (cont *UOffsetSliceC) {
	if l > UOffsetSliceMax {
		UOffsetSliceMax = l
	}

	cont = UOffsetPool.Get().(*UOffsetSliceC)

	if cap(cont.Slice) < UOffsetSliceMax {
		cont.Slice = make([]flatbuffers.UOffsetT, UOffsetSliceMax)
	} else {
		cont.Slice = cont.Slice[:l]

		for k := range cont.Slice {
			cont.Slice[k] = 0
		}
	}

	return
}

func putUOffsetSlice(cont *UOffsetSliceC) {
	if cap(cont.Slice) >= UOffsetSliceMax {
		UOffsetPool.Put(cont)

		UOffsetSliceMax = cap(cont.Slice)
	}
}

//getUOffsetSlice

type OffsetSliceC struct {
	Slice []offset
}

var OffsetSliceMax = 0
var OffsetPool = sync.Pool{
	New: func() interface{} {
		return &OffsetSliceC{
			Slice: make([]offset, OffsetSliceMax),
		}
	},
}

func getOffsetSlice(l int) (cont *OffsetSliceC) {
	if l > OffsetSliceMax {
		OffsetSliceMax = l
	}

	cont = OffsetPool.Get().(*OffsetSliceC)

	if cap(cont.Slice) < OffsetSliceMax {
		cont.Slice = make([]offset, OffsetSliceMax)
	} else {
		cont.Slice = cont.Slice[:l]

		for k := range cont.Slice {
			cont.Slice[k] = offset{}
		}
	}

	return
}

func putOffsetSlice(cont *OffsetSliceC) {
	if cap(cont.Slice) >= OffsetSliceMax {
		OffsetPool.Put(cont)

		OffsetSliceMax = cap(cont.Slice)
	}
}

//getUOffsetSlice

type BuffersSliceC struct {
	Slice []*Buffer
}

var BuffersSliceMax = 0

var BuffersContPool = sync.Pool{
	New: func() interface{} {
		return &BuffersSliceC{}
	},
}

var BuffersPool = sync.Pool{
	New: func() interface{} {
		return &BuffersSliceC{
			Slice: make([]*Buffer, BuffersSliceMax),
		}
	},
}

func getBufferSlice(l int) (b []*Buffer) {
	if l > BuffersSliceMax {
		BuffersSliceMax = l
	}

	c := BuffersPool.Get().(*BuffersSliceC)
	b = c.Slice
	c.Slice = nil
	BuffersContPool.Put(c)

	if cap(b) < BuffersSliceMax {
		b = make([]*Buffer, BuffersSliceMax)
	} else {
		b = b[:l]

		for k := range b {
			b[k] = nil
		}
	}

	return
}

func putBufferSlice(b []*Buffer) {
	if cap(b) >= BuffersSliceMax {
		c := BuffersContPool.Get().(*BuffersSliceC)
		c.Slice = b

		BuffersPool.Put(c)

		BuffersSliceMax = cap(b)
	}
}

//getStringSlice

type StringsSliceC struct {
	Slice []string
}

var StringsSliceMax = 0

var StringsContPool = sync.Pool{
	New: func() interface{} {
		return &StringsSliceC{}
	},
}

var StringsPool = sync.Pool{
	New: func() interface{} {
		return &StringsSliceC{
			Slice: make([]string, StringsSliceMax),
		}
	},
}

func getStringSlice(l int) (b []string) {
	if l > StringsSliceMax {
		StringsSliceMax = l
	}

	c := StringsPool.Get().(*StringsSliceC)
	b = c.Slice
	c.Slice = nil
	StringsContPool.Put(c)

	if cap(b) < StringsSliceMax {
		b = make([]string, StringsSliceMax)
	} else {
		b = b[:l]

		for k := range b {
			b[k] = ""
		}
	}

	return
}

func putStringSlice(b []string) {
	if cap(b) >= StringsSliceMax {
		c := StringsContPool.Get().(*StringsSliceC)
		c.Slice = b

		StringsPool.Put(c)

		StringsSliceMax = cap(b)
	}
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
