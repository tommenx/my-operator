package store

import (
	"context"
	"github.com/golang/glog"
	v3 "go.etcd.io/etcd/clientv3"
	"time"
)

var (
	DefaultTimeout = 5 * time.Second
)

type EtcdHandler struct {
	client *v3.Client
}

type DB interface {
	Put(ctx context.Context, key, val string) error
}

func NewEtcdHandler(endpoints []string) DB {
	cli, err := v3.New(v3.Config{
		Endpoints:   endpoints,
		DialTimeout: DefaultTimeout,
	})
	if err != nil {
		glog.Errorf("create ETCD client error, err=%+v", err)
		panic(err)
	}
	return &EtcdHandler{
		client: cli,
	}
}

func (h *EtcdHandler) Put(ctx context.Context, key, val string) error {
	_, err := h.client.Put(ctx, key, val)
	if err != nil {
		return nil
	}
	return nil
}
