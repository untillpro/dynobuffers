/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"sync"
	"sync/atomic"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/untillpro/gojay"
)

const defaultBufferSize = 10

var (
	buffersInUse      uint64
	bufferSlicesInUse uint64
	uOffsetsInUse     uint64
	offsetsInUse      uint64
	objectArraysInUse uint64
)

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
	uOffsetPool = sync.Pool{
		New: func() interface{} {
			return &uOffsetSlice{
				Slice: make([]flatbuffers.UOffsetT, defaultBufferSize),
			}
		},
	}
	offsetPool = sync.Pool{
		New: func() interface{} {
			return &offsetSlice{
				Slice: make([]offset, defaultBufferSize),
			}
		},
	}
	buffersPool = sync.Pool{
		New: func() interface{} {
			return &buffersSlice{
				Slice: make([]*Buffer, defaultBufferSize),
			}
		},
	}
)

func getBuffer() (b *Buffer) {
	b = bufferPool.Get().(*Buffer)
	atomic.AddUint64(&buffersInUse, 1)
	return
}

func putBuffer(b *Buffer) {
	bufferPool.Put(b)
	atomic.AddUint64(&buffersInUse, ^uint64(0))
}

//getUOffsetSlice

type uOffsetSlice struct {
	Slice []flatbuffers.UOffsetT
}

func getUOffsetSlice(l int) (c *uOffsetSlice) {
	c = uOffsetPool.Get().(*uOffsetSlice)
	atomic.AddUint64(&uOffsetsInUse, 1)

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
	atomic.AddUint64(&uOffsetsInUse, ^uint64(0))
}

//getUOffsetSlice

type offsetSlice struct {
	Slice []offset
}

func getOffsetSlice(l int) (c *offsetSlice) {
	c = offsetPool.Get().(*offsetSlice)
	atomic.AddUint64(&offsetsInUse, 1)

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
	atomic.AddUint64(&offsetsInUse, ^uint64(0))
}

//getUOffsetSlice

type buffersSlice struct {
	Slice      []*Buffer
	Owner      *Buffer
	Scheme     *Scheme
	isReleased bool
}

func (b *buffersSlice) UnmarshalJSONArray(dec *gojay.Decoder) error {
	buffer := NewBuffer(b.Scheme)
	buffer.owner = b.Owner

	if err := dec.Object(buffer); err != nil {
		buffer.Release()
		return err
	}

	b.Slice = append(b.Slice, buffer)

	return nil
}

func (b *buffersSlice) Release() {
	if b.isReleased {
		return
	}
	for _, buf := range b.Slice {
		// could be nil e.g. if errors during ApplyMap(): got from pool with lenght 2, 1 ok, second - failed to encode
		if buf != nil {
			buf.Release()
		}
	}
	b.Scheme = nil
	b.Owner = nil
	b.isReleased = true
	buffersPool.Put(b)
	atomic.AddUint64(&bufferSlicesInUse, ^uint64(0))
}

func getBufferSlice(l int) (b *buffersSlice) {
	b = buffersPool.Get().(*buffersSlice)
	atomic.AddUint64(&bufferSlicesInUse, 1)

	b.isReleased = false
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

func getObjectArray() *ObjectArray {
	res := objectArrayPool.Get().(*ObjectArray)
	res.isReleased = false
	atomic.AddUint64(&objectArraysInUse, 1)
	return res
}

func putObjectArray(oa *ObjectArray) {
	objectArrayPool.Put(oa)
	atomic.AddUint64(&objectArraysInUse, ^uint64(0))
}

// GetObjectsInUse returns pooled objects amount which are currently in use, i.e. not released
// useful for testing and metrics accounting
func GetObjectsInUse() uint64 {
	return atomic.LoadUint64(&bufferSlicesInUse) +
		atomic.LoadUint64(&buffersInUse) +
		atomic.LoadUint64(&offsetsInUse) +
		atomic.LoadUint64(&uOffsetsInUse) +
		atomic.LoadUint64(&objectArraysInUse)
}
