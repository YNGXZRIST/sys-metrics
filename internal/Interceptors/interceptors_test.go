package Interceptors

import (
	"context"
	"net"
	"sys-metrics/internal/common"
	"sys-metrics/internal/middleware"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryClientXRealIPInterceptor_setsMetadata(t *testing.T) {
	const wantIP = "192.168.1.10"
	ic := UnaryClientXRealIPInterceptor(wantIP)

	err := ic(
		context.Background(),
		"/test.Method",
		nil,
		nil,
		nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Fatal("expected outgoing metadata")
			}
			vals := md.Get(common.MetaXRealIP)
			if len(vals) != 1 || vals[0] != wantIP {
				t.Fatalf("x-real-ip = %v, want %q", vals, wantIP)
			}
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnaryClientXRealIPInterceptor_emptyIP(t *testing.T) {
	ic := UnaryClientXRealIPInterceptor("")

	err := ic(
		context.Background(),
		"/test.Method",
		nil,
		nil,
		nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			if _, ok := metadata.FromOutgoingContext(ctx); ok {
				t.Fatal("expected no outgoing metadata")
			}
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnaryServerXRealIPInterceptor_noMetadata(t *testing.T) {
	ic := UnaryServerXRealIPInterceptor(nil)

	_, err := ic(
		context.Background(),
		nil,
		nil,
		func(ctx context.Context, req any) (any, error) {
			if ctx.Value(middleware.CtxClientIPKey) != nil {
				t.Fatal("expected no client IP in context")
			}
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnaryServerXRealIPInterceptor_setsClientIP(t *testing.T) {
	const wantIP = "10.1.2.3"
	ic := UnaryServerXRealIPInterceptor(nil)
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(common.MetaXRealIP, wantIP),
	)

	_, err := ic(ctx, nil, nil, func(ctx context.Context, req any) (any, error) {
		ip, ok := ctx.Value(middleware.CtxClientIPKey).(string)
		if !ok || ip != wantIP {
			t.Fatalf("client IP = %q, want %q", ip, wantIP)
		}
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnaryServerXRealIPInterceptor_emptyMetadataValue(t *testing.T) {
	ic := UnaryServerXRealIPInterceptor(nil)
	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(common.MetaXRealIP, ""),
	)
	_, err := ic(ctx, nil, nil, func(ctx context.Context, req any) (any, error) {
		if ctx.Value(middleware.CtxClientIPKey) != nil {
			t.Fatal("expected no IP for empty metadata")
		}
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnaryServerXRealIPInterceptor_trustedSubnet(t *testing.T) {
	_, ipNet, err := net.ParseCIDR("127.0.0.0/8")
	if err != nil {
		t.Fatal(err)
	}
	ic := UnaryServerXRealIPInterceptor(ipNet)

	t.Run("allowed", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.Pairs(common.MetaXRealIP, "127.0.0.1"),
		)
		_, err := ic(ctx, nil, nil, func(ctx context.Context, req any) (any, error) {
			return "ok", nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("denied", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.Pairs(common.MetaXRealIP, "8.8.8.8"),
		)
		_, err := ic(ctx, nil, nil, func(ctx context.Context, req any) (any, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.PermissionDenied {
			t.Fatalf("err = %v, want PermissionDenied", err)
		}
	})

	t.Run("invalid ip", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(
			context.Background(),
			metadata.Pairs(common.MetaXRealIP, "not-an-ip"),
		)
		_, err := ic(ctx, nil, nil, func(ctx context.Context, req any) (any, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		})
		st, ok := status.FromError(err)
		if !ok || st.Code() != codes.PermissionDenied {
			t.Fatalf("err = %v, want PermissionDenied", err)
		}
	})
}
