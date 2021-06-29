package filemeta

import (
	"bytes"
	"os"
	"time"
)

type Data struct {
	Path    string
	Info    os.FileInfo
	Attr    *Attributes
	Hashed  bool // the file has just been hashed
	Changed bool // the file had attributes, but they are no longer valid
}

func (d *Data) Verify() (res bool, err error) {
	defer handlePanic(&err)
	res = bytes.Compare(d.Attr.Hash, getFileHash(d.Path, d.Info.Size())) == 0
	return
}

func (d *Data) Rename(newPath string) (err error) {
	defer handlePanic(&err)
	check(os.Rename(d.Path, newPath))
	d.Path = newPath
	return
}

func (d *Data) SetTime(tim time.Time) (err error) {
	defer handlePanic(&err)
	check(os.Chtimes(d.Path, time.Now(), tim))
	if d.Attr != nil {
		d.Attr.TimeNs = tim.UnixNano()
		d.Attr.write(d.Path)
	}
	return
}
