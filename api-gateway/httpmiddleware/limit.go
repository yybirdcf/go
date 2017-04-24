package httpmiddleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/valyala/fasthttp"

	"golang.org/x/time/rate"
)

type limitHandler struct {
	limits  map[string]*rate.Limiter
	mutex   sync.RWMutex
	ttl     time.Duration
	max     int
	handler func(ctx *fasthttp.RequestCtx)
}

func (lh *limitHandler) ServeHTTP(ctx *fasthttp.RequestCtx) {

	//同一ip限制策略
	ip := ctx.RemoteIP().String()
	if lh.limitReached(ip) {
		ctx.Response.Header.SetStatusCode(fasthttp.StatusTooManyRequests)
		fmt.Fprintf(ctx, "too many requests")
		return
	}

	//同一用户限制策略

	lh.handler(ctx)
}

func (lh *limitHandler) limitReached(key string) bool {
	lh.mutex.Lock()
	defer lh.mutex.Unlock()
	if _, found := lh.limits[key]; !found {
		lh.limits[key] = rate.NewLimiter(rate.Every(lh.ttl), lh.max)
	}

	return !lh.limits[key].AllowN(time.Now(), 1)
}

func LimitHandler(ttl time.Duration, max int, h func(ctx *fasthttp.RequestCtx)) func(ctx *fasthttp.RequestCtx) {
	lh := limitHandler{
		limits:  make(map[string]*rate.Limiter),
		ttl:     ttl,
		max:     max,
		handler: h,
	}

	return lh.ServeHTTP
}
