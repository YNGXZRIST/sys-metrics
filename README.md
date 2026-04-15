## Итерация 17. Бенчмарки 

В пакете `internal/repository/utils` доработана логика **`ApplyGauge`** / **`ApplyCounter`** при batch-обновлениях: при наличии метрики в хранилище выполняется **обновление на месте** вместо создания нового объекта на каждый элемент; для gauge в **`internal/model/metrics`** в `SetValue` переиспользуется уже выделенный `*float64`, чтобы не уводить значение в heap на каждое присваивание.

На бенчмарке **`BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed`** (`internal/service/metrics`) число аллокаций на операцию заметно снизилось (порядка **~260 → ~140** `allocs/op` на типичном прогоне. Проверка:

```bash
go test ./internal/service/metrics -run '^$' -bench BenchmarkServiceMetrics_BatchUpdateMetrics_Mixed -benchmem
```
