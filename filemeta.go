package filemeta

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ddirect/check"
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

func OpString(op Op) string {
	if uint(op) < uint(len(opStrings)) {
		return opStrings[uint(op)]
	}
	return "<unknown>"
}

func core(m Op, fileName string) (data Data) {
	data.Operation = m
	data.Path = fileName
	st, err := os.Stat(fileName)
	if err != nil {
		data.Error = err
		return
	}
	if !st.Mode().IsRegular() {
		data.Error = fmt.Errorf("'%s' is not regular", fileName)
		return
	}
	data.Info = st

	fileSize := st.Size()
	fileTimeNs := st.ModTime().UnixNano()
	attr, err := readAttributes(fileName)
	if err != nil && m == OpInspect {
		data.Error = err
		return
	}
	if err != nil || attr.Size != fileSize || attr.TimeNs != fileTimeNs {
		data.Changed = err == nil
		if m != OpRefresh {
			return
		}
		attr.Size = fileSize
		attr.TimeNs = fileTimeNs
		data.Attr = attr
		data.hashNeeded = true
		return
	}

	data.Attr = attr
	data.hashNeeded = m == OpVerify
	return
}

func Operation(m Op, fileName string) (data Data) {
	defer check.Recover(&data.Error)
	data = core(m, fileName)
	if data.hashNeeded {
		data.notifyHash(getFileHash(data.Path, data.Attr.Size))
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
