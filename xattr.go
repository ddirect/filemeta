package filemeta

import (
	"encoding/binary"
	"errors"
	"hash/crc32"

	"github.com/ddirect/check"
	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/proto"
)

var crcTable = crc32.MakeTable(crc32.Castagnoli)
var checkValue = crc32.Checksum([]byte{0, 0, 0, 0}, crcTable)

func removeCrc(b []byte) []byte {
	if crc32.Checksum(b, crcTable) != checkValue {
		panic(errors.New("crc check failed"))
	}
	return b[:len(b)-4]
}

func appendCrc(b []byte) []byte {
	crc := make([]byte, 4)
	binary.LittleEndian.PutUint32(crc, crc32.Checksum(b, crcTable))
	return append(b, crc...)
}

func readXattr(fileName string, attrName string, data proto.Message, buf []byte) []byte {
	if cap(buf) > 0 {
		buf = buf[:cap(buf)]
	} else {
		buf = make([]byte, 64)
	}
again:
	siz, err := unix.Getxattr(fileName, attrName, buf)
	if err == unix.ERANGE {
		buf = make([]byte, len(buf)*3/2)
		goto again
	}
	check.Efile("getxattr", fileName, err)
	buf = removeCrc(buf[:siz])
	check.E(proto.Unmarshal(buf, data))
	return buf
}

func writeXattr(fileName string, attrName string, data proto.Message, buf []byte) []byte {
	var err error
	buf, err = proto.MarshalOptions{}.MarshalAppend(buf[:0], data)
	check.E(err)
	buf = appendCrc(buf)
	check.Efile("setxattr", fileName, unix.Setxattr(fileName, attrName, buf, 0))
	return buf
}
