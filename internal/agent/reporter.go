package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/common"
	"sys-metrics/internal/errors/labelerrors"
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
	httpClient    *http.Client
	serverAddr    string
	logger        *zap.Logger
	authenticator authenticate.Authenticator
}

func NewReporter(serverAddr string, logger *zap.Logger, a authenticate.Authenticator) *Reporter {
	httpClient := &http.Client{}
	return &Reporter{httpClient, serverAddr, logger, a}
}
func (r *Reporter) Send(c *Collector) error {
	for _, m := range c.Gauges {
		err := r.sendMetricToServer(m.Metrics)
		if err != nil {
			return labelerrors.NewLabelError("SEND METRIC", fmt.Errorf("error send gauge metric to server:: %w", err))
		}
	}
	for _, m := range c.Counters {
		err := r.sendMetricToServer(m.Metrics)
		if err != nil {
			return labelerrors.NewLabelError("SEND METRIC", fmt.Errorf("error send counter metric to server:: %w", err))
		}
	}
	return nil
}
func (r *Reporter) sendMetricsToServer(c *Collector) error {
	reqData := make([]*metrics.Metrics, 0, len(c.Gauges)+len(c.Counters))
	for _, m := range c.Gauges {
		reqData = append(reqData, &m.Metrics)
	}
	for _, m := range c.Counters {
		reqData = append(reqData, &m.Metrics)
	}
	if len(reqData) == 0 {
		return nil
	}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return labelerrors.NewLabelError("SEND METRICS", fmt.Errorf("error marshalling metrics to JSON: %w", err))
	}
	url := r.BuildUpdatesURL()
	res, err := r.sendUpdateRequest(url, jsonData)
	if err != nil {
		return labelerrors.NewLabelError("SEND METRICS", fmt.Errorf("error sending update request: %w", err))
	}
	r.logger.Info("Response ", zap.String("body", string(res)))
	return nil
}
func (r *Reporter) sendMetricToServer(m metrics.Metrics) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("error marshal metric: %w", err)
	}
	url := r.BuildUpdateURL()
	res, err := r.sendUpdateRequest(url, jsonData)
	if err != nil {
		return fmt.Errorf("error send update request: %w", err)
	}
	fmt.Println(string(res))
	return nil

}
func (r *Reporter) sendUpdateRequest(url string, reqData []byte) ([]byte, error) {
	r.logger.Info("Request", zap.String("url", url), zap.String("json", string(reqData)))

	writer := bytes.NewReader(reqData)
	req, err := http.NewRequest(http.MethodPost, url, writer)
	if err != nil {
		return nil, fmt.Errorf("error create request: %w", err)
	}
	req.Header.Set(common.ContentTypeHeader, common.ApplicationJSON)
	req.Header.Set(httpcompressor.AcceptEncodingHeader, httpcompressor.GzipEncoding)
	if r.authenticator != nil {
		key := r.authenticator.GetHashHeaderKey()
		req.Header.Set(key, r.authenticator.SignBody(reqData))
		r.logger.Info("Authenticated request", zap.String("key", key), zap.String("url", url), zap.Any("headers", req.Header))
	}
	response, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error do request: %w", err)
	}
	r.logger.Info("Response status", zap.Int("status", response.StatusCode), zap.String("encoding", response.Header.Get(httpcompressor.ContentEncodingHeader)))
	defer response.Body.Close()
	res, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error read response: %w", err)
	}
	return res, nil
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
func (r *Reporter) BuildUpdatesURL() string {
	return r.serverAddr + "/updates"
}
