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
	"github.com/untillpro/gojay"
)

const defaultBufferSize = 10

type offset struct {
	str flatbuffers.UOffsetT
	obj flatbuffers.UOffsetT
	arr flatbuffers.UOffsetT
}

// Begin pools

var (
	builderPool = sync.Pool{
		New: func() interface{} { return flatbuffers.NewBuilder(0) },
	}
	bufferPool = sync.Pool{
		New: func() interface{} {
			return &Buffer{
				builder:        flatbuffers.NewBuilder(0),
				modifiedFields: make([]*modifiedField, defaultBufferSize),
			}
		},
	}
	objectArrayPool = sync.Pool{
		New: func() interface{} { return &ObjectArray{} },
	}
)

func getBuffer() (b *Buffer) {
	b = bufferPool.Get().(*Buffer)
	return
}

func putBuffer(b *Buffer) {
	bufferPool.Put(b)
}

//getUOffsetSlice

type uOffsetSlice struct {
	Slice []flatbuffers.UOffsetT
}

var uOffsetPool = sync.Pool{
	New: func() interface{} {
		return &uOffsetSlice{
			Slice: make([]flatbuffers.UOffsetT, defaultBufferSize),
		}
	},
}

func getUOffsetSlice(l int) (c *uOffsetSlice) {
	c = uOffsetPool.Get().(*uOffsetSlice)

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

func putUOffsetSlice(c *uOffsetSlice) {
	uOffsetPool.Put(c)
}

//getUOffsetSlice

type offsetSlice struct {
	Slice []offset
}

var offsetPool = sync.Pool{
	New: func() interface{} {
		return &offsetSlice{
			Slice: make([]offset, defaultBufferSize),
		}
	},
}

func getOffsetSlice(l int) (c *offsetSlice) {
	c = offsetPool.Get().(*offsetSlice)

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

func putOffsetSlice(c *offsetSlice) {
	offsetPool.Put(c)
}

//getUOffsetSlice

type buffersSlice struct {
	Slice []*Buffer

	Owner  *Buffer
	Scheme *Scheme
}

func (b *buffersSlice) UnmarshalJSONArray(dec *gojay.Decoder) error {
	buffer := NewBuffer(b.Scheme)
	buffer.owner = b.Owner

	if err := dec.Object(buffer); err != nil {
		return err
	}

	b.Slice = append(b.Slice, buffer)

	return nil
}

var buffersPool = sync.Pool{
	New: func() interface{} {
		return &buffersSlice{
			Slice: make([]*Buffer, defaultBufferSize),
		}
	},
}

func getBufferSlice(l int) (b *buffersSlice) {
	b = buffersPool.Get().(*buffersSlice)

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

func putBufferSlice(b *buffersSlice) {
	b.Scheme = nil
	b.Owner = nil

	buffersPool.Put(b)
}
