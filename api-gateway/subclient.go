package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
)

type Subclient struct {
	ctx     context.Context
	keysApi etcd.KeysAPI
	prefix  string //watch prefix
}

func NewSubclient(ctx context.Context, machines []string, prefix string) (*Subclient, error) {
	client := &Subclient{
		ctx:    ctx,
		prefix: prefix,
	}

	//new etcd client
	c, err := etcd.New(etcd.Config{
		Endpoints:               machines,
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: 3 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	client.keysApi = etcd.NewKeysAPI(c)

	return client, nil
}

func (c *Subclient) Register(key string, value string) error {
	if key == "" {
		return errors.New("key empty")
	}
	if value == "" {
		return errors.New("value empty")
	}

	key = fmt.Sprintf("%s%s", c.prefix, key)
	_, err := c.keysApi.Create(c.ctx, key, value)
	return err
}

func (c *Subclient) Deregister(key string) error {
	if key == "" {
		return errors.New("key empty")
	}

	key = fmt.Sprintf("%s%s", c.prefix, key)
	_, err := c.keysApi.Delete(c.ctx, key, &etcd.DeleteOptions{
		Recursive: true,
	})
	return err
}

func (c *Subclient) GetEntries() (map[string][]string, error) {
	data := make(map[string][]string)
	resp, err := c.keysApi.Get(c.ctx, c.prefix, &etcd.GetOptions{Recursive: true})
	if err != nil {
		return nil, err
	}

	if len(resp.Node.Nodes) == 0 && resp.Node.Value != "" {
		return data, nil
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
		data[s] = instanceList
	}

	return data, nil
}
