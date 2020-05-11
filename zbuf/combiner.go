package zbuf

import (
	"fmt"
	"io"

	"github.com/brimsec/zq/expr"
	"github.com/brimsec/zq/zng"
)

var (
	SortTsAscending  = sortTsFn(false)
	SortTsDescending = sortTsFn(true)
)

type Combiner struct {
	readers []Reader
	hol     []*zng.Record
	done    []bool
	sortFn  expr.SortFn
}

func NewCombiner(readers []Reader, sortFn expr.SortFn) *Combiner {
	return &Combiner{
		readers: readers,
		hol:     make([]*zng.Record, len(readers)),
		done:    make([]bool, len(readers)),
		sortFn:  sortFn,
	}
}

func (c *Combiner) Read() (*zng.Record, error) {
	idx := -1
	for k, l := range c.readers {
		if c.done[k] {
			continue
		}
		if c.hol[k] == nil {
			tup, err := l.Read()
			if err != nil {
				return nil, fmt.Errorf("%s: %w", c.readers[k], err)
			}
			if tup == nil {
				c.done[k] = true
				continue
			}
			c.hol[k] = tup
		}
		if idx == -1 || c.sortFn(c.hol[k], c.hol[idx]) == -1 {
			idx = k
		}
	}
	if idx == -1 {
		return nil, nil
	}
	tup := c.hol[idx]
	c.hol[idx] = nil
	return tup, nil
}

func (c *Combiner) closeReader(r Reader) error {
	if closer, ok := r.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Close closes underlying Readers implementing the io.Closer
// interface if they haven't already been closed.
func (c *Combiner) Close() error {
	var err error
	for k, r := range c.readers {
		c.done[k] = true
		// Return only the first error, but closing everything else if there is
		// an error.
		if e := c.closeReader(r); err == nil {
			err = e
		}
	}
	return err
}

func sortTsFn(desc bool) expr.SortFn {
	return func(a, b *zng.Record) int {
		var i int
		if a.Ts == b.Ts {
			i = 0
		} else if a.Ts < b.Ts {
			i = -1
		} else {
			i = 1
		}
		if desc {
			i = i * -1
		}
		return i
	}
}
