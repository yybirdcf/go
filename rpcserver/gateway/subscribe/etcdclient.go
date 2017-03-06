package subscribe

import (
	"errors"
	"time"

	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
)

type EtcdClient struct {
	ctx     context.Context
	keysApi etcd.KeysAPI
}

func NewEtcdClient(ctx context.Context, machines []string) (*EtcdClient, error) {
	etcdClient := &EtcdClient{
		ctx: ctx,
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

	etcdClient.keysApi = etcd.NewKeysAPI(client)

	return etcdClient, nil
}

func (c *EtcdClient) Register(key string, value string) error {
	if key == "" {
		return errors.New("key empty")
	}
	if value == "" {
		return errors.New("value empty")
	}
	_, err := c.keysApi.Create(c.ctx, key, value)
	return err
}

func (c *EtcdClient) Deregister(key string) error {
	if key == "" {
		return errors.New("key empty")
	}
	_, err := c.keysApi.Delete(c.ctx, key, &etcd.DeleteOptions{
		Recursive: true,
	})
	return err
}
