package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/pkg/httpcompressor"

	"go.uber.org/zap"
)

type Request struct {
	ID    string `json:"id"`
	MType string `json:"type"`
	Value string `json:"value"`
}
type Response struct {
	Code   int
	Result string
}
type Reporter struct {
	httpClient *http.Client
	serverAddr string
	logger     *zap.Logger
}

func NewReporter(serverAddr string, logger *zap.Logger) *Reporter {
	httpClient := &http.Client{}
	return &Reporter{httpClient, serverAddr, logger}
}
func (r *Reporter) Send(c *Collector) error {
	for _, m := range c.Gauges {
		err := r.sendMetricToServer(m.Metrics)
		if err != nil {
			return fmt.Errorf("error send gauge metric to server: %w", err)
		}
	}
	for _, m := range c.Counters {
		err := r.sendMetricToServer(m.Metrics)
		if err != nil {
			return fmt.Errorf("error send counter metric to server: %w", err)
		}
	}
	return nil
}
func (r *Reporter) sendMetricToServer(m metrics.Metrics) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("error marshal metric: %w", err)
	}
	writer := bytes.NewReader(jsonData)
	url := r.BuildUpdateURL()
	r.logger.Info(url)
	r.logger.Info("request:" + string(jsonData))
	req, err := http.NewRequest(http.MethodPost, url, writer)
	if err != nil {
		return fmt.Errorf("error create request: %w", err)
	}
	req.Header.Set(common.ContentTypeHeader, common.ApplicationJSON)
	req.Header.Set(httpcompressor.AcceptEncodingHeader, httpcompressor.GzipEncoding)
	response, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error do request: %w", err)
	}
	r.logger.Info("response encoding: " + response.Header.Get(httpcompressor.ContentEncodingHeader))
	defer response.Body.Close()
	res, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error read response: %w", err)
	}
	r.logger.Info("response: " + response.Status + "\n" + string(res))
	r.logger.Info(strconv.Itoa(response.StatusCode))
	return nil
}
func (r *Reporter) ConvertMetricValue(m string, v float64) string {
	var s string
	if m == common.Counter {
		s = strconv.FormatInt(int64(v), 10)
	} else {
		s = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return s
}
func (r *Reporter) BuildUpdateURL() string {
	return r.serverAddr + "/update"
}
