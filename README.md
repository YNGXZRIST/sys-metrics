## Итерация 17

Правка функции батчей метрик: gauge/counter обновляются на месте, если метрика уже есть. У gauge `SetValue` не создаёт новый `*float64` каждый раз.

## Сервер,diff heap 


```
      flat  flat%   sum%        cum   cum%
 -516.01kB 11.71% 11.71%  -516.01kB 11.71%  io.init.func1
 -512.05kB 11.62% 23.32%  -512.05kB 11.62%  net/textproto.NewReader (inline)
         0     0% 23.32%  -516.01kB 11.71%  bufio.(*Writer).Flush
         0     0% 23.32%  -516.01kB 11.71%  io.Copy (inline)
         0     0% 23.32%  -516.01kB 11.71%  io.CopyN
         0     0% 23.32%  -516.01kB 11.71%  io.copyBuffer
         0     0% 23.32%  -516.01kB 11.71%  io.discard.ReadFrom
         0     0% 23.32%  -516.01kB 11.71%  net/http.(*chunkWriter).Write
         0     0% 23.32%  -516.01kB 11.71%  net/http.(*chunkWriter).writeHeader
         0     0% 23.32%  -512.05kB 11.62%  net/http.(*conn).readRequest
         0     0% 23.32% -1028.06kB 23.32%  net/http.(*conn).serve
         0     0% 23.32%  -516.01kB 11.71%  net/http.(*response).finishRequest
         0     0% 23.32%  -512.05kB 11.62%  net/http.newTextprotoReader
         0     0% 23.32%  -512.05kB 11.62%  net/http.readRequest
         0     0% 23.32%  -516.01kB 11.71%  sync.(*Pool).Get
```

## Бенч `BatchUpdateMetrics` (до и после)

Было: 269 allocs/op, 90120 B/op. Стало: 141 allocs/op, 81928 B/op (darwin/arm64, 5 прогонов).

Diff mem-профилей по alloc_space:

```
Showing nodes accounting for -2606.64MB, 9.21% of 28298.40MB total
      flat  flat%   sum%        cum   cum%
-2554.16MB  9.03%  9.03% -2554.16MB  9.03%  sys-metrics/internal/repository/utils.ApplyGauge
  -89.05MB  0.31%  9.34%  -126.77MB  0.45%  sys-metrics/internal/repository/memory.(*Service).GetAllMetricsLocked
   74.78MB  0.26%  9.08% -2479.87MB  8.76%  sys-metrics/internal/repository/utils.ApplyBatchToStorages
  -37.72MB  0.13%  9.21%   -37.72MB  0.13%  sys-metrics/pkg/storage.(*MemStorage[go.shape.string,go.shape.*uint8]).All
```

`-benchmem`, сервис, до:

```
goos: darwin
goarch: arm64
pkg: sys-metrics/internal/service/metrics
cpu: Apple M4 Pro
BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed-12    	   54140	     20458 ns/op	   90120 B/op	     269 allocs/op
BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed-12    	   58359	     21397 ns/op	   90120 B/op	     269 allocs/op
BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed-12    	   58044	     20536 ns/op	   90120 B/op	     269 allocs/op
BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed-12    	   58304	     20678 ns/op	   90120 B/op	     269 allocs/op
BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed-12    	   58587	     21010 ns/op	   90120 B/op	     269 allocs/op
PASS
ok  	sys-metrics/internal/service/metrics	7.249s
```

после:

```
goos: darwin
goarch: arm64
pkg: sys-metrics/internal/service/metrics
cpu: Apple M4 Pro
BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed-12    	   52008	     21158 ns/op	   81928 B/op	     141 allocs/op
BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed-12    	   56978	     21079 ns/op	   81928 B/op	     141 allocs/op
BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed-12    	   57106	     20994 ns/op	   81929 B/op	     141 allocs/op
BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed-12    	   57019	     21778 ns/op	   81929 B/op	     141 allocs/op
BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed-12    	   55334	     21064 ns/op	   81928 B/op	     141 allocs/op
PASS
ok  	sys-metrics/internal/service/metrics	7.760s
```
