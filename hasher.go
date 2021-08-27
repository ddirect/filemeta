package filemeta

import (
	"errors"
	"hash"
	"io"
	"os"
	"sync"

	"github.com/ddirect/check"
	"golang.org/x/crypto/blake2b"
)

const HashSize = blake2b.Size256

type HashKey [HashSize]byte

func ToHashKey(x []byte) (k HashKey) {
	copy(k[:], x)
	return
}

func newHashGen() hash.Hash {
	gen, err := blake2b.New256(nil)
	check.E(err)
	return gen
}

type hasher struct {
	gen hash.Hash
	buf []byte
}

var hasherPool = sync.Pool{New: func() interface{} {
	return &hasher{newHashGen(), make([]byte, 0x10000)}
}}

func getHasher() *hasher {
	return hasherPool.Get().(*hasher)
}

func (h *hasher) done() {
	hasherPool.Put(h)
}

func (h *hasher) run(fileName string, expectedSize int64) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	h.gen.Reset()
	size, err := io.CopyBuffer(h.gen, file, h.buf)
	if err != nil {
		return nil, err
	}
	if expectedSize != size {
		return nil, errors.New("file size changed")
	}
	return h.gen.Sum(nil), nil
}
