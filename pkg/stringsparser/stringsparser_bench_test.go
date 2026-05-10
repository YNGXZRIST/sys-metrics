package stringsparser

import "testing"

func BenchmarkCapitalize_Short(b *testing.B) {
	const s = "alloc"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Capitalize(s)
	}
}

func BenchmarkCapitalize_Medium(b *testing.B) {
	const s = "SomeMetricNameFromCollector"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Capitalize(s)
	}
}
