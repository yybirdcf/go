package grpc_middleware

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func LogUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var addr string
		if pr, ok := peer.FromContext(ctx); ok {
			if tcpAddr, ok := pr.Addr.(*net.TCPAddr); ok {
				addr = tcpAddr.IP.String()
			} else {
				addr = pr.Addr.String()
			}
		}

		fmt.Printf("date: %s, remote addr: %v, method: %s, request params: %+v \n", time.Now().Format("2006-01-02 15:04:05"), addr, info.FullMethod, req)

		return handler(ctx, req)
	}
}
