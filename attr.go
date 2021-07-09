package filemeta

import (
	"time"
)

const fileMetaAttr = "user.FILEMETA"

func readAttributes(fileName string) (attr *Attributes, errOut error) {
	defer handlePanic(&errOut)
	// ensure an attribute is always created (it's used also in case of error)
	attr = new(Attributes)
	readXattr(fileName, fileMetaAttr, attr)
	return
}

func (attr *Attributes) write(fileName string) {
	writeXattr(fileName, fileMetaAttr, attr)
}

func (attr *Attributes) Time() time.Time {
	return time.Unix(0, attr.TimeNs)
}
