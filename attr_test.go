package filemeta

import (
	"sync"
	"testing"

	"github.com/ddirect/xrand"
	"google.golang.org/protobuf/proto"
)

func BenchmarkAttrAppend(b *testing.B) {
	buf := xrand.New().Buffer(32)
	attr := new(Attributes)
	attr.Hash = buf
	var res []byte
	for i := 0; i < b.N; i++ {
		attr.Size = int64(i)
		attr.TimeNs = int64(i)
		res, _ = proto.MarshalOptions{}.MarshalAppend(res[:0], attr)
	}
}

func BenchmarkAttrReuse(b *testing.B) {
	buf := xrand.New().Buffer(32)
	attr := new(Attributes)
	attr.Hash = buf
	for i := 0; i < b.N; i++ {
		attr.Size = int64(i)
		attr.TimeNs = int64(i)
		proto.MarshalOptions{}.MarshalAppend(buf[:0], attr)
	}
}

func BenchmarkAttrRewrite(b *testing.B) {
	buf := xrand.New().Buffer(32)
	attr := new(Attributes)
	for i := 0; i < b.N; i++ {
		attr.Hash = buf
		attr.Size = int64(i)
		attr.TimeNs = int64(i)
		proto.Marshal(attr)
	}
}

func BenchmarkAttrReallocate(b *testing.B) {
	buf := xrand.New().Buffer(32)
	for i := 0; i < b.N; i++ {
		attr := new(Attributes)
		attr.Hash = buf
		attr.Size = int64(i)
		attr.TimeNs = int64(i)
		proto.Marshal(attr)
	}
}

func BenchmarkAttrPool(b *testing.B) {
	buf := xrand.New().Buffer(32)
	pool := sync.Pool{New: func() interface{} {
		return new(Attributes)
	}}
	for i := 0; i < b.N; i++ {
		attr := pool.Get().(*Attributes)
		attr.Hash = buf
		attr.Size = int64(i)
		attr.TimeNs = int64(i)
		proto.Marshal(attr)
		pool.Put(attr)
	}
}

type Marshaler struct {
	attr Attributes
	buf  []byte
}

func BenchmarkAttrPoolAppend(b *testing.B) {
	buf := xrand.New().Buffer(32)
	pool := sync.Pool{New: func() interface{} {
		return new(Marshaler)
	}}
	for i := 0; i < b.N; i++ {
		m := pool.Get().(*Marshaler)
		m.attr.Hash = buf
		m.attr.Size = int64(i)
		m.attr.TimeNs = int64(i)
		m.buf, _ = proto.MarshalOptions{}.MarshalAppend(m.buf[:0], &m.attr)
		pool.Put(m)
	}
}
