package main

//./sa --addr=127.0.0.1:9081 --etcd.addr=http://127.0.0.1:2379 --service=goods

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/net/context"

	"github.com/valyala/fasthttp"
)

func main() {
	var (
		addr     = flag.String("addr", ":9080", "TCP address to listen to")
		service  = flag.String("service", "goods", "service name")
		compress = flag.Bool("compress", false, "Whether to enable transparent response compression")
		etcdAddr = flag.String("etcd.addr", "", "Etcd agent address, like: http://192.168.0.1:2379,http://192.168.0.2:2379")
	)

	flag.Parse()

	errc := make(chan error)
	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	h := requestHandler
	if *compress {
		h = fasthttp.CompressHandler(h)
	}

	ctx := context.Background()

	machines := strings.Split(*etcdAddr, ",")
	client, err := NewClient(ctx, machines)
	if err != nil {
		log.Fatalf("etcd client error: %s", err)
	}

	key := fmt.Sprintf("services/%s/%s", *service, *addr)
	client.Register(key, *addr)
	defer func() {
		client.Deregister(key)
	}()

	go func() {
		if err := fasthttp.ListenAndServe(*addr, h); err != nil {
			fmt.Printf("Error in ListenAndServe: %s", err)
			errc <- err
		}
	}()

	<-errc
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hello, world!\n\n")

	// fmt.Fprintf(ctx, "Request method is %q\n", ctx.Method())
	// fmt.Fprintf(ctx, "RequestURI is %q\n", ctx.RequestURI())
	// fmt.Fprintf(ctx, "Requested path is %q\n", ctx.Path())
	// fmt.Fprintf(ctx, "Host is %q\n", ctx.Host())
	// fmt.Fprintf(ctx, "Query string is %q\n", ctx.QueryArgs())
	// fmt.Fprintf(ctx, "User-Agent is %q\n", ctx.UserAgent())
	// fmt.Fprintf(ctx, "Connection has been established at %s\n", ctx.ConnTime())
	// fmt.Fprintf(ctx, "Request has been started at %s\n", ctx.Time())
	// fmt.Fprintf(ctx, "Serial request number for the current connection is %d\n", ctx.ConnRequestNum())
	// fmt.Fprintf(ctx, "Your ip is %q\n\n", ctx.Request.Header.Peek("Client-Real-IP"))
	// fmt.Fprintf(ctx, "Request header is %q\n", ctx.Request.Header.Peek("token"))
	//
	// fmt.Fprintf(ctx, "Raw request is:\n---CUT---\n%s\n---CUT---", &ctx.Request)
	//
	// ctx.SetContentType("text/plain; charset=utf8")
	//
	// // Set arbitrary headers
	// ctx.Response.Header.Set("X-My-Header", "my-header-value")
	//
	// // Set cookies
	// var c fasthttp.Cookie
	// c.SetKey("cookie-name")
	// c.SetValue("cookie-value")
	// ctx.Response.Header.SetCookie(&c)

	fmt.Println("Hello, world!")
}
