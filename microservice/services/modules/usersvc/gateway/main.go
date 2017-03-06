package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	etcdsd "github.com/go-kit/kit/sd/etcd"
	"github.com/go-kit/kit/sd/lb"

	"go/microservice/services/modules/usersvc"
	usersvcgrpcclient "go/microservice/services/modules/usersvc/client/grpc"
	"go/microservice/services/modules/usersvc/pb"
)

func main() {
	var (
		debugAddr    = flag.String("debug.addr", ":8088", "Debug and metrics listen address")
		grpcAddr     = flag.String("grpc.addr", ":8082", "gRPC (HTTP) listen address")
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		etcdAddr     = flag.String("etcd.addr", "", "Etcd agent address, like: http://192.168.0.1:2379,http://192.168.0.2:2379")
		retryMax     = flag.Int("retry.max", 3, "per-request retries to different instances")
		retryTimeout = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
	)
	flag.Parse()

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}
	logger.Log("msg", "hello")
	defer logger.Log("msg", "goodbye")

	// Service discovery domain.
	var client etcdsd.Client
	{
		machines := strings.Split(*etcdAddr, ",")
		client, _ = etcdsd.NewClient(context.Background(), machines, etcdsd.ClientOptions{})
	}

	var (
		endpoints = usersvc.Endpoints{}
	)
	{
		factory := usersvcFactory(usersvc.MakeGetUserinfoEndpoint, logger)
		subscriber, err := etcdsd.NewSubscriber(client, "services/usersvc/", factory, logger)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}

		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(*retryMax, *retryTimeout, balancer)
		endpoints.GetUserinfoEndpoint = retry
	}

	// Mechanical domain.
	errc := make(chan error)
	ctx := context.Background()

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, usersvc.MakeHTTPHandler(ctx, endpoints, logger))
	}()

	// gRPC transport.
	go func() {
		logger := log.NewContext(logger).With("transport", "gRPC")

		ln, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			errc <- err
			return
		}

		srv := usersvc.MakeGRPCServer(ctx, endpoints, logger)
		s := grpc.NewServer()
		pb.RegisterUsersvcServer(s, srv)

		logger.Log("addr", *grpcAddr)
		errc <- s.Serve(ln)
	}()

	// Debug listener.
	go func() {
		logger := log.NewContext(logger).With("transport", "debug")

		m := http.NewServeMux()
		m.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		m.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		m.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		m.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		m.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		m.Handle("/metrics", stdprometheus.Handler())

		logger.Log("addr", *debugAddr)
		errc <- http.ListenAndServe(*debugAddr, m)
	}()

	// Run!
	logger.Log("exit", <-errc)
}

func usersvcFactory(makeEndpoint func(usersvc.Service) endpoint.Endpoint, logger log.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
		if err != nil {
			return nil, nil, err
		}
		service := usersvcgrpcclient.New(conn, logger)
		endpoint := makeEndpoint(service)

		return endpoint, conn, nil
	}
}
