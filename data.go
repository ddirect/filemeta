package filemeta

import (
	"bytes"
	"os"
	"time"

	"github.com/ddirect/check"
	"github.com/ddirect/sys"
)

type Data struct {
	Path         string
	Info         sys.FileInfo
	Hash         []byte
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

func (d *Data) SetModTime(tim time.Time) (err error) {
	defer check.Recover(&err)
	check.E(os.Chtimes(d.Path, time.Now(), tim))
	d.Info.ModTimeNs = tim.UnixNano()
	d.writeAttributes()
	return
}

func (d *Data) GetModTime() time.Time {
	return time.Unix(0, d.Info.ModTimeNs)
}

func (d *Data) GetAttr() Attr {
	return Attr{d.Info.Size, d.Info.ModTimeNs, d.Hash}
}

func (d *Data) writeAttributes() {
	mode := d.Info.Mode.Perm()
	neededMode := mode | 0200
	if mode != neededMode {
		check.E(os.Chmod(d.Path, neededMode))
		defer check.DeferredE(func() error { return os.Chmod(d.Path, mode) })
	}
	attr := d.GetAttr()
	attr.write(d.Path)
}

func (d *Data) notifyHash(hash []byte, err error) {
	defer check.Recover(&d.Error)
	if err == nil {
		d.Hashed = true
		switch d.Operation {
		case OpRefresh:
			d.Hash = hash
			d.writeAttributes()
		case OpVerify:
			d.VerifyFailed = bytes.Compare(d.Hash, hash) != 0
		}
	} else {
		d.Error = err
		switch d.Operation {
		case OpVerify:
			d.VerifyFailed = true
		}
	}
}
