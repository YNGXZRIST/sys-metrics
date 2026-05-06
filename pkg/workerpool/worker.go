package workerpool

import (
	"context"
)

// Worker reads tasks from tCh, runs them, and optionally sends results to rCh.
type Worker struct {
	ctx context.Context
	tCh chan Task
	rCh chan Task
}

// NewWorker creates a worker bound to the given task and result channels.
func NewWorker(ctx context.Context, tCh, rCh chan Task) *Worker {
	return &Worker{ctx: ctx, tCh: tCh, rCh: rCh}
}

// StartBg runs the processing loop until the context is canceled or tCh is closed.
func (w *Worker) StartBg() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case t, ok := <-w.tCh:
			if !ok {
				return
			}
			t.process()
			if t.NeedResult {
				select {
				case <-w.ctx.Done():
					return
				case w.rCh <- t:
				}
			}
		}
	}
}
