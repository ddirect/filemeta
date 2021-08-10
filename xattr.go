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

func readXattr(fileName string, attrName string, data proto.Message) {
	buf := make([]byte, 256)
	siz, err := unix.Getxattr(fileName, attrName, buf)
	check.Efile("getxattr", fileName, err)
	check.E(proto.Unmarshal(removeCrc(buf[:siz]), data))
	return
}

func writeXattr(fileName string, attrName string, data proto.Message) {
	buf, err := proto.Marshal(data)
	check.E(err)
	check.Efile("setxattr", fileName, unix.Setxattr(fileName, attrName, appendCrc(buf), 0))
}
