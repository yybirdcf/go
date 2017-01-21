package main

import (
	"context"
	"flag"
	"fmt"
	"go/microservice/services/modules/usersvc/pb"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type Stat struct {
	N        int    //number of requests
	C        int    //concurrency , number of workers
	grpcAddr string //rpc服务地址
}

//gorouting 执行stat.N / stat.C请求
func (stat *Stat) runWorker(n int) {
	conn, err := grpc.Dial(stat.grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
	}
	defer conn.Close()

	client := pb.NewUsersvcClient(conn)
	for i := 0; i < n; i++ {
		_, err = client.GetUserinfo(context.Background(), &pb.GetUserinfoRequest{Id: 97})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v", err)
		}
	}
}

//起多个worker执行任务，每一个执行一样数量请求
func (stat *Stat) runWorkers() {
	var wg sync.WaitGroup
	wg.Add(stat.C)

	for i := 0; i < stat.C; i++ {
		go func() {
			stat.runWorker(stat.N / stat.C)
			wg.Done()
		}()
	}

	wg.Wait()
}

func main() {
	var (
		grpcAddr = flag.String("grpc.addr", "", "gRPC (HTTP) address of usersvc")
		n        = flag.Int("n", 1000, "number of requests")
		c        = flag.Int("c", 100, "concurrency , number of workers")
	)
	flag.Parse()

	stat := &Stat{
		N:        *n,
		C:        *c,
		grpcAddr: *grpcAddr,
	}

	start := time.Now()
	stat.runWorkers()
	spend := time.Now().Sub(start)
	fmt.Printf("%v", spend.Seconds())
}
