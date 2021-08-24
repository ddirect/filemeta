module github.com/ddirect/filemeta

go 1.16

replace github.com/ddirect/check => ../check

replace github.com/ddirect/filetest => ../filetest

replace github.com/ddirect/xrand => ../xrand

replace github.com/ddirect/format => ../format

require (
	github.com/ddirect/check v0.0.0-00010101000000-000000000000
	github.com/ddirect/filetest v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/sys v0.0.0-20210820121016-41cdb8703e55
	google.golang.org/protobuf v1.27.1
)
