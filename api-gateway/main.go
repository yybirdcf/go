package main

import (
	"api-gateway/httpmiddleware"
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/howeyc/fsnotify"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
)

func main() {
	var (
		path = flag.String("config", "./config.dev.yml", "Config file path")
	)
	flag.Parse()

	//初始化log组件
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var cfg *Config
	{
		cfg = GetConfig(*path)
	}

	//初始化日志文件
	accessName := fmt.Sprintf("%s/access.log", cfg.Log.Dir)
	faccess, err := createFile(accessName)
	if err != nil {
		logger.Log("log_access", err.Error())
		return
	}
	errorName := fmt.Sprintf("%s/error.log", cfg.Log.Dir)
	ferror, err := createFile(errorName)
	if err != nil {
		logger.Log("log_error", err.Error())
		return
	}

	//监控access文件
	go func() {
		watch, err := fsnotify.NewWatcher()
		if err != nil {
			logger.Log("log_file_monitor", err)
			return
		}

		err = watch.WatchFlags(cfg.Log.Dir, fsnotify.FSN_DELETE|fsnotify.FSN_RENAME)
		if err != nil {
			logger.Log("monitor_err", err)
		}

		go func() {
			for {
				select {
				case w := <-watch.Event:
					if w.IsDelete() || w.IsRename() {
						faccess.Close()
						ferror.Close()

						faccess, err = createFile(accessName)
						if err != nil {
							logger.Log("log_error", err.Error())
						}

						ferror, err = createFile(errorName)
						if err != nil {
							logger.Log("log_error", err.Error())
						}
					}
				case err := <-watch.Error:
					logger.Log("monitor_err", err)
				}
			}
		}()
	}()

	var duration metrics.Histogram
	{
		//请求耗时metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "apigateway",
			Name:      "request_duration_ns",
			Help:      "Request duration in nanoseconds.",
		}, []string{"code"})
	}
	var httpRequestResponseSize metrics.Histogram
	{
		//请求响应数据大小metrics.
		httpRequestResponseSize = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "apigateway",
			Name:      "request_response_size_bytes",
			Help:      "Request and response byte size.",
		}, []string{"size"})
	}
	var httpResponsesTotal metrics.Counter
	{
		httpResponsesTotal = prometheus.NewCounterFrom(
			stdprometheus.CounterOpts{
				Namespace: "apigateway",
				Name:      "http_responses_total",
				Help:      "The count of http responses issued, classified by code.",
			}, []string{"code"})
	}
	var servicesCounter metrics.Counter
	{
		servicesCounter = prometheus.NewCounterFrom(
			stdprometheus.CounterOpts{
				Namespace: "apigateway",
				Name:      "service_request",
				Help:      "Total count of every service requests",
			}, []string{"name", "service", "code"})
	}

	errc := make(chan error)

	//启动debug服务
	go func() {
		logger = log.With(logger, "transport", "debug")

		m := http.NewServeMux()
		m.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		m.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		m.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		m.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		m.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
		m.Handle("/metrics", stdprometheus.Handler())

		logger.Log("addr", cfg.Debug)
		errc <- http.ListenAndServe(cfg.Debug, m)
	}()

	ctx := context.Background()

	//开启订阅etcd服务
	var sub *Subscribe
	{
		sub, err = NewSubscribe(ctx, cfg.Etcd, "/services/")
		if err != nil {
			logger.Log("subscribe", err.Error())
			return
		}
	}

	//订阅客户端
	var client *Subclient
	{
		client, err = NewSubclient(ctx, cfg.Etcd, "/services/")
		if err != nil {
			logger.Log("subscribe client", err.Error())
			return
		}
	}

	//启动proxy http服务
	var proxy *Proxy
	{
		proxy = NewProxy(sub, servicesCounter)

		//http
		go func() {
			logger = log.With(logger, "transport", "HTTP")
			logger.Log("proxy http addr", cfg.Proxy)
			// errc <- fasthttp.ListenAndServe(cfg.Proxy,
			// 	httpmiddleware.LoggingHandler(os.Stdout,
			// 		httpmiddleware.LimitHandler(time.Duration(cfg.Limit.TTL)*time.Millisecond, cfg.Limit.Max,
			// 			httpmiddleware.RecoveryHandler(os.Stdout,
			// 				httpmiddleware.DurationHandler(duration, httpResponsesTotal, httpRequestResponseSize, proxy.ServeHTTP)))))
			errc <- fasthttp.ListenAndServe(cfg.Proxy,
				httpmiddleware.LoggingHandler(&faccess,
					httpmiddleware.RecoveryHandler(&ferror,
						httpmiddleware.DurationHandler(duration, httpResponsesTotal, httpRequestResponseSize, proxy.ServeHTTP))))
		}()

		//https
		if cfg.Https.Enable {
			go func() {
				logger = log.With(logger, "transport", "HTTP")
				logger.Log("proxy http addr", cfg.Proxy)
				errc <- fasthttp.ListenAndServeTLS(cfg.Https.Addr, cfg.Https.Cert, cfg.Https.Key,
					httpmiddleware.LoggingHandler(&faccess,
						httpmiddleware.LimitHandler(time.Duration(cfg.Limit.TTL)*time.Millisecond, cfg.Limit.Max,
							httpmiddleware.RecoveryHandler(&ferror,
								httpmiddleware.DurationHandler(duration, httpResponsesTotal, httpRequestResponseSize, proxy.ServeHTTP)))))
			}()
		}
	}

	//启动admin管理工具http服务
	var admin *Admin
	{
		admin, err = NewAdmin(logger, cfg, proxy, client)
		if err != nil {
			logger.Log("admin", err.Error())
			return
		}

		go func() {
			logger = log.With(logger, "transport", "HTTP")
			logger.Log("admin http addr", cfg.Admin)

			errc <- fasthttp.ListenAndServe(cfg.Admin, admin.HandleFastHttp)
		}()
	}

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("msg", "hello")
	defer func() {
		faccess.Close()
		ferror.Close()
		logger.Log("msg", "goodbye")
	}()

	logger.Log("exit", <-errc)
}
