package filemeta

import (
	"errors"
	"hash"
	"io"
	"os"

	"github.com/ddirect/check"
	"golang.org/x/crypto/blake2b"
)

const HashSize = blake2b.Size256

type HashKey [HashSize]byte

func ToHashKey(x []byte) (k HashKey) {
	copy(k[:], x)
	return
}

func newHasher() hash.Hash {
	gen, err := blake2b.New256(nil)
	check.E(err)
	return gen
}

func newHasherBuffer() []byte {
	return make([]byte, 0x10000)
}

func getFileHash(fileName string, expectedSize int64) ([]byte, error) {
	return hashCore(fileName, expectedSize, newHasher(), newHasherBuffer())
}

func hashCore(fileName string, expectedSize int64, gen hash.Hash, buf []byte) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	size, err := io.CopyBuffer(gen, file, buf)
	if err != nil {
		return nil, err
	}
	if expectedSize != size {
		return nil, errors.New("file size changed")
	}
	return gen.Sum(nil), nil
}
