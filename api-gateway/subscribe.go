package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
)

type Subscribe struct {
	ctx     context.Context
	keysApi etcd.KeysAPI
	prefix  string //watch prefix
	mutex   sync.Mutex
	entries map[string][]string //可用service节点
	quitc   chan struct{}
}

func NewSubscribe(ctx context.Context, machines []string, prefix string) (*Subscribe, error) {
	sub := &Subscribe{
		ctx:     ctx,
		prefix:  prefix,
		entries: make(map[string][]string),
		quitc:   make(chan struct{}),
	}

	//new etcd client
	client, err := etcd.New(etcd.Config{
		Endpoints:               machines,
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: 3 * time.Second,
	})

	if err != nil {
		return nil, err
	}
	sub.keysApi = etcd.NewKeysAPI(client)
	sub.keysApi.Create(sub.ctx, fmt.Sprintf("%s%s/%s", prefix, "s", "ip:port"), "ip:port")
	sub.keysApi.Delete(sub.ctx, fmt.Sprintf("%s%s/%s", prefix, "s", "ip:port"), &etcd.DeleteOptions{
		Recursive: true,
	})
	sub.keysApi.Delete(sub.ctx, fmt.Sprintf("%s%s", prefix, "s"), &etcd.DeleteOptions{
		Recursive: true,
	})
	//拉取所有配置
	err = sub.getEntries()
	if err != nil {
		fmt.Printf("failed to retrieve entries: %s\n", err.Error())
		return nil, err
	}

	//开启loop监听
	go sub.loop()

	return sub, nil
}

func (sub *Subscribe) GetEntries(service string) []string {
	sub.mutex.Lock()
	defer sub.mutex.Unlock()

	if _, ok := sub.entries[service]; ok {
		return sub.entries[service]
	}

	return nil
}

func (sub *Subscribe) Stop() {
	close(sub.quitc)
}

func (sub *Subscribe) loop() {
	ch := make(chan struct{})
	go sub.watchPrefix(sub.prefix, ch)
	for {
		select {
		case <-ch:
			err := sub.getEntries()
			if err != nil {
				fmt.Printf("failed to retrieve entries: %s\n", err.Error())
				continue
			}
		case <-sub.quitc:
			return
		}
	}
}

func (sub *Subscribe) getEntries() error {
	resp, err := sub.keysApi.Get(sub.ctx, sub.prefix, &etcd.GetOptions{Recursive: true})
	if err != nil {
		return err
	}

	sub.mutex.Lock()
	defer sub.mutex.Unlock()

	//清空之前的配置
	for s, _ := range sub.entries {
		delete(sub.entries, s)
	}

	if len(resp.Node.Nodes) == 0 && resp.Node.Value != "" {
		return nil
	}

	//加载新的配置
	for _, node := range resp.Node.Nodes {
		segs := strings.Split(strings.Trim(node.Key, "/"), "/")
		s := segs[1]
		instanceList := make([]string, len(node.Nodes))
		for i, n := range node.Nodes {
			instance := n.Value
			instanceList[i] = instance
		}
		sub.entries[s] = instanceList
	}

	return nil
}

func (sub *Subscribe) watchPrefix(prefix string, ch chan struct{}) {
	watch := sub.keysApi.Watcher(prefix, &etcd.WatcherOptions{AfterIndex: 0, Recursive: true})
	ch <- struct{}{} // make sure caller invokes GetEntries
	for {
		if _, err := watch.Next(sub.ctx); err != nil {
			return
		}
		ch <- struct{}{}
	}
}
