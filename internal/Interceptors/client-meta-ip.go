// Package Interceptors provides gRPC unary interceptors for client and server metadata (x-real-ip, trusted subnet).
package Interceptors

import (
	"context"
	"sys-metrics/internal/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryClientXRealIPInterceptor(localIp string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if localIp != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, common.MetaXRealIP, localIp)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
