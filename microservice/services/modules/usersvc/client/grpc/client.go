// Package grpc provides a gRPC client for the add service.
package grpc

import (
	"time"

	stdopentracing "github.com/opentracing/opentracing-go"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	jujuratelimit "github.com/juju/ratelimit"

	"go/microservice/services/modules/usersvc"
	"go/microservice/services/modules/usersvc/pb"
)

// New returns an AddService backed by a gRPC client connection. It is the
// responsibility of the caller to dial, and later close, the connection.
func New(conn *grpc.ClientConn, tracer stdopentracing.Tracer, logger log.Logger) usersvc.Service {
	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.

	//qps限制在10000
	limiter := ratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(10000, 10000))

	var getUserinfoEndpoint endpoint.Endpoint
	{
		getUserinfoEndpoint = grpctransport.NewClient(
			conn,
			"Usersvc",
			"GetUserinfo",
			usersvc.EncodeGRPCGetUserinfoRequest,
			usersvc.DecodeGRPCGetUserinfoResponse,
			pb.GetUserinfoResponse{},
			grpctransport.ClientBefore(opentracing.ToGRPCRequest(tracer, logger)),
		).Endpoint()
		getUserinfoEndpoint = opentracing.TraceClient(tracer, "GetUserinfo")(getUserinfoEndpoint)
		getUserinfoEndpoint = limiter(getUserinfoEndpoint)
		getUserinfoEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "GetUserinfo",
			Timeout: 30 * time.Second,
		}))(getUserinfoEndpoint)
	}

	return usersvc.Endpoints{
		GetUserinfoEndpoint: getUserinfoEndpoint,
	}
}
