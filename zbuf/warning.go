package zbuf

import (
	"fmt"

	"github.com/brimsec/zq/zng"
)

type WarningReader struct {
	Reader
	ch chan string
}

// WarningReader returns a Reader that reads from zr.  Any error encountered is
// sent to ch, and then a nil *zng.Record and nil error are returned.
func NewWarningReader(zr Reader, ch chan string) *WarningReader {
	return &WarningReader{Reader: zr, ch: ch}
}

func (w *WarningReader) Read() (*zng.Record, error) {
	rec, err := w.Reader.Read()
	if err != nil {
		w.ch <- fmt.Sprintf("%s: %s", w.Reader, err)
		return nil, nil
	}
	return rec, nil
}
