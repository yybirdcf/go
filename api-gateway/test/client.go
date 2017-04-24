package main

import (
	"errors"
	"time"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
)

type Client struct {
	ctx     context.Context
	keysApi etcd.KeysAPI
}

func NewClient(ctx context.Context, machines []string) (*Client, error) {
	client := &Client{
		ctx: ctx,
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

func (c *Client) Register(key string, value string) error {
	if key == "" {
		return errors.New("key empty")
	}
	if value == "" {
		return errors.New("value empty")
	}
	_, err := c.keysApi.Create(c.ctx, key, value)
	return err
}

func (c *Client) Deregister(key string) error {
	if key == "" {
		return errors.New("key empty")
	}
	_, err := c.keysApi.Delete(c.ctx, key, &etcd.DeleteOptions{
		Recursive: true,
	})
	return err
}
