package sender

import (
	"sys-metrics/internal/authenticate"
	"sys-metrics/internal/common"
	"sys-metrics/internal/secure"

	"go.uber.org/zap"
)

type SenderConfig struct {
	Transport        string
	ServerURL        string
	Endpoint         string
	Logger           *zap.Logger
	LocalIPv4        string
	Authenticator    authenticate.Authenticator
	RequestEncryptor *secure.RequestEncryptor
}

func NewMetricsSender(cfg SenderConfig) (MetricsSender, error) {

	switch cfg.Transport {
	case common.ReportTransportGRPC:
		client, err := newGrpcSender(grpcSenderConfig{
			senderDeps: depsFrom(cfg),
			endpoint:   cfg.Endpoint,
		})
		return client, err
	default:
		client := newHttpSender(httpSenderConfig{
			senderDeps:       depsFrom(cfg),
			authenticator:    cfg.Authenticator,
			requestEncryptor: cfg.RequestEncryptor,
			serverURL:        cfg.ServerURL,
		})
		return client, nil
	}
}

func depsFrom(cfg SenderConfig) senderDeps {
	return senderDeps{
		logger:    cfg.Logger,
		localIPv4: cfg.LocalIPv4,
	}
}
