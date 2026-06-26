// Package common holds shared constants used across the whole module:
// metric type names, runtime mode strings, HTTP headers, and context keys.
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
	TotalMemory   = "TotalMemory"
	FreeMemory    = "FreeMemory"
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
const (
	EncryptHeader = "X-Encrypted"
	RSA           = "RSA"
)

const (
	SHA256     = "sha256"
	HashSHA256 = "HashSHA256"
)
const NotApplicable = "N/A"
const (
	HeaderXRealIP = "X-Real-Ip"
	MetaXRealIP   = "x-real-ip"
)

const (
	ReportTransportHTTP = "http"
	ReportTransportGRPC = "grpc"
)
