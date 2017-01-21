package main

import (
	"flag"
	"fmt"
	"go/microservice/services/modules/usersvc"
	"os"
	"strconv"
	"time"

	stdopentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	grpcclient "go/microservice/services/modules/usersvc/client/grpc"

	"github.com/go-kit/kit/log"
)

//./cli  --grpc.addr=127.0.0.1:8082  --method=getUserinfo 97

func main() {
	// The addcli presumes no service discovery system, and expects users to
	// provide the direct address of an addsvc. This presumption is reflected in
	// the addcli binary and the the client packages: the -transport.addr flags
	// and various client constructors both expect host:port strings. For an
	// example service with a client built on top of a service discovery system,
	// see profilesvc.

	var (
		grpcAddr = flag.String("grpc.addr", "", "gRPC (HTTP) address of usersvc")
		method   = flag.String("method", "getUserinfo", "getUserinfo, updateUserinfo, insertUserinfo, getUserinfos")
	)
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "usage: cli [flags] ...\n")
		os.Exit(1)
	}

	// This is a demonstration client, which supports multiple tracers.
	// Your clients will probably just use one tracer.
	var tracer stdopentracing.Tracer
	{
		tracer = stdopentracing.GlobalTracer() // no-op
	}

	// This is a demonstration client, which supports multiple transports.
	// Your clients will probably just define and stick with 1 transport.

	var (
		service usersvc.Service
		err     error
	)

	conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	service = grpcclient.New(conn, tracer, log.NewNopLogger())

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch *method {
	case "getUserinfo":
		id, _ := strconv.ParseInt(flag.Args()[0], 10, 64)
		v, err := service.GetUserinfo(context.Background(), id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "getUserinfo(%d) = %v\n", id, v)
	default:
		fmt.Fprintf(os.Stderr, "error: invalid method %q\n", method)
		os.Exit(1)
	}
}
