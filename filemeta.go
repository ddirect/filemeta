package filemeta

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ddirect/check"
	"github.com/ddirect/sys"
	"google.golang.org/protobuf/proto"
)

type Op int8

type FetchFunc func(fileName string) Data

const (
	OpGet Op = iota
	OpVerify
	OpRefresh
	OpInspect
)

var opStrings = []string{
	"get",
	"verify",
	"refresh",
	"inspect",
}

func (op Op) String() string {
	if uint(op) < uint(len(opStrings)) {
		return opStrings[uint(op)]
	}
	return "<unknown>"
}

func core(m Op, fileName string) (d Data) {
	d.Operation = m
	if d.FileInfo, d.Error = sys.Stat(fileName); d.Error != nil {
		return
	}
	if !d.Mode.IsRegular() {
		d.Error = fmt.Errorf("'%s' is not regular", fileName)
		return
	}

	attr, err := readAttr(fileName)
	if err != nil && m == OpInspect {
		d.Error = err
		return
	}
	if err != nil || attr.Size != d.Size || attr.TimeNs != d.ModTimeNs {
		d.Changed = err == nil
		if m != OpRefresh {
			return
		}
		d.hashNeeded = true
		return
	}
	d.Hash = attr.Hash
	d.hashNeeded = m == OpVerify
	return
}

func Operation(op Op, fileName string) (data Data) {
	data = core(op, fileName)
	if data.hashNeeded {
		h := getHasher()
		defer h.done()
		data.notifyHash(h.run(fileName, data.Size))
	}
	return
}

// Gets the metadata if available; returns an error if not
func Inspect(fileName string) Data {
	return Operation(OpInspect, fileName)
}

// Gets the metadata if available; if not available data.Attr is nil
func Get(fileName string) Data {
	return Operation(OpGet, fileName)
}

// Like get, but additional it verifies the hash (scrub)
func Verify(fileName string) Data {
	return Operation(OpVerify, fileName)
}

// Gets the metadata, refreshing it if necessary
func Refresh(fileName string) Data {
	return Operation(OpRefresh, fileName)
}

func customCore(fileName string, attrName string, data proto.Message, core func(string, string, proto.Message, []byte) []byte) (err error) {
	if attrName == "" {
		return errors.New("attribute name cannot be empty")
	}
	defer check.Recover(&err)
	asd := attrPool.Get().(*attrSerDes)
	defer attrPool.Put(asd)
	asd.buf = core(fileName, fmt.Sprintf("%s.%s", fileMetaAttr, strings.ToUpper(attrName)), data, asd.buf)
	return
}

func ReadCustom(fileName string, attrName string, data proto.Message) error {
	return customCore(fileName, attrName, data, readXattr)
}

func WriteCustom(fileName string, attrName string, data proto.Message) error {
	return customCore(fileName, attrName, data, writeXattr)
}
