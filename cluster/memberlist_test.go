/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemberList(t *testing.T) {
	defaultAliveTTL = time.Second

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
	return newMemberList(kv, Member{AllocationID: "a1", Host: "10.1.0.2", Port: "2786"}), kv
}
