package storage

import (
	"context"
	"fmt"
	"strconv"
	"testing"
)

func BenchmarkMemStorage_SetGet(b *testing.B) {
	ctx := context.Background()
	s := NewMemStorage[string, int]()
	const keys = 4096
	for i := 0; i < keys; i++ {
		_ = s.Set(ctx, strconv.Itoa(i), i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := strconv.Itoa(i % keys)
		_, _ = s.Get(ctx, k)
	}
}

func BenchmarkMemStorage_All(b *testing.B) {
	ctx := context.Background()
	st := NewMemStorage[string, int]()
	for i := 0; i < 512; i++ {
		_ = st.Set(ctx, fmt.Sprintf("k%d", i), i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = st.All(ctx)
	}
}
