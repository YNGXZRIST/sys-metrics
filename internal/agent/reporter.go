package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sys-metrics/internal/common"
	"sys-metrics/internal/model/metrics"

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
	serverAddr string
	logger     *zap.Logger
}

func NewReporter(serverAddr string, logger *zap.Logger) *Reporter {
	return &Reporter{serverAddr, logger}
}
func (r *Reporter) Send(c *Collector) error {
	for _, m := range c.Gauges {
		err := r.sendMetricToServer(m.Metrics)
		if err != nil {
			return err
		}
	}
	for _, m := range c.Counters {
		err := r.sendMetricToServer(m.Metrics)
		if err != nil {
			return err
		}
	}
	return nil
}
func (r *Reporter) sendMetricToServer(m metrics.Metrics) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return err
	}
	writer := bytes.NewReader(jsonData)
	url := r.BuildUpdateURL()
	r.logger.Info(url)
	r.logger.Info("request:" + string(jsonData))
	response, err := http.Post(url, common.ApplicationJSON, writer)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	res, err := io.ReadAll(response.Body)
	if err != nil {
		return err
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
