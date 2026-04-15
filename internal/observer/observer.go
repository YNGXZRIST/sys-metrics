package observer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sys-metrics/internal/logger"
	"sys-metrics/pkg/workerpool"

	"go.uber.org/zap"
)

const (
	typeObserver = "observer"
)

type Observer interface {
	Notify(data any)
	Register(data any) error
}
type MetricObserverConfig struct {
	FilePath  string
	URL       string
	RateLimit int
	Mode      string
}
type MetricsObserver struct {
	cfg              MetricObserverConfig
	ReportWorkerPool *workerpool.Pool
	logger           *zap.Logger
	file             *os.File
	client           *http.Client
}

type MetricsEvent struct {
	Ts      int64    `json:"ts"`
	IP      string   `json:"ip"`
	Metrics []string `json:"metrics"`
}
type MetricObserverOptions func(*MetricsObserver)

func NewMetricsObserver(cfg MetricObserverConfig) (*MetricsObserver, error) {
	log, err := logger.Initialize(cfg.Mode, typeObserver)
	if err != nil {
		return nil, err
	}
	file, err := cfg.openAuditFile()
	if err != nil {
		return nil, err
	}
	client := cfg.CreateClient()
	obs := &MetricsObserver{
		cfg,
		nil,
		log,
		file,
		client,
	}
	return obs, nil
}
func (c *MetricObserverConfig) openAuditFile() (*os.File, error) {
	if c.FilePath == "" {
		return nil, nil
	}
	dir := filepath.Dir(c.FilePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("observer: create audit directory %q: %w", dir, err)
		}
	}
	file, err := os.OpenFile(c.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("observer: open audit file %q: %w", c.FilePath, err)
	}
	return file, nil
}
func (c *MetricObserverConfig) CreateClient() *http.Client {
	if c.URL == "" {
		return nil
	}
	client := &http.Client{}
	return client
}
func (o *MetricsObserver) Register(data any) error {
	ctx, ok := data.(context.Context)
	if !ok {
		return fmt.Errorf("observer: Register observer: invalid data type")

	}
	reportPool := workerpool.NewPool(o.cfg.RateLimit)
	reportPool.StartBg(ctx)
	o.ReportWorkerPool = reportPool
	return nil
}
func (o *MetricsObserver) Notify(data any) {
	if o.ReportWorkerPool == nil {
		return
	}
	event, ok := data.(MetricsEvent)
	if !ok {
		return
	}
	task := o.ReportTask(event)
	o.ReportWorkerPool.Add(task)
}
func (o *MetricsObserver) ReportTask(event MetricsEvent) *workerpool.Task {
	task := workerpool.NewTask(func(a any) (any, error) {
		if o.file != nil {
			go o.SaveToFilePath(event)
		}
		if o.client != nil {
			go o.SendEventToAccrualServer(event)
		}
		return nil, nil
	})
	return task
}
func (o *MetricsObserver) SaveToFilePath(event MetricsEvent) {
	marshal, err := json.Marshal(event)
	if err != nil {
		o.logger.Error("observer: Marshal event failed", zap.Error(err))
		return
	}
	if o.file == nil {
		return
	}
	if _, err := o.file.Write(append(marshal, '\n')); err != nil {
		o.logger.Error("observer: write audit event failed", zap.Error(err))
		return
	}
}
func (o *MetricsObserver) SendEventToAccrualServer(event MetricsEvent) {
	marshal, err := json.Marshal(event)
	if err != nil {
		o.logger.Error("observer: Marshal event failed", zap.Error(err))
		return
	}
	if o.client == nil {
		return
	}
	res, err := o.client.Post(o.cfg.URL, "application/json", bytes.NewBuffer(marshal))
	if err != nil {
		o.logger.Error("observer: send event to accrual server failed", zap.Error(err))
		return
	}
	defer res.Body.Close()
	o.logger.Info("observer: send event to accrual server", zap.Int("status", res.StatusCode))

}
