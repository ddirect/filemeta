package filemeta

import (
	"bytes"
	"os"
	"time"

	"github.com/ddirect/check"
)

type Data struct {
	Path         string
	Info         os.FileInfo
	Attr         *Attributes
	Hashed       bool // the file has just been hashed
	Changed      bool // the file had attributes, but they are no longer valid
	VerifyFailed bool
}

func (d *Data) verify() {
	d.VerifyFailed = bytes.Compare(d.Attr.Hash, getFileHash(d.Path, d.Info.Size())) != 0
}

func (d *Data) Rename(newPath string) (err error) {
	defer check.Recover(&err)
	check.E(os.Rename(d.Path, newPath))
	d.Path = newPath
	return
}

func (d *Data) SetTime(tim time.Time) (err error) {
	defer check.Recover(&err)
	check.E(os.Chtimes(d.Path, time.Now(), tim))
	if d.Attr != nil {
		d.Attr.TimeNs = tim.UnixNano()
		d.Attr.write(d.Path)
	}
	return
}
