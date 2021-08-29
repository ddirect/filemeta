package filemeta

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ddirect/check"

	ft "github.com/ddirect/filetest"
)

type filemap map[string][]byte

func checkCore(t *testing.T, op Op, base string, refFm filemap, stats ft.DirStats) (Op, func(cb func(string)), func(Data), func()) {
	fm := make(filemap)
	inodes := make(map[uint64]struct{})
	hashes := make(map[HashKey]struct{})
	return op,
		func(cb func(string)) {
			filepath.Walk(base, func(path string, info fs.FileInfo, err error) error {
				check.E(err)
				if !info.IsDir() {
					cb(path)
				}
				return nil
			})
		},
		func(data Data) {
			check.E(data.Error)
			fm[data.Path] = data.Hash
			hashes[ToHashKey(data.Hash)] = struct{}{}
			inodes[data.Info.Inode] = struct{}{}
		},
		func() {
			if !reflect.DeepEqual(fm, refFm) {
				t.Fatal(op, "failed")
			}
			if len(hashes) != stats.UniqueHashes {
				t.Fatalf("hash count mismatch: %d != %d", len(hashes), stats.UniqueHashes)
			}
			fileCount := stats.UniqueHashes + stats.ClonedFiles
			if len(inodes) != fileCount {
				t.Fatalf("unique file count mismatch: %d != %d", len(inodes), fileCount)
			}
		}
}

func testAsyncOp(op Op, walk func(func(string)), core func(Data), epilogue func()) {
	async := AsyncOperations(op, 0, 0)
	go func() {
		defer close(async.FileIn)
		walk(func(path string) {
			async.FileIn <- path
		})
	}()
	for data := range async.DataOut {
		core(data)
	}
	epilogue()
}

func testSyncOp(op Op, walk func(func(string)), core func(Data), epilogue func()) {
	walk(func(path string) {
		core(Operation(op, path))
	})
	epilogue()
}

func makeTree(t *testing.T) (string, filemap, ft.DirStats) {
	base, tree, stats := ft.CommitNewDefaultRandomTree(t)
	fm := make(filemap)
	tree.EachFileRecursive(func(f *ft.File) {
		fm[f.PathFrom(base)] = f.Hash
	})
	return base, fm, stats
}

func TestAsync(t *testing.T) {
	base, fm, stats := makeTree(t)
	testAsyncOp(checkCore(t, OpRefresh, base, fm, stats))
	testAsyncOp(checkCore(t, OpGet, base, fm, stats))
}

func TestSync(t *testing.T) {
	base, fm, stats := makeTree(t)
	testSyncOp(checkCore(t, OpRefresh, base, fm, stats))
	testSyncOp(checkCore(t, OpGet, base, fm, stats))
}
