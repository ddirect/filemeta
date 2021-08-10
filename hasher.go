package filemeta

import (
	"context"
	"errors"
	"hash"
	"io"
	"os"

	"github.com/ddirect/check"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/sync/semaphore"
)

const HashSize = blake2b.Size256

var hashsem = semaphore.NewWeighted(4)

func newHasher() hash.Hash {
	gen, err := blake2b.New256(nil)
	check.E(err)
	return gen
}

func getFileHash(fileName string, expectedSize int64) []byte {
	hashsem.Acquire(context.Background(), 1)
	defer hashsem.Release(1)

	gen := newHasher()
	file, err := os.Open(fileName)
	check.E(err)
	defer file.Close()
	if expectedSize != check.I64E(io.Copy(gen, file)) {
		panic(errors.New("file size changed"))
	}
	return gen.Sum(nil)
}
