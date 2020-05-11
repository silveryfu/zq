package ingest

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/brimsec/zq/zbuf"
	"github.com/brimsec/zq/zio/detector"
	"github.com/brimsec/zq/zio/ndjsonio"
	"github.com/brimsec/zq/zng/resolver"
	"github.com/brimsec/zq/zqd/api"
	"github.com/brimsec/zq/zqd/storage"
	"github.com/brimsec/zq/zqe"
)

// x509.14:00:00-15:00:00.log.gz (open source zeek)
// x509_20191101_14:00:00-15:00:00+0000.log.gz (corelight)
const DefaultJSONPathRegexp = `([a-zA-Z0-9_]+)(?:\.|_\d{8}_)\d\d:\d\d:\d\d\-\d\d:\d\d:\d\d(?:[+\-]\d{4})?\.log(?:$|\.gz)`

type LogTransaction struct {
	bytesRead  int64
	bytesTotal int64
	warnings   []string
	readers    []zbuf.Reader
	files      []*os.File
	err        error

	warningCh chan string
	doneCh    chan struct{}
}

// Logs ingests the provided list of files into the provided space.
// Like ingest.Pcap, this overwrites any existing data in the space.
func NewLogTransaction(ctx context.Context, store *storage.ZngStorage, req api.LogPostRequest) (*LogTransaction, error) {
	p := &LogTransaction{
		warningCh: make(chan string, 5),
		doneCh:    make(chan struct{}),
	}
	var cfg detector.OpenConfig
	if req.JSONTypeConfig != nil {
		cfg.JSONTypeConfig = req.JSONTypeConfig
		cfg.JSONPathRegex = DefaultJSONPathRegexp
	}
	zctx := resolver.NewContext()
	for _, path := range req.Paths {
		if err := p.openFile(zctx, path, cfg); err != nil {
			return nil, err
		}
	}
	go p.start(ctx, store)
	return p, nil
}

func (p *LogTransaction) openFile(zctx *resolver.Context, path string, cfg detector.OpenConfig) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return zqe.E(zqe.Invalid, "path is a directory")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	// Initialize zbuf.Reader and wrap in NamedReader (to make file easier to
	// identify for errors) and WarningReader (so Read warnings propogate to
	// warning channel).
	zr, err := detector.NewReader(io.TeeReader(f, p), zctx)
	if zjr, ok := zr.(*ndjsonio.Reader); ok {
		if err := detector.ConfigureNDJSONReader(zjr, cfg, path); err != nil {
			return err
		}
	}
	zr = zbuf.NamedReader(zr, path)
	zr = zbuf.NewWarningReader(zr, p.warningCh)
	// If an error occur here just log it as a warning and do not add to the
	// list of readers.
	if err != nil {
		p.warnings = append(p.warnings, fmt.Sprintf("%s: %s", path, err))
		return f.Close()
	}
	p.bytesTotal += info.Size()
	p.readers = append(p.readers, zr)
	p.files = append(p.files, f)
	return nil
}

func (p *LogTransaction) start(ctx context.Context, store *storage.ZngStorage) {
	// first drain warnings
	for _, warning := range p.warnings {
		p.warningCh <- warning
	}
	r := zbuf.MultiReader(p.readers...)
	p.err = store.WriteTo(ctx, r)
	close(p.doneCh)
}

func (p *LogTransaction) Status() api.LogPostStatus {
	return api.LogPostStatus{
		Type:         "LogPostStatus",
		LogTotalSize: p.bytesTotal,
		LogReadSize:  atomic.LoadInt64(&p.bytesRead),
	}
}

// Write is used to Tee all input from ingest files and keep track of the total
// bytes read.
func (p *LogTransaction) Write(b []byte) (n int, err error) {
	n = len(b)
	atomic.AddInt64(&p.bytesRead, int64(n))
	return
}

func (p *LogTransaction) Warning() <-chan string {
	return p.warningCh
}

func (p *LogTransaction) Done() <-chan struct{} {
	return p.doneCh
}

func (p *LogTransaction) Error() error {
	<-p.doneCh
	return p.err
}
