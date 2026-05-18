// Package workerpool implements a worker pool and a task queue with result delivery.
package workerpool

import (
	"context"
	"errors"
	"sync"
)

// Pool runs a fixed number of workers that pull tasks from tCh and send results to rCh.
type Pool struct {
	rCh       chan Task
	tCh       chan Task
	tRes      *[]Task
	workers   []*Worker
	isStopped bool
	mu        sync.Mutex
	wg        sync.WaitGroup
}

// NewPool creates a pool with buffered channels of size c (worker count is fixed in StartBg).
func NewPool(c int) *Pool {
	w := &Pool{
		workers:   make([]*Worker, 0, c),
		rCh:       make(chan Task, c),
		tCh:       make(chan Task, c),
		tRes:      &[]Task{},
		wg:        sync.WaitGroup{},
		isStopped: false,
	}
	return w
}

// StartBg starts workers in the background; the count equals cap(workers), which is set by NewPool(c).
func (p *Pool) StartBg(ctx context.Context) {
	for i := 0; i < cap(p.workers); i++ {
		worker := NewWorker(ctx, p.tCh, p.rCh)
		p.workers = append(p.workers, worker)
		p.wg.Add(1)
		go func() {
			worker.StartBg()
			p.wg.Done()
		}()
	}

}

// Add enqueues a task for execution.
func (p *Pool) Add(ctx context.Context, task *Task) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isStopped {
		return
	}
	select {
	case <-ctx.Done():
	case p.tCh <- *task:
		return
	}
}

// Get blocks until a task with NeedResult is received from the result channel.
func (p *Pool) Get(ctx context.Context) Task {
	t := Task{}
	select {
	case <-ctx.Done():
		t.Err = ctx.Err()
		return t
	case t, ok := <-p.rCh:
		if !ok {
			t.Err = errors.New("pool closed")
		}
		return t

	}
}

// Shutdown waiting while all workers compete tasks and quit
func (p *Pool) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	p.isStopped = true
	p.mu.Unlock()
	close(p.tCh)
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil

	}
}
