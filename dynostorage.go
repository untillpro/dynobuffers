/*
 * Copyright (c) 2018-present unTill Pro, Ltd. and Contributors
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

package dynobuffers

import (
	"fmt"
	"sync"

	"github.com/Yohanson555/gojay"
	flatbuffers "github.com/google/flatbuffers/go"
)

const DefaultBufferSize = 10

var mux sync.Mutex
var poolStat map[string]int = map[string]int{}

func StatIncrement(bucket string, value int) {
	mux.Lock()
	poolStat[bucket] += value
	mux.Unlock()
}

func FlushPoolStat() {
	mux.Lock()
	poolStat = map[string]int{}
	mux.Unlock()
}

func PrintPoolStat() {
	fmt.Printf("IBuffers pool stat begins: ------------ \n\n")
	for k, v := range poolStat {
		fmt.Printf("s.%v: %v\n", k, v)
	}
}

// Begin pools

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
	b.isReleased = false
	StatIncrement("BufferGet", 1)
	return
}

func putBuffer(b *Buffer) {
	if !b.isReleased {
		b.isReleased = true
		BufferPool.Put(b)
		StatIncrement("BufferPut", 1)
	}
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

	StatIncrement("UOffsetSliceGet", 1)
	return
}

func putUOffsetSlice(c *UOffsetSlice) {
	UOffsetPool.Put(c)
	StatIncrement("UOffsetSlicePut", 1)
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

	StatIncrement("OffsetSliceGet", 1)
	return
}

func putOffsetSlice(c *OffsetSlice) {
	OffsetPool.Put(c)
	StatIncrement("OffsetSlicePut", 1)
}

//getUOffsetSlice

type BuffersSlice struct {
	Slice []*Buffer

	Owner  *Buffer
	Scheme *Scheme
}

func (b *BuffersSlice) UnmarshalJSONArray(dec *gojay.Decoder) error {
	buffer := NewBuffer(b.Scheme)
	buffer.owner = b.Owner

	if err := dec.Object(buffer); err != nil {
		return err
	}

	b.Slice = append(b.Slice, buffer)

	return nil
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

	StatIncrement("BuffersSliceGet", 1)
	return
}

func putBufferSlice(b *BuffersSlice) {
	b.Scheme = nil
	b.Owner = nil

	BuffersPool.Put(b)
	StatIncrement("BuffersSlicePut", 1)
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

	StatIncrement("StringSliceGet", 1)
	return
}

func putStringSlice(b *StringsSlice) {
	StringsPool.Put(b)
	StatIncrement("StringsSlicePut", 1)
}
