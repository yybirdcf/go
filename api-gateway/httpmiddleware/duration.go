package httpmiddleware

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/valyala/fasthttp"
)

type durationHandler struct {
	duration            metrics.Histogram
	requestResponseSize metrics.Histogram
	httpResponsesTotal  metrics.Counter
	handler             func(ctx *fasthttp.RequestCtx)
}

func (d *durationHandler) ServeHTTP(ctx *fasthttp.RequestCtx) {
	defer func(begin time.Time) {
		code := fmt.Sprint(ctx.Response.StatusCode())
		d.httpResponsesTotal.With("code", code).Add(1)
		d.duration.With("code", code).Observe(time.Since(begin).Seconds())
		d.requestResponseSize.With("size", "request").Observe(float64(len(ctx.Request.Body())))
		d.requestResponseSize.With("size", "response").Observe(float64(len(ctx.Response.Body())))
	}(time.Now())

	d.handler(ctx)
}

func DurationHandler(d metrics.Histogram, hrt metrics.Counter, rrs metrics.Histogram, h func(ctx *fasthttp.RequestCtx)) func(ctx *fasthttp.RequestCtx) {
	dh := durationHandler{
		duration:            d,
		httpResponsesTotal:  hrt,
		requestResponseSize: rrs,
		handler:             h,
	}

	return dh.ServeHTTP
}
