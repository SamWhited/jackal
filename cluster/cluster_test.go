/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ortuman/jackal/cluster/etcd"
	"github.com/stretchr/testify/require"
)

func TestCluster_Shutdown(t *testing.T) {
	leader, kv := setupTest()
	cl, err := New(&Config{
		Type: Etcd,
		Etcd: &etcd.Config{Endpoints: []string{"http://127.0.0.1:2379"}},
	}, "a1")
	require.Nil(t, err)

	_ = cl.Elect()
	_ = cl.Join()
	time.Sleep(time.Second)

	require.True(t, leader.IsLeader())

	kv.mtx.Lock()
	require.Len(t, kv.m, 1)
	kv.mtx.Unlock()

	_ = cl.Shutdown(context.Background())
	time.Sleep(time.Second)

	require.False(t, leader.IsLeader())

	kv.mtx.Lock()
	require.Len(t, kv.m, 0)
	kv.mtx.Unlock()
}

func setupTest() (*mockLeader, *mockedKV) {
	var mockLeader mockLeader
	var mockKV = newMockKV()
	interfaceAddrs = func() (addrs []net.Addr, err error) {
		var ip net.IPNet
		ip.IP = net.ParseIP("10.1.0.1")
		return []net.Addr{&ip}, nil
	}
	initEtcd = func(_ *etcd.Config) (leader Leader, kv KV, err error) {
		return &mockLeader, mockKV, nil
	}
	return &mockLeader, mockKV
}
