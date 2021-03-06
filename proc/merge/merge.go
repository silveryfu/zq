package merge

import (
	"sync"

	"github.com/brimsec/zq/proc"
	"github.com/brimsec/zq/zbuf"
)

// A Merge proc merges multiple upstream inputs into one output.
type Proc struct {
	pctx     *proc.Context
	once     sync.Once
	ch       <-chan proc.Result
	doneCh   chan struct{}
	parents  []*runnerProc
	nparents int
}

type runnerProc struct {
	pctx   *proc.Context
	parent proc.Interface
	ch     chan<- proc.Result
	doneCh <-chan struct{}
}

func (r *runnerProc) run() {
	for {
		batch, err := r.parent.Pull()
		select {
		case r.ch <- proc.Result{batch, err}:
			if proc.EOS(batch, err) {
				return
			}
		case <-r.doneCh:
			r.parent.Done()
			return
		case <-r.pctx.Done():
			return
		}
	}
}

func New(pctx *proc.Context, parents []proc.Interface) *Proc {
	ch := make(chan proc.Result)
	doneCh := make(chan struct{})
	var runners []*runnerProc
	for _, parent := range parents {
		runners = append(runners, &runnerProc{
			pctx:   pctx,
			parent: parent,
			ch:     ch,
			doneCh: doneCh,
		})
	}
	return &Proc{
		pctx:     pctx,
		ch:       ch,
		doneCh:   doneCh,
		parents:  runners,
		nparents: len(parents),
	}
}

// Pull implements the merge logic for returning data from the upstreams.
func (p *Proc) Pull() (zbuf.Batch, error) {
	p.once.Do(func() {
		for _, m := range p.parents {
			go m.run()
		}
	})
	for {
		if p.nparents == 0 {
			return nil, nil
		}
		select {
		case res := <-p.ch:
			if res.Batch != nil || res.Err != nil {
				return res.Batch, res.Err
			}
			p.nparents--
		case <-p.pctx.Done():
			return nil, p.pctx.Err()
		}
	}
}

func (m *Proc) Done() {
	close(m.doneCh)
}
