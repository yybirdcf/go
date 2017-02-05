package grpc_middleware

import (
	"errors"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func AuthUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromContext(ctx)
		if !ok {
			return nil, errors.New("authentication required")
		}

		if appid, ok := md["appid"]; ok {
			if token, ok := md["token"]; ok {
				if appid[0] == "1" && token[0] == "12345678" {
					//
				} else {
					return nil, errors.New("authentication failed")
				}
			} else {
				return nil, errors.New("authentication appid field token not found")
			}
		} else {
			return nil, errors.New("authentication appid field appid not found")
		}

		return handler(ctx, req)
	}
}

func AuthStream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var ctx = ss.Context()
		md, ok := metadata.FromContext(ctx)
		if !ok {
			return errors.New("authentication required")
		}

		if appid, ok := md["appid"]; ok {
			if token, ok := md["token"]; ok {
				if appid[0] == "1" && token[0] == "12345678" {
					//
				} else {
					return errors.New("authentication failed")
				}
			} else {
				return errors.New("authentication appid field token not found")
			}
		} else {
			return errors.New("authentication appid field appid not found")
		}

		return handler(srv, ss)
	}
}
