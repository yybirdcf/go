package httpmiddleware

import (
	"os"
	"runtime"

	"github.com/valyala/fasthttp"
)

type recoveryHandler struct {
	writer  **os.File
	handler func(ctx *fasthttp.RequestCtx)
}

func (r *recoveryHandler) ServeHTTP(ctx *fasthttp.RequestCtx) {
	defer func() {
		if re := recover(); re != nil {
			ctx.Response.Header.SetStatusCode(fasthttp.StatusInternalServerError)
			//获取错误堆栈信息
			stack := make([]byte, 4*1024)
			stack = stack[:runtime.Stack(stack, false)]

			(*r.writer).Write(stack)
		}
	}()

	r.handler(ctx)
}

func RecoveryHandler(out **os.File, h func(ctx *fasthttp.RequestCtx)) func(ctx *fasthttp.RequestCtx) {
	rh := recoveryHandler{
		writer:  out,
		handler: h,
	}

	return rh.ServeHTTP
}
