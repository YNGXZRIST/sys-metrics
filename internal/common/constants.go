package common

const (
	Gauge   = "gauge"
	Counter = "counter"
)

const (
	PollCount   = "PollCount"
	RandomValue = "RandomValue"

	Alloc         = "Alloc"
	BuckHashSys   = "BuckHashSys"
	Frees         = "Frees"
	GCCPUFraction = "GCCPUFraction"
	GCSys         = "GCSys"
	HeapAlloc     = "HeapAlloc"
	HeapIdle      = "HeapIdle"
	HeapInuse     = "HeapInuse"
	HeapObjects   = "HeapObjects"
	HeapReleased  = "HeapReleased"
	HeapSys       = "HeapSys"
	LastGC        = "LastGC"
	Lookups       = "Lookups"
	MCacheInuse   = "MCacheInuse"
	MCacheSys     = "MCacheSys"
	MSpanInuse    = "MSpanInuse"
	MSpanSys      = "MSpanSys"
	Mallocs       = "Mallocs"
	NextGC        = "NextGC"
	NumForcedGC   = "NumForcedGC"
	NumGC         = "NumGC"
	OtherSys      = "OtherSys"
	PauseTotalNs  = "PauseTotalNs"
	StackInuse    = "StackInuse"
	StackSys      = "StackSys"
	Sys           = "Sys"
	TotalAlloc    = "TotalAlloc"
)
const (
	TypeModeDevelopment = "development"
	TypeModeProduction  = "production"
	TypeModeDefault
	TypeModeTest = "test"
)
const (
	TypeAgent   = "agent"
	TypeServer  = "server"
	TypeBackups = "backups"
)
const (
	ContentTypeHeader = "Content-Type"
	ApplicationJSON   = "application/json"
	TextHTMLUTF8      = "text/html; charset=utf-8"
)

type contextKey string

const (
	ConfigContextKey contextKey = "config"
	ContextLoggerKey contextKey = "logger"
)
