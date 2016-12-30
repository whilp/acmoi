package acmoi

import (
	"bytes"
	"fmt"
)

// Archive stores several files and provides access to them in a format compatible with the go guru. https://godoc.org/golang.org/x/tools/cmd/guru
type Archive struct {
	buf *bytes.Buffer
}

// NewArchive creates a new archive.
func NewArchive() *Archive {
	buf := new(bytes.Buffer)
	return &Archive{buf: buf}
}

func (a *Archive) Write(name string, body []byte) error {
	header := fmt.Sprintf("%s\n%d\n", name, len(body))
	if _, err := a.buf.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := a.buf.Write(body); err != nil {
		return err
	}
	return nil
}

// Buffer returns a reader with the formatted contents of the archive.
func (a *Archive) Buffer() *bytes.Reader {
	return bytes.NewReader(a.buf.Bytes())
}
