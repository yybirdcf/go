package main

import (
	"flag"
	"fmt"
	"log"
  "encoding/json"
	"github.com/valyala/fasthttp"
)

type Response struct {
  Errcode int
  Errmsg string
  Data interface{}
}

var (
	addr     = flag.String("addr", ":8080", "TCP address to listen to")
	compress = flag.Bool("compress", false, "Whether to enable transparent response compression")
)

func main() {
	flag.Parse()

	h := requestHandler
	if *compress {
		h = fasthttp.CompressHandler(h)
	}

	if err := fasthttp.ListenAndServe(*addr, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {
  path := ctx.Path()
  switch string(path) {
  case "/":
    handleRoot(ctx)
  case "/hello":
    handleHello(ctx)
  default:
    handle404(ctx)
  }

	ctx.SetContentType("application/json; charset=utf8")

	// Set arbitrary headers
	ctx.Response.Header.Set("X-My-Header", "my-header-value")

	// Set cookies
	var c fasthttp.Cookie
	c.SetKey("cookie-name")
	c.SetValue("cookie-value")
	ctx.Response.Header.SetCookie(&c)
}

func handleRoot(ctx *fasthttp.RequestCtx) {
  body, _ := json.Marshal(response(0, "", nil))

  fmt.Fprintf(ctx, "%s", body)
}

func handleHello(ctx *fasthttp.RequestCtx) {
  x := map[string]string{
    "hello": "world",
  }
  body, _ := json.Marshal(response(0, "", x))

  fmt.Fprintf(ctx, "%s", body)
}

func handle404(ctx *fasthttp.RequestCtx) {
  body, _ := json.Marshal(response(404, "404 NOT FOUND", nil))

  fmt.Fprintf(ctx, "%s", body)
}

func response(errcode int, errmsg string, data interface{}) Response{
  return Response{
    Errcode: errcode,
    Errmsg: errmsg,
    Data: data,
  }
}
