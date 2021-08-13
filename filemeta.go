package filemeta

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ddirect/check"
	"google.golang.org/protobuf/proto"
)

type Op int

type FetchFunc func(fileName string) (Data, error)

const (
	OpGet Op = iota
	OpVerify
	OpRefresh
	OpInspect
)

func Operation(m Op, fileName string) (data Data, errOut error) {
	defer check.Recover(&errOut)
	st, err := os.Stat(fileName)
	check.E(err)
	if !st.Mode().IsRegular() {
		panic(fmt.Errorf("'%s' is not regular", fileName))
	}
	data.Path = fileName
	data.Info = st

	fileSize := st.Size()
	fileTimeNs := st.ModTime().UnixNano()
	attr, err := readAttributes(fileName)
	if err != nil && m == OpInspect {
		errOut = err
		return
	}
	if err != nil || attr.Size != fileSize || attr.TimeNs != fileTimeNs {
		data.Changed = err == nil
		if m != OpRefresh {
			return
		}
		attr.Hash = getFileHash(fileName, fileSize)
		attr.Size = fileSize
		attr.TimeNs = fileTimeNs
		attr.write(fileName)
		data.Hashed = true
	}

	data.Attr = attr
	if m == OpVerify {
		data.verify()
	}
	return
}

// Gets the metadata if available; returns an error if not
func Inspect(fileName string) (data Data, errOut error) {
	return Operation(OpInspect, fileName)
}

// Gets the metadata if available; if not available data.Attr is nil
func Get(fileName string) (data Data, errOut error) {
	return Operation(OpGet, fileName)
}

// Like get, but additional it verifies the hash (scrub)
func Verify(fileName string) (data Data, errOut error) {
	return Operation(OpVerify, fileName)
}

// Gets the metadata, refreshing it if necessary
func Refresh(fileName string) (data Data, errOut error) {
	return Operation(OpRefresh, fileName)
}

func customCore(fileName string, attrName string, data proto.Message, core func(string, string, proto.Message)) (err error) {
	if attrName == "" {
		return errors.New("attribute name cannot be empty")
	}
	defer check.Recover(&err)
	core(fileName, fmt.Sprintf("%s.%s", fileMetaAttr, strings.ToUpper(attrName)), data)
	return
}

func ReadCustom(fileName string, attrName string, data proto.Message) error {
	return customCore(fileName, attrName, data, readXattr)
}

func WriteCustom(fileName string, attrName string, data proto.Message) error {
	return customCore(fileName, attrName, data, writeXattr)
}
