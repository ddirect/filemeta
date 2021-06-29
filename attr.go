package filemeta

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"time"

	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/proto"
)

const attrName = "user.FILEMETA"

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

func readAttributes(fileName string) (attr *Attributes, errOut error) {
	defer handlePanic(&errOut)
	// ensure an attribute is always created (it's used also in case of error)
	attr = new(Attributes)
	buf := make([]byte, 256)
	siz, err := unix.Getxattr(fileName, attrName, buf)
	check_det("getxattr", fileName, err)
	check(proto.Unmarshal(removeCrc(buf[:siz]), attr))
	return
}

func (attr *Attributes) write(fileName string) {
	buf, err := proto.Marshal(attr)
	check(err)
	check_det("setxattr", fileName, unix.Setxattr(fileName, attrName, appendCrc(buf), 0))
}

func (attr *Attributes) Time() time.Time {
	return time.Unix(0, attr.TimeNs)
}
