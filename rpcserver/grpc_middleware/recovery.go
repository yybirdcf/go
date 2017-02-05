package grpc_middleware

import (
	"fmt"
	"runtime"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func RecoveryUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				//获取错误堆栈信息
				stack := make([]byte, 4*1024)
				stack = stack[:runtime.Stack(stack, false)]
				fmt.Printf("error: %s, request params: %+v\n", string(stack), req)
			}
		}()

		return handler(ctx, req)
	}
}
