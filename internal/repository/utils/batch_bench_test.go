package utils

import (
	"context"
	"fmt"
	"math/rand"
	"sys-metrics/internal/common"
	models "sys-metrics/internal/model/metrics"
	"testing"
)

func BenchmarkUtils_ApplyBatchToStorages_Mixed(b *testing.B) {
	ctx := context.Background()
	r := rand.New(rand.NewSource(1))
	const n = 256
	payload := make([]models.Metrics, 0, n)
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			name := fmt.Sprintf("g_%d_%x", i, r.Uint64())
			v := r.Float64() * 1e6
			payload = append(payload, models.Metrics{ID: name, MType: common.Gauge, Value: &v})
		} else {
			name := fmt.Sprintf("c_%d_%x", i, r.Uint64())
			v := int64(r.Intn(1_000_000))
			payload = append(payload, models.Metrics{ID: name, MType: common.Counter, Delta: &v})
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gauges, counters := newTestStorages()
		_, err := ApplyBatchToStorages(ctx, payload, gauges, counters)
		if err != nil {
			b.Fatal(err)
		}
	}
}
