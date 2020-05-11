package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/brimsec/zq/pkg/nano"
	"github.com/brimsec/zq/zbuf"
	"github.com/brimsec/zq/zio"
	"github.com/brimsec/zq/zio/zngio"
	"github.com/brimsec/zq/zng"
	"github.com/brimsec/zq/zng/resolver"
	"github.com/brimsec/zq/zqe"
)

const (
	allZngFile    = "all.zng"
	allZngTmpFile = "all.zng.tmp"
)

type ZngStorage struct {
	path       string
	writing    uint32
	span       nano.Span
	index      *zngio.TimeIndex
	streamsize int
}

func NewZng(path string) *ZngStorage {
	return &ZngStorage{path: path, index: zngio.NewTimeIndex()}
}

func (s *ZngStorage) join(args ...string) string {
	args = append([]string{s.path}, args...)
	return filepath.Join(args...)
}

func (s *ZngStorage) Open(span nano.Span) (zbuf.ReadCloser, error) {
	zctx := resolver.NewContext()
	f, err := os.Open(s.join(allZngFile))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		// Couldn't read all.zng, check for an old space with all.bzng
		bzngFile := strings.TrimSuffix(allZngFile, filepath.Ext(allZngFile)) + ".bzng"
		f, err = os.Open(s.join(bzngFile))
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			r := zngio.NewReader(strings.NewReader(""), zctx)
			return zbuf.NopReadCloser(r), nil
		}
	}
	// XXX plz keep track of how long this takes
	return s.index.NewReader(f, zctx, span)
}

type spanWriter struct {
	span nano.Span
}

func (w *spanWriter) Write(rec *zng.Record) error {
	if rec.Ts == 0 {
		return nil
	}
	first := w.span == nano.Span{}
	s := nano.Span{Ts: rec.Ts, Dur: 1}
	if first {
		w.span = s
	} else {
		w.span = w.span.Union(s)
	}
	return nil
}

// XXX only one writeto allowed at a time
func (s *ZngStorage) WriteTo(ctx context.Context, zr zbuf.Reader) error {
	if !atomic.CompareAndSwapUint32(&s.writing, 0, 1) {
		return zqe.E(zqe.Conflict, "simultaneous writing already occurring")
	}
	defer func() { atomic.StoreUint32(&s.writing, 0) }()

	// For the time being, this endpoint will overwrite any underlying data.
	// In order to get rid errors on any concurrent searches on this space,
	// write zng to a temp file and rename on successful conversion.
	tmppath := s.join(allZngTmpFile)
	zngfile, err := os.Create(tmppath)
	if err != nil {
		return err
	}
	sr := zngio.NewTsSort(ctx, zr, 0, true, "")
	fileWriter := zngio.NewWriter(zngfile, zio.WriterFlags{StreamRecordsMax: s.streamsize})
	spanWriter := &spanWriter{}
	zw := zbuf.MultiWriter(fileWriter, spanWriter)

	if err := zbuf.CopyWithContext(ctx, zw, sr); err != nil {
		zngfile.Close()
		os.RemoveAll(tmppath)
		return err
	}
	if err := zngfile.Close(); err != nil {
		os.RemoveAll(tmppath)
		return err
	}
	if err := os.Rename(tmppath, s.join(allZngFile)); err != nil {
		return err
	}
	return s.SetSpan(spanWriter.span)
}

func (s *ZngStorage) Close() error {
	return nil
}

// XXX
func (s *ZngStorage) SetSpan(span nano.Span) error {
	first := s.span == nano.Span{}
	if first {
		s.span = span
	} else {
		s.span = s.span.Union(span)
	}
	return nil

}

func (s *ZngStorage) Span() nano.Span {
	return s.span
}

func (s *ZngStorage) Size() (int64, error) {
	zngpath := s.join(allZngFile)
	f, err := os.Stat(zngpath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	return f.Size(), nil
}
