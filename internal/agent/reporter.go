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
	"sys-metrics/internal/secure"
	"sys-metrics/pkg/httpcompressor"

	"go.uber.org/zap"
)

// generate:reset

// Request is a simplified metric view for debugging scenarios.
type Request struct {
	ID    string `json:"id"`
	MType string `json:"type"`
	Value string `json:"value"`
}

// generate:reset

// Response holds status code and body from the server for manual sends.
type Response struct {
	Result string
	Code   int
}

// generate:reset

// Reporter posts metrics to the server HTTP API with gzip and optional body signing.
type Reporter struct {
	authenticator    authenticate.Authenticator
	requestEncryptor *secure.RequestEncryptor
	httpClient       *http.Client
	logger           *zap.Logger
	serverAddr       string
}

// NewReporter creates a client that posts to serverAddr (metrics server base URL).
func NewReporter(addr string, l *zap.Logger, a authenticate.Authenticator, r *secure.RequestEncryptor) *Reporter {
	c := &http.Client{}
	return &Reporter{
		authenticator:    a,
		requestEncryptor: r,
		httpClient:       c,
		logger:           l,
		serverAddr:       addr,
	}
}

// Send posts each metric with a separate POST to /update (legacy one-metric path).
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
	plain, err := r.checkAndEncodeRequest(reqData)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}
	req, err := r.createRequest(url, plain)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	r.setHeaders(req, plain)
	if r.authenticator != nil {
		key := r.authenticator.GetHashHeaderKey()
		req.Header.Set(key, r.authenticator.SignBody(plain))
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
func (r *Reporter) checkAndEncodeRequest(reqData []byte) ([]byte, error) {
	var err error
	if r.requestEncryptor != nil && r.requestEncryptor.IsEnabled {
		reqData, err = r.requestEncryptor.Encrypt(reqData)
		if err != nil {
			return nil, fmt.Errorf("error encrypting request: %w", err)
		}
	}
	return reqData, nil
}
func (r *Reporter) setHeaders(req *http.Request, plaintext []byte) {
	req.Header.Set(httpcompressor.AcceptEncodingHeader, httpcompressor.GzipEncoding)
	if r.requestEncryptor != nil && r.requestEncryptor.IsEnabled {
		req.Header.Set(common.EncryptHeader, common.RSA)
	} else {
		req.Header.Set(common.ContentTypeHeader, common.ApplicationJSON)
	}
	if r.authenticator != nil {
		key := r.authenticator.GetHashHeaderKey()
		req.Header.Set(key, r.authenticator.SignBody(plaintext))
	}

}
func (r *Reporter) createRequest(url string, plain []byte) (*http.Request, error) {
	writer := bytes.NewReader(plain)
	req, err := http.NewRequest(http.MethodPost, url, writer)
	if err != nil {
		return nil, fmt.Errorf("error create request: %w", err)
	}
	return req, nil
}

// ConvertMetricValue formats a number as an integer for counters or float for gauges.
func (r *Reporter) ConvertMetricValue(m string, v float64) string {
	var s string
	if m == common.Counter {
		s = strconv.FormatInt(int64(v), 10)
	} else {
		s = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return s
}

// BuildUpdateURL returns the single-metric update endpoint URL.
func (r *Reporter) BuildUpdateURL() string {
	return r.serverAddr + "/update"
}

// BuildUpdatesURL returns the batch update URL for /updates.
func (r *Reporter) BuildUpdatesURL() string {
	return r.serverAddr + "/updates"
}
