package subscribe

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
)

type EtcdSub struct {
	ctx     context.Context
	keysApi etcd.KeysAPI
	prefix  string //watch prefix
	mutex   sync.Mutex
	entries []string //可用service节点
	quitc   chan struct{}
}

func NewEtcdSub(ctx context.Context, machines []string, prefix string) (*EtcdSub, error) {
	etcdSub := &EtcdSub{
		ctx:    ctx,
		prefix: prefix,
		quitc:  make(chan struct{}),
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

	etcdSub.keysApi = etcd.NewKeysAPIWithPrefix(client, etcdSub.prefix)
	//get entries
	etcdSub.entries, err = etcdSub.getEntries()
	if err != nil {
		return nil, err
	}

	//开启loop监听
	go etcdSub.loop()

	return etcdSub, nil
}

func (sub *EtcdSub) GetEntries() []string {
	return sub.entries
}

func (sub *EtcdSub) Stop() {
	close(sub.quitc)
}

func (sub *EtcdSub) loop() {
	ch := make(chan struct{})
	go sub.watchPrefix(sub.prefix, ch)
	for {
		select {
		case <-ch:
			instances, err := sub.getEntries()
			if err != nil {
				fmt.Printf("failed to retrieve entries: %s\n", err.Error())
				continue
			}

			sub.mutex.Lock()
			sub.entries = instances
			sub.mutex.Unlock()

		case <-sub.quitc:
			return
		}
	}
}

func (sub *EtcdSub) getEntries() ([]string, error) {
	resp, err := sub.keysApi.Get(sub.ctx, sub.prefix, &etcd.GetOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	// Special case. Note that it's possible that len(resp.Node.Nodes) == 0 and
	// resp.Node.Value is also empty, in which case the key is empty and we
	// should not return any entries.
	if len(resp.Node.Nodes) == 0 && resp.Node.Value != "" {
		return []string{resp.Node.Value}, nil
	}

	entries := make([]string, len(resp.Node.Nodes))
	for i, node := range resp.Node.Nodes {
		entries[i] = node.Value
	}
	return entries, nil
}

func (sub *EtcdSub) watchPrefix(prefix string, ch chan struct{}) {
	watch := sub.keysApi.Watcher(prefix, &etcd.WatcherOptions{AfterIndex: 0, Recursive: true})
	ch <- struct{}{} // make sure caller invokes GetEntries
	for {
		if _, err := watch.Next(sub.ctx); err != nil {
			return
		}
		ch <- struct{}{}
	}
}
