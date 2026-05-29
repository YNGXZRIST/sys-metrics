package proto

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func startBufGRPCServer(t *testing.T, srv MetricsServer) (*grpc.ClientConn, func()) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	gs := grpc.NewServer()
	RegisterMetricsServer(gs, srv)
	go func() { _ = gs.Serve(lis) }()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	return conn, func() {
		_ = conn.Close()
		gs.Stop()
	}
}

func TestUnimplementedMetricsServer(t *testing.T) {
	conn, stop := startBufGRPCServer(t, UnimplementedMetricsServer{})
	defer stop()

	client := NewMetricsClient(conn)
	_, err := client.UpdateMetrics(context.Background(), UpdateMetricsRequest_builder{}.Build())
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unimplemented {
		t.Fatalf("err = %v", err)
	}
}

func TestRegisterMetricsServer(t *testing.T) {
	srv := grpc.NewServer()
	RegisterMetricsServer(srv, UnimplementedMetricsServer{})
}
