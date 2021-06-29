package filemeta

import (
	"context"
	"errors"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/sync/semaphore"
)

const HashSize = blake2b.Size256

var hashsem = semaphore.NewWeighted(4)

func getFileHash(fileName string, expectedSize int64) []byte {
	hashsem.Acquire(context.Background(), 1)
	defer hashsem.Release(1)

	gen, err := blake2b.New256(nil)
	check(err)
	file, err := os.Open(fileName)
	check(err)
	defer file.Close()
	if expectedSize != checkI64(io.Copy(gen, file)) {
		panic(errors.New("file size changed"))
	}
	return gen.Sum(nil)
}
