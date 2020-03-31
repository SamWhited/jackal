/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockedKV struct {
	putInvCount int32
	mtx         sync.RWMutex
	m           map[string]string
}

func newMockKV() *mockedKV {
	return &mockedKV{m: make(map[string]string)}
}

func (kv *mockedKV) Put(_ context.Context, key string, value string, ttlInSeconds int64) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()
	kv.m[key] = value
	atomic.AddInt32(&kv.putInvCount, 1)
	return nil
}

func (kv *mockedKV) Del(_ context.Context, key string) error {
	kv.mtx.Lock()
	defer kv.mtx.Unlock()
	delete(kv.m, key)
	return nil
}

func (kv *mockedKV) GetPrefix(_ context.Context, prefix string) (map[string]string, error) {
	kv.mtx.RLock()
	defer kv.mtx.RUnlock()
	res := make(map[string]string)
	for k, v := range kv.m {
		if strings.HasPrefix(k, prefix) {
			res[k] = v
		}
	}
	return res, nil
}

func TestMemberList(t *testing.T) {
	ml, kv := testMemberList()

	err := ml.Join()
	require.Nil(t, err)

	ms := ml.Members()
	require.Len(t, ms, 1)

	require.Equal(t, "10.1.0.2", ms[0].Host)
	require.Equal(t, "2786", ms[0].Port)

	time.Sleep(time.Millisecond * 750) // wait refresh

	require.Equal(t, int32(2), atomic.LoadInt32(&kv.putInvCount))

	err = ml.Leave()
	require.Nil(t, err)

	ms = ml.Members()
	require.Len(t, ms, 0)
}

func testMemberList() (*memberList, *mockedKV) {
	kv := newMockKV()
	return newMemberList(kv, Member{AllocationID: "a1", Host: "10.1.0.2", Port: "2786"}, time.Second), kv
}
