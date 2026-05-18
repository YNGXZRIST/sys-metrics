package workerpool

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewTask_ProcessSuccess(t *testing.T) {
	task := NewTask(func(x any) (any, error) {
		return 42, nil
	})

	if !task.NeedResult {
		t.Fatalf("NewTask().NeedResult = false, want true")
	}

	task.process()

	if task.Err != nil {
		t.Fatalf("process() Err = %v, want nil", task.Err)
	}
	if task.Result != 42 {
		t.Fatalf("process() Result = %v, want 42", task.Result)
	}
}

func TestNewTask_ProcessError(t *testing.T) {
	testErr := assertError{}
	task := NewTask(func(x any) (any, error) {
		return nil, testErr
	})

	task.process()

	if task.Err == nil {
		t.Fatalf("process() Err = nil, want non-nil")
	}
	if !errors.Is(testErr, task.Err) {
		t.Fatalf("process() Err = %v, want %v", task.Err, testErr)
	}
}

type assertError struct{}

func (assertError) Error() string { return "assertError" }

func TestWorker_StartBg_WithResult(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tCh := make(chan Task, 1)
	rCh := make(chan Task, 1)

	worker := NewWorker(ctx, tCh, rCh)
	go worker.StartBg()

	task := NewTask(func(x any) (any, error) {
		return "ok", nil
	})

	tCh <- *task

	select {
	case got := <-rCh:
		if got.Err != nil {
			t.Fatalf("worker returned Err = %v, want nil", got.Err)
		}
		if got.Result != "ok" {
			t.Fatalf("worker returned Result = %v, want %v", got.Result, "ok")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timeout waiting for worker result")
	}
}

func TestWorker_StartBg_NoResult(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tCh := make(chan Task, 1)
	rCh := make(chan Task, 1)

	worker := NewWorker(ctx, tCh, rCh)
	go worker.StartBg()

	called := make(chan struct{}, 1)
	task := NewTask(func(x any) (any, error) {
		called <- struct{}{}
		return "ignored", nil
	})
	task.NeedResult = false

	tCh <- *task

	select {
	case <-called:
		// ok, function executed
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("worker did not execute task function")
	}

	if got := len(rCh); got != 0 {
		t.Fatalf("len(rCh) = %d, want 0 when NeedResult is false", got)
	}
}

func TestPool_StartBg_CreatesWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const n = 3
	p := NewPool(n)

	if got := len(p.workers); got != 0 {
		t.Fatalf("len(p.workers) before StartBg = %d, want 0", got)
	}

	p.StartBg(ctx)

	if got := len(p.workers); got != n {
		t.Fatalf("len(p.workers) after StartBg = %d, want %d", got, n)
	}
}

func TestPool_AddAndGet_TasksProcessed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := NewPool(2)
	p.StartBg(ctx)

	task1 := NewTask(func(x any) (any, error) {
		return "first", nil
	})
	task2 := NewTask(func(x any) (any, error) {
		return "second", nil
	})

	p.Add(ctx, task1)
	p.Add(ctx, task2)

	results := make(map[any]bool)
	for i := 0; i < 2; i++ {
		select {
		case got := <-p.rCh:
			if got.Err != nil {
				t.Fatalf("Get() Err = %v, want nil", got.Err)
			}
			results[got.Result] = true
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("timeout waiting for pool result %d", i+1)
		}
	}

	if !results["first"] || !results["second"] {
		t.Fatalf("pool results = %#v, want both \"first\" and \"second\"", results)
	}
}

func TestPool_Get_ReturnsTask(t *testing.T) {
	ctx := context.Background()
	p := NewPool(1)
	want := NewTask(func(x any) (any, error) {
		return "done", nil
	})
	want.Result = "done"

	p.rCh <- *want

	got := p.Get(ctx)
	if got.Err != nil {
		t.Fatalf("Get() Err = %v, want nil", got.Err)
	}
	if got.Result != "done" {
		t.Fatalf("Get() Result = %v, want %v", got.Result, "done")
	}
}

func TestPool_Get_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := NewPool(1).Get(ctx)
	if !errors.Is(got.Err, context.Canceled) {
		t.Fatalf("Get() Err = %v, want context.Canceled", got.Err)
	}
}

func TestPool_Get_ClosedResults(t *testing.T) {
	p := NewPool(1)
	close(p.rCh)

	got := p.Get(context.Background())
	if got.Err == nil || got.Err.Error() != "pool closed" {
		t.Fatalf("Get() Err = %v, want pool closed", got.Err)
	}
}

func TestPool_Shutdown(t *testing.T) {
	ctx := context.Background()
	p := NewPool(1)
	p.StartBg(ctx)

	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	p.Add(ctx, NewTask(func(x any) (any, error) {
		t.Fatal("task added after shutdown was executed")
		return nil, nil
	}))
}

func TestPool_Shutdown_ContextCanceled(t *testing.T) {
	p := NewPool(1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := p.Shutdown(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Shutdown() error = %v, want context.Canceled", err)
	}
}
