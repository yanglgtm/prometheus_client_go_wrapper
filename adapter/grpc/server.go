package grpc

import (
	"context"
	"time"

	pw "github.com/itsmikej/prometheus_client_go_wrapper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type AdapterGrpcServer struct {
	prom *pw.PrometheusWrapper
}

func NewAdapterGrpcServer(p *pw.PrometheusWrapper) *AdapterGrpcServer {
	return &AdapterGrpcServer{prom: p}
}

func (a *AdapterGrpcClient) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		b := time.Now()
		resp, err := handler(ctx, req)
		st, _ := status.FromError(err)
		serviceName, methodName := splitMethodName(info.FullMethod)
		codeStr := st.Code().String()
		if err != nil {
			a.prom.ExceptionLog(info.FullMethod, codeStr)
		}
		a.prom.SummaryLatencyLog(serviceName, methodName, Unary, float64(time.Now().Sub(b).Nanoseconds()/1000000))
		a.prom.RequestLog(serviceName, methodName, Unary, codeStr)
		return resp, err
	}
}

func (a *AdapterGrpcClient) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		b := time.Now()
		err := handler(srv, ss)
		st, _ := status.FromError(err)
		serviceName, methodName := splitMethodName(info.FullMethod)
		codeStr := st.Code().String()
		if err != nil {
			a.prom.ExceptionLog(info.FullMethod, codeStr)
		}
		a.prom.SummaryLatencyLog(serviceName, methodName, ServerStream, float64(time.Now().Sub(b).Nanoseconds()/1000000))
		a.prom.RequestLog(serviceName, methodName, ServerStream, codeStr)
		return err
	}
}
