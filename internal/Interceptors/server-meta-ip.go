package Interceptors

import (
	"context"
	"net"
	"sys-metrics/internal/common"
	"sys-metrics/internal/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerXRealIPInterceptor(ipNet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}
		values := md.Get(common.MetaXRealIP)
		if len(values) == 0 || values[0] == "" {
			return handler(ctx, req)
		}
		ip := values[0]
		if ipNet != nil {
			clientIP := net.ParseIP(ip)
			if clientIP == nil || !ipNet.Contains(clientIP) {
				return nil, status.Error(codes.PermissionDenied, "forbidden")
			}
		}
		ctx = context.WithValue(ctx, middleware.CtxClientIPKey, ip)
		return handler(ctx, req)
	}
}
