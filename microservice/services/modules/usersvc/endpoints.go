package usersvc

import (
	"fmt"
	"runtime"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

type Endpoints struct {
	GetUserinfoEndpoint endpoint.Endpoint
}

func (e Endpoints) GetUserinfo(ctx context.Context, id int64) (Userinfo, error) {
	request := getUserinfoRequest{Id: id}
	response, err := e.GetUserinfoEndpoint(ctx, request)
	if err != nil {
		return Userinfo{}, err
	}

	return response.(getUserinfoResponse).V, response.(getUserinfoResponse).Err
}

// MakeSumEndpoint returns an endpoint that invokes Sum on the service.
// Primarily useful in a server.
func MakeGetUserinfoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getUserinfoRequest)
		v, err := s.GetUserinfo(ctx, req.Id)

		return getUserinfoResponse{
			V:   v,
			Err: err,
		}, nil
	}
}

// EndpointInstrumentingMiddleware returns an endpoint middleware that records
// the duration of each invocation to the passed histogram. The middleware adds
// a single field: "success", which is "true" if no error is returned, and
// "false" otherwise.
func EndpointInstrumentingMiddleware(duration metrics.Histogram) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			defer func(begin time.Time) {
				duration.With("success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
			}(time.Now())
			return next(ctx, request)
		}
	}
}

// EndpointLoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, and the resulting error, if any.
func EndpointLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			defer func(begin time.Time) {
				logger.Log("error", err, "request", fmt.Sprintf("%+v", request), "took", time.Since(begin))
			}(time.Now())
			return next(ctx, request)
		}
	}
}

func EndpointRecoveryMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			defer func() {
				if r := recover(); r != nil {
					//获取错误堆栈信息
					stack := make([]byte, 4*1024)
					stack = stack[:runtime.Stack(stack, false)]
					logger.Log("error", string(stack), "request", fmt.Sprintf("%+v", request))
				}
			}()
			return next(ctx, request)
		}
	}
}

type getUserinfoRequest struct {
	Id int64
}

type getUserinfoResponse struct {
	V   Userinfo
	Err error
}

type getUserinfoResponseHttp struct {
	V   Userinfo
	Err string
}

type Userinfo struct {
	Id             int64
	Username       string
	Phone          string
	Sex            int64
	Avatar         string
	Gouhao         int64
	Birthday       int64
	Avatars        string
	Signature      string
	Appfrom        string
	Appver         string
	BackgroudImage string
	UpdateAppver   string
	Privacy        int64
	LoadRecTags    int64
	GamePower      int64
	Mark           int64
	Level          int64
	QuestionPhoto  string
	Lan            string
	Notify         int64
	BindGameIds    []int64
	UserPosition   string
	UserStatus     int64
	ImToken        string
}
