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

type SafeCounter struct {
	v   int
	mux sync.Mutex
}

func (c *SafeCounter) Value() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.v
}

func (c *SafeCounter) Set(val int) {
	c.mux.Lock()
	c.v = val
	c.mux.Unlock()
}

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

var UOffsetSliceMax = SafeCounter{}
var UOffsetPool = sync.Pool{
	New: func() interface{} {
		l := UOffsetSliceMax.Value()

		return &UOffsetSliceC{
			Slice: make([]flatbuffers.UOffsetT, l),
		}
	},
}

func getUOffsetSlice(l int) (cont *UOffsetSliceC) {
	max := UOffsetSliceMax.Value()

	if l > max {
		max = l
		UOffsetSliceMax.Set(l)
	}

	cont = UOffsetPool.Get().(*UOffsetSliceC)

	if cap(cont.Slice) < max {
		cont.Slice = make([]flatbuffers.UOffsetT, max)
	} else {
		cont.Slice = cont.Slice[:l]

		for k := range cont.Slice {
			cont.Slice[k] = 0
		}
	}

	return
}

func putUOffsetSlice(cont *UOffsetSliceC) {
	max := UOffsetSliceMax.Value()
	c := cap(cont.Slice)
	if c >= max {
		UOffsetPool.Put(cont)

		if c > max {
			UOffsetSliceMax.Set(c)
		}
	}
}

//getUOffsetSlice

type OffsetSliceC struct {
	Slice []offset
}

var OffsetSliceMax = SafeCounter{}
var OffsetPool = sync.Pool{
	New: func() interface{} {
		l := OffsetSliceMax.Value()

		return &OffsetSliceC{

			Slice: make([]offset, l),
		}
	},
}

func getOffsetSlice(l int) (cont *OffsetSliceC) {
	max := OffsetSliceMax.Value()

	if l > max {
		max = l
		OffsetSliceMax.Set(l)
	}

	cont = OffsetPool.Get().(*OffsetSliceC)

	if cap(cont.Slice) < max {
		cont.Slice = make([]offset, max)
	} else {
		cont.Slice = cont.Slice[:l]

		for k := range cont.Slice {
			cont.Slice[k] = offset{}
		}
	}

	return
}

func putOffsetSlice(cont *OffsetSliceC) {
	max := OffsetSliceMax.Value()
	c := cap(cont.Slice)

	if c >= max {
		OffsetPool.Put(cont)

		if c > max {
			OffsetSliceMax.Set(c)
		}
	}
}

//getUOffsetSlice

type BuffersSliceC struct {
	Slice []*Buffer
}

var BuffersSliceMax = SafeCounter{}

var BuffersContPool = sync.Pool{
	New: func() interface{} {
		return &BuffersSliceC{}
	},
}

var BuffersPool = sync.Pool{
	New: func() interface{} {
		l := BuffersSliceMax.Value()

		return &BuffersSliceC{
			Slice: make([]*Buffer, l),
		}
	},
}

func getBufferSlice(l int) (b []*Buffer) {

	max := BuffersSliceMax.Value()

	if l > max {
		max = l
		BuffersSliceMax.Set(l)
	}

	c := BuffersPool.Get().(*BuffersSliceC)
	b = c.Slice
	c.Slice = nil
	BuffersContPool.Put(c)

	if cap(b) < max {
		b = make([]*Buffer, max)
	} else {
		b = b[:l]

		for k := range b {
			b[k] = nil
		}
	}

	return
}

func putBufferSlice(b []*Buffer) {
	max := BuffersSliceMax.Value()
	l := cap(b)

	if l >= max {
		c := BuffersContPool.Get().(*BuffersSliceC)
		c.Slice = b

		BuffersPool.Put(c)

		if l > max {
			BuffersSliceMax.Set(l)
		}

	}
}

//getStringSlice

type StringsSliceC struct {
	Slice []string
}

var StringsSliceMax = SafeCounter{}

var StringsContPool = sync.Pool{
	New: func() interface{} {
		return &StringsSliceC{}
	},
}

var StringsPool = sync.Pool{
	New: func() interface{} {
		l := StringsSliceMax.Value()

		return &StringsSliceC{
			Slice: make([]string, l),
		}
	},
}

func getStringSlice(l int) (b []string) {
	max := StringsSliceMax.Value()

	if l > max {
		max = l
		StringsSliceMax.Set(l)
	}

	c := StringsPool.Get().(*StringsSliceC)
	b = c.Slice
	c.Slice = nil
	StringsContPool.Put(c)

	if cap(b) < max {
		b = make([]string, max)
	} else {
		b = b[:l]

		for k := range b {
			b[k] = ""
		}
	}

	return
}

func putStringSlice(b []string) {
	max := StringsSliceMax.Value()
	l := cap(b)

	if l >= max {
		c := StringsContPool.Get().(*StringsSliceC)
		c.Slice = b

		StringsPool.Put(c)

		if l > max {
			StringsSliceMax.Set(l)
		}
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
