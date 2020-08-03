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

type UOffsetSlice struct {
	Slice []flatbuffers.UOffsetT
}

var UOffsetPool = sync.Pool{
	New: func() interface{} {
		return &UOffsetSlice{
			Slice: make([]flatbuffers.UOffsetT, DefaultBufferSize),
		}
	},
}

func getUOffsetSlice(l int) (c *UOffsetSlice) {
	c = UOffsetPool.Get().(*UOffsetSlice)

	if cap(c.Slice) < l {
		c.Slice = make([]flatbuffers.UOffsetT, l)
	} else {
		c.Slice = c.Slice[:l]

		for k := range c.Slice {
			c.Slice[k] = 0
		}
	}

	return
}

func putUOffsetSlice(c *UOffsetSlice) {
	UOffsetPool.Put(c)
}

//getUOffsetSlice

type OffsetSlice struct {
	Slice []offset
}

var OffsetPool = sync.Pool{
	New: func() interface{} {
		return &OffsetSlice{
			Slice: make([]offset, DefaultBufferSize),
		}
	},
}

func getOffsetSlice(l int) (c *OffsetSlice) {
	c = OffsetPool.Get().(*OffsetSlice)

	if cap(c.Slice) < l {
		c.Slice = make([]offset, l)
	} else {
		c.Slice = c.Slice[:l]

		for k := range c.Slice {
			c.Slice[k] = offset{}
		}
	}

	return
}

func putOffsetSlice(c *OffsetSlice) {
	OffsetPool.Put(c)
}

//getUOffsetSlice

type BuffersSlice struct {
	Slice []*Buffer
}

var BuffersPool = sync.Pool{
	New: func() interface{} {
		return &BuffersSlice{
			Slice: make([]*Buffer, DefaultBufferSize),
		}
	},
}

func getBufferSlice(l int) (b *BuffersSlice) {
	b = BuffersPool.Get().(*BuffersSlice)

	if cap(b.Slice) < l {
		b.Slice = make([]*Buffer, l)
	} else {
		b.Slice = b.Slice[:l]

		for k := range b.Slice {
			b.Slice[k] = nil
		}
	}

	return
}

func putBufferSlice(b *BuffersSlice) {
	BuffersPool.Put(b)
}

//getStringSlice

type StringsSlice struct {
	Slice []string
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
