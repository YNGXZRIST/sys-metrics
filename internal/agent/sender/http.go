package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/common"
	"sys-metrics/internal/errors/labelerrors"
	"sys-metrics/internal/model/metrics"
	"sys-metrics/internal/secure"
	"sys-metrics/pkg/httpcompressor"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const PathUpdate = "/update"
const PathUpdates = "/updates"

type httpSender struct {
	client *resty.Client
	httpSenderConfig
}
type httpSenderConfig struct {
	senderDeps
	authenticator    authenticate.Authenticator
	requestEncryptor *secure.RequestEncryptor
	serverURL        string
}

var _ MetricsSender = (*httpSender)(nil)

func newHttpSender(cfg httpSenderConfig) *httpSender {
	s := &httpSender{httpSenderConfig: cfg}
	s.initClient()
	return s
}
func (s *httpSender) initClient() {
	s.client = resty.New()
	s.client.SetTimeout(10 * time.Second)
}

func (s *httpSender) SendBatch(ctx context.Context, metrics []*metrics.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return labelerrors.NewLabelError("SEND METRICS", fmt.Errorf("error marshalling metrics to JSON: %w", err))
	}
	url := s.buildUpdatesURL()
	res, errS := s.Request(ctx, url, jsonData)
	if errS != nil {
		return labelerrors.NewLabelError("SEND METRICS", fmt.Errorf("error sending update request: %w", err))
	}
	s.senderDeps.logger.Info("Response ", zap.String("body", string(res)))
	return nil
}

func (s *httpSender) Close() error {
	return nil
}

func (s *httpSender) buildUpdateURL() string {
	return s.serverURL + PathUpdate
}

func (s *httpSender) buildUpdatesURL() string {
	return s.serverURL + PathUpdates
}

func (s *httpSender) Request(ctx context.Context, url string, reqData []byte) ([]byte, error) {
	s.senderDeps.logger.Info("Request", zap.String("url", url), zap.String("json", string(reqData)))
	plain, err := s.checkAndEncodeRequest(reqData)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}
	req := s.client.R().SetContext(ctx).SetBody(plain)
	s.setHeaders(req, plain)
	if s.authenticator != nil {
		key := s.authenticator.GetHashHeaderKey()
		req.Header.Set(key, s.authenticator.SignBody(plain))
	}
	resp, errP := req.Post(url)
	if errP != nil {
		return nil, fmt.Errorf("error do request: %w", errP)
	}
	s.senderDeps.logger.Info("Response status", zap.Int("status", resp.StatusCode()), zap.String("encoding", resp.Header().Get(httpcompressor.ContentEncodingHeader)))
	return resp.Body(), nil
}
func (s *httpSender) sendMetricToServer(ctx context.Context, m metrics.Metrics) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("error marshal metric: %w", err)
	}
	url := s.buildUpdateURL()
	res, err := s.Request(ctx, url, jsonData)
	if err != nil {
		return fmt.Errorf("error send update request: %w", err)
	}
	fmt.Println(string(res))
	return nil

}

func (s *httpSender) setHeaders(req *resty.Request, plaintext []byte) {
	req.Header.Set(httpcompressor.AcceptEncodingHeader, httpcompressor.GzipEncoding)
	if s.requestEncryptor != nil && s.requestEncryptor.IsEnabled {
		req.Header.Set(common.EncryptHeader, common.RSA)
	} else {
		req.Header.Set(common.ContentTypeHeader, common.ApplicationJSON)
	}
	if s.authenticator != nil {
		key := s.authenticator.GetHashHeaderKey()
		req.Header.Set(key, s.authenticator.SignBody(plaintext))
	}
	if s.senderDeps.localIPv4 != "" {
		req.Header.Set(common.HeaderXRealIP, s.senderDeps.localIPv4)
	}

}
func (s *httpSender) checkAndEncodeRequest(reqData []byte) ([]byte, error) {
	var err error
	if s.requestEncryptor != nil && s.requestEncryptor.IsEnabled {
		reqData, err = s.requestEncryptor.Encrypt(reqData)
		if err != nil {
			return nil, fmt.Errorf("error encrypting request: %w", err)
		}
	}
	return reqData, nil
}
