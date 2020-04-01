/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import "sync/atomic"

type mockLeader struct{ isLeader int32 }

func (m *mockLeader) IsLeader() bool {
	return atomic.LoadInt32(&m.isLeader) == 1
}

func (m *mockLeader) Elect() error {
	atomic.StoreInt32(&m.isLeader, 1)
	return nil
}

func (m *mockLeader) Resign() error {
	atomic.StoreInt32(&m.isLeader, 0)
	return nil
}
