package workerpool

import (
	"context"
	"testing"
)

func BenchmarkPool_AddGet(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := NewPool(4)
	p.StartBg(ctx)

	task := NewTask(func(any) (any, error) {
		return 1, nil
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.Add(ctx, task)
		_ = p.Get(ctx)
	}
}
