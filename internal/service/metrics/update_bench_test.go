package metrics

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/repository/memory"
	storage "sys-metrics/internal/repository/metrics"
	"testing"
)

func BenchmarkServiceMetrics_Update_Gauge(b *testing.B) {
	storage.Init(memory.NewService())
	ctx := context.Background()
	ms := NewService(nil)

	r := rand.New(rand.NewSource(1))
	const k = 4096
	names := make([]string, k)
	vals := make([]string, k)
	for i := 0; i < k; i++ {
		names[i] = fmt.Sprintf("g_%d_%x", i, r.Uint64())
		vals[i] = strconv.FormatFloat(r.Float64()*1e6, 'f', -1, 64)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		j := i % k
		if err := ms.Update(ctx, common.Gauge, names[j], vals[j]); err != nil {
			b.Fatal(err)
		}
	}
}

// resetEveryNewName caps map growth when each iteration uses a unique name (otherwise b.N can use gigabytes of RAM).
const resetEveryNewName = 65536

func BenchmarkServiceMetrics_Update_Gauge_NewNameEach(b *testing.B) {
	ctx := context.Background()
	ms := NewService(nil)
	const val = "123.456"
	seq := 0

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if seq%resetEveryNewName == 0 {
			storage.Init(memory.NewService())
		}
		name := strconv.Itoa(seq)
		seq++
		if err := ms.Update(ctx, common.Gauge, name, val); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServiceMetrics_Update_Counter_NewNameEach(b *testing.B) {
	ctx := context.Background()
	ms := NewService(nil)
	const val = "7"
	seq := 0

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if seq%resetEveryNewName == 0 {
			storage.Init(memory.NewService())
		}
		name := strconv.Itoa(seq)
		seq++
		if err := ms.Update(ctx, common.Counter, name, val); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServiceMetrics_Update_Counter(b *testing.B) {
	storage.Init(memory.NewService())
	ctx := context.Background()
	ms := NewService(nil)

	r := rand.New(rand.NewSource(2))
	const k = 4096
	names := make([]string, k)
	vals := make([]string, k)
	for i := 0; i < k; i++ {
		names[i] = fmt.Sprintf("c_%d_%x", i, r.Uint64())
		vals[i] = strconv.FormatInt(int64(r.Intn(1_000_000)), 10)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		j := i % k
		if err := ms.Update(ctx, common.Counter, names[j], vals[j]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed(b *testing.B) {
	storage.Init(memory.NewService())
	ctx := context.Background()
	ms := NewService(nil)

	r := rand.New(rand.NewSource(3))
	const batchSize = 256
	payload := make([]metrics.Metrics, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		if i%2 == 0 {
			name := fmt.Sprintf("bg_%d_%x", i, r.Uint64())
			v := r.Float64() * 1e6
			payload = append(payload, metrics.Metrics{ID: name, MType: common.Gauge, Value: &v})
		} else {
			name := fmt.Sprintf("bc_%d_%x", i, r.Uint64())
			v := int64(r.Intn(1_000_000))
			payload = append(payload, metrics.Metrics{ID: name, MType: common.Counter, Delta: &v})
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := ms.BatchUpdateMetrics(ctx, payload); err != nil {
			b.Fatal(err)
		}
	}
}
