package filemeta

import (
	"os"
	"time"

	"github.com/ddirect/check"
)

type FileWriter struct {
	Open  func(fileName string, fileFlags int, filePerm os.FileMode) error
	Write func([]byte) error
	Close func(fileTimeNs int64) (*Attributes, error)
}

func NewFileWriter() FileWriter {
	gen := newHasher()
	var size int64
	var file *os.File
	return FileWriter{
		func(fileName string, fileFlags int, filePerm os.FileMode) (err error) {
			gen.Reset()
			size = 0
			file, err = os.OpenFile(fileName, fileFlags, filePerm)
			return
		},
		func(data []byte) error {
			gen.Write(data)
			_, err := file.Write(data)
			size += int64(len(data))
			return err
		},
		func(fileTimeNs int64) (attr *Attributes, err error) {
			defer check.Recover(&err)
			check.E(file.Close())
			attr = new(Attributes)
			attr.Hash = gen.Sum(nil)
			attr.TimeNs = fileTimeNs
			attr.Size = size
			attr.write(file.Name())
			check.E(os.Chtimes(file.Name(), time.Now(), time.Unix(0, fileTimeNs)))
			return
		},
	}
}
