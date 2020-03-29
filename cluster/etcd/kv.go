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

func (k *KV) Del(ctx context.Context, key string) error {
	_, err := k.cli.Delete(ctx, key)
	return err
}

func (k *KV) GetPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	resp, err := k.cli.Get(ctx, prefix, v3.WithPrefix())
	if err != nil {
		return nil, err
	}
	values := make(map[string]string, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		values[string(kv.Key)] = string(kv.Value)
	}
	return values, nil
}
