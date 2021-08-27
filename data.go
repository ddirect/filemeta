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
	Error        error
	Operation    Op
	Hashed       bool // the file has just been hashed
	Changed      bool // the file had attributes, but they are no longer valid
	VerifyFailed bool
	hashNeeded   bool
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

func (d *Data) notifyHash(hash []byte, err error) {
	defer check.Recover(&d.Error)
	if err == nil {
		d.Hashed = true
		switch d.Operation {
		case OpRefresh:
			d.Attr.Hash = hash
			d.writeAttr()
		case OpVerify:
			d.VerifyFailed = bytes.Compare(d.Attr.Hash, hash) != 0
		}
	} else {
		d.Error = err
		switch d.Operation {
		case OpVerify:
			d.VerifyFailed = true
		}
	}
}

func (d *Data) writeAttr() {
	mode := d.Info.Mode().Perm()
	neededMode := mode | 0200
	if mode != neededMode {
		check.E(os.Chmod(d.Path, neededMode))
		defer check.DeferredE(func() error { return os.Chmod(d.Path, mode) })
	}
	d.Attr.write(d.Path)
}
