package filemeta

import (
	"errors"
	"hash"
	"io"
	"os"

	"github.com/ddirect/check"
	"golang.org/x/crypto/blake2b"
)

func newHasher() hash.Hash {
	gen, err := blake2b.New256(nil)
	check.E(err)
	return gen
}

func getFileHash(fileName string, expectedSize int64) []byte {
	gen := newHasher()
	file, err := os.Open(fileName)
	check.E(err)
	defer file.Close()
	if expectedSize != check.I64E(io.Copy(gen, file)) {
		panic(errors.New("file size changed"))
	}
	return gen.Sum(nil)
}
