package filemeta

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	"github.com/ddirect/check"

	ft "github.com/ddirect/filetest"
)

type filemap map[string][]byte

func testOp(t *testing.T, op Op, base string, refFm filemap, stats ft.DirStats) {
	async := AsyncOperations(op, 0, 0)
	go func() {
		defer close(async.FileIn)
		filepath.Walk(base, func(path string, info fs.FileInfo, err error) error {
			check.E(err)
			if !info.IsDir() {
				async.FileIn <- path
			}
			return nil
		})
	}()
	fm := make(filemap)
	inodes := make(map[uint64]struct{})
	hashes := make(map[HashKey]struct{})
	for data := range async.DataOut {
		check.E(data.Error)
		fm[data.Path] = data.Attr.Hash
		hashes[ToHashKey(data.Attr.Hash)] = struct{}{}
		si := data.Info.Sys().(*syscall.Stat_t)
		inodes[si.Ino] = struct{}{}
	}
	if !reflect.DeepEqual(fm, refFm) {
		t.Fatal(OpString(op), "failed")
	}
	if len(hashes) != stats.UniqueHashes {
		t.Fatalf("hash count mismatch: %d != %d", len(hashes), stats.UniqueHashes)
	}
	fileCount := stats.UniqueHashes + stats.ClonedFiles
	if len(inodes) != fileCount {
		t.Fatalf("unique file count mismatch: %d != %d", len(inodes), fileCount)
	}
}

func TestAsync(t *testing.T) {
	base, tree, stats := ft.CommitNewDefaultRandomTree(t)
	fm := make(filemap)
	tree.EachFileRecursive(func(f *ft.File) {
		fm[f.PathFrom(base)] = f.Hash
	})
	testOp(t, OpRefresh, base, fm, stats)
	testOp(t, OpGet, base, fm, stats)
}
