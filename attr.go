package filemeta

import (
	"sync"
	"time"

	"github.com/ddirect/check"
)

const fileMetaAttr = "user.FILEMETA"

type attrSerDes struct {
	attr Attributes
	buf  []byte
}

var attrPool = sync.Pool{New: func() interface{} {
	return new(attrSerDes)
}}

type Attr struct {
	Size   int64
	TimeNs int64
	Hash   []byte
}

func readAttr(fileName string) (_ Attr, err error) {
	defer check.Recover(&err)
	asd := attrPool.Get().(*attrSerDes)
	defer attrPool.Put(asd)
	a := &asd.attr
	asd.buf = readXattr(fileName, fileMetaAttr, a, asd.buf)
	return Attr{a.Size, a.TimeNs, a.Hash}, nil
}

func (attr *Attr) write(fileName string) {
	asd := attrPool.Get().(*attrSerDes)
	defer attrPool.Put(asd)
	a := &asd.attr
	a.Size, a.TimeNs, a.Hash = attr.Size, attr.TimeNs, attr.Hash
	asd.buf = writeXattr(fileName, fileMetaAttr, a, asd.buf)
}

func (attr *Attributes) Time() time.Time {
	return time.Unix(0, attr.TimeNs)
}
