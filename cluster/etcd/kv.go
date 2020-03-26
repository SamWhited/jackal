/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package etcd

import (
	"context"

	v3 "github.com/coreos/etcd/clientv3"
)

type KV struct {
	cli *v3.Client
}

func (k *KV) Put(ctx context.Context, key string, value string, ttlInSeconds int64) error {
	resp, err := k.cli.Grant(ctx, ttlInSeconds)
	if err != nil {
		return err
	}
	_, err = k.cli.Put(ctx, key, value, v3.WithLease(resp.ID))
	return err
}

func (k *KV) Get(ctx context.Context, key string) (string, error) {
	resp, err := k.cli.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return string(resp.Kvs[0].Value), nil
}

func (k *KV) GetPrefix(ctx context.Context, prefix string) ([]string, error) {
	resp, err := k.cli.Get(ctx, prefix, v3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var values []string
	for _, kv := range resp.Kvs {
		values = append(values, string(kv.Value))
	}
	return values, nil
}
