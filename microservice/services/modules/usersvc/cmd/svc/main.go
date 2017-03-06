package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	etcdsd "github.com/go-kit/kit/sd/etcd"

	"go/microservice/services/config"
	"go/microservice/services/modules/usersvc"
	"go/microservice/services/modules/usersvc/pb"
)

func main() {
	var (
		debugAddr = flag.String("debug.addr", ":8080", "Debug and metrics listen address")
		etcdAddr  = flag.String("etcd.addr", "", "Etcd agent address, like: http://192.168.0.1:2379,http://192.168.0.2:2379")
		grpcAddr  = flag.String("grpc.addr", ":8082", "gRPC (HTTP) listen address")
		httpAddr  = flag.String("http.addr", ":8081", "HTTP listen address")
		etcdName  = flag.String("etcd.name", "", "etcd namespace: like grpc 10.25.11.78:8082")
		nodeName  = flag.String("localredis.name", "", "localredis cache namespace: like 10.25.11.78")
		cfgEnv    = flag.String("env", "dev", "config environment: dev, test, pro")
	)
	flag.Parse()

	var cfg = config.GetConfig(*cfgEnv)

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
	defer func() {
		//clean etcd register
		client.Deregister(etcdsd.Service{
			Key: fmt.Sprintf("services/usersvc/%s", *etcdName),
		})
	}()

	// Metrics domain.
	var getUserinfoRequests metrics.Counter
	{
		// Business level metrics.
		getUserinfoRequests = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "usersvc",
			Name:      "getUserinfo_requests",
			Help:      "Total count of requests via the GetUserinfo method.",
		}, []string{})
	}
	var duration metrics.Histogram
	{
		// Transport level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "usersvc",
			Name:      "request_duration_ns",
			Help:      "Request duration in nanoseconds.",
		}, []string{"method", "success"})
	}

	// Business domain.
	var service usersvc.Service
	{
		service = usersvc.NewBasicService(cfg, *nodeName)
		service = usersvc.ServiceLoggingMiddleware(logger)(service)
		service = usersvc.ServiceInstrumentingMiddleware(
			getUserinfoRequests)(service)
	}

	// Endpoint domain.
	var getUserinfoEndpoint endpoint.Endpoint
	{
		getUserinfoDuration := duration.With("method", "GetUserinfo")
		getUserinfoLogger := log.NewContext(logger).With("method", "GetUserinfo")

		getUserinfoEndpoint = usersvc.MakeGetUserinfoEndpoint(service)
		getUserinfoEndpoint = usersvc.EndpointInstrumentingMiddleware(getUserinfoDuration)(getUserinfoEndpoint)
		getUserinfoEndpoint = usersvc.EndpointLoggingMiddleware(getUserinfoLogger)(getUserinfoEndpoint)
		getUserinfoEndpoint = usersvc.EndpointRecoveryMiddleware(getUserinfoLogger)(getUserinfoEndpoint)
	}

	endpoints := usersvc.Endpoints{
		GetUserinfoEndpoint: getUserinfoEndpoint,
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

	// HTTP transport.
	go func() {
		logger := log.NewContext(logger).With("transport", "HTTP")
		h := usersvc.MakeHTTPHandler(ctx, endpoints, logger)
		logger.Log("addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, h)
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

	//etcd register
	go func() {
		err := client.Register(etcdsd.Service{
			Key:   fmt.Sprintf("services/usersvc/%s", *etcdName),
			Value: *etcdName,
		})
		if err != nil {
			errc <- err
			return
		}
	}()

	// Run!
	logger.Log("exit", <-errc)
}
