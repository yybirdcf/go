package httpmiddleware

import (
	"io"
	"runtime"

	"github.com/valyala/fasthttp"
)

type recoveryHandler struct {
	writer  io.Writer
	handler func(ctx *fasthttp.RequestCtx)
}

func (r *recoveryHandler) ServeHTTP(ctx *fasthttp.RequestCtx) {
	defer func() {
		if re := recover(); re != nil {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
			//获取错误堆栈信息
			stack := make([]byte, 4*1024)
			stack = stack[:runtime.Stack(stack, false)]

			r.writer.Write(stack)
		}
	}()

	r.handler(ctx)
}

func RecoveryHandler(out io.Writer, h func(ctx *fasthttp.RequestCtx)) func(ctx *fasthttp.RequestCtx) {
	rh := recoveryHandler{
		writer:  out,
		handler: h,
	}

	return rh.ServeHTTP
}
