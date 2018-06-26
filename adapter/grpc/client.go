package grpc

import (
	"context"
	"time"

	pw "github.com/itsmikej/prometheus_client_go_wrapper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type AdapterGrpcClient struct {
	prom *pw.PrometheusWrapper
}

func NewAdapterGrpcClient(p *pw.PrometheusWrapper) *AdapterGrpcClient {
	return &AdapterGrpcClient{prom: p}
}

func (a *AdapterGrpcClient) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		b := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		serviceName, methodName := splitMethodName(method)
		st, _ := status.FromError(err)
		codeStr := st.Code().String()
		if err != nil {
			a.prom.ExceptionLog(method, codeStr)
		}
		a.prom.SummaryLatencyLog(serviceName, methodName, Unary, float64(time.Now().Sub(b).Nanoseconds()/1000000))
		a.prom.RequestLog(serviceName, methodName, Unary, codeStr)
		return err
	}
}

func (a *AdapterGrpcClient) StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		b := time.Now()
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		serviceName, methodName := splitMethodName(method)
		st, _ := status.FromError(err)
		codeStr := st.Code().String()
		if err != nil {
			a.prom.ExceptionLog(method, codeStr)
		}
		a.prom.SummaryLatencyLog(serviceName, methodName, ClientStream, float64(time.Now().Sub(b).Nanoseconds()/1000000))
		a.prom.RequestLog(serviceName, methodName, ClientStream, codeStr)
		return clientStream, err
	}
}
