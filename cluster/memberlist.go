/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ortuman/jackal/log"
)

const (
	memberListOpTimeout = time.Second * 3

	memberListPrefix = "jackal/memberlist"
)

type MemberList interface {
	Join(localMember Member) error
	Leave() error

	Members() Members
}

type memberList struct {
	kv       KV
	aliveTTL time.Duration
	closeCh  chan chan bool
	joined   int32
	mu       sync.RWMutex
	members  Members
}

func newMemberList(kv KV, aliveTTL time.Duration) *memberList {
	return &memberList{
		kv:       kv,
		aliveTTL: aliveTTL,
		closeCh:  make(chan chan bool),
	}
}

func (m *memberList) Join(localMember Member) error {
	if !atomic.CompareAndSwapInt32(&m.joined, 0, 1) {
		return nil // already joined
	}
	ctx, cancel := context.WithTimeout(context.Background(), memberListOpTimeout)
	defer cancel()

	if err := m.refreshMembers(ctx, localMember); err != nil {
		atomic.StoreInt32(&m.joined, 0)
		return err
	}
	go m.loop(localMember)
	return nil
}

func (m *memberList) Leave() error {
	if atomic.LoadInt32(&m.joined) == 0 {
		return nil
	}
	ch := make(chan bool)
	m.closeCh <- ch
	<-ch
	return nil
}

func (m *memberList) Members() Members {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.members
}

func (m *memberList) loop(localMember Member) {
	tc := time.NewTicker((m.aliveTTL * 5) / 10)
	defer tc.Stop()

	for {
		select {
		case <-tc.C:
			ctx, cancel := context.WithTimeout(context.Background(), memberListOpTimeout)
			if err := m.refreshMembers(ctx, localMember); err != nil {
				log.Warnf("cluster: failed to refresh member list: %v", err)
			}
			cancel()

		case ch := <-m.closeCh:
			close(ch)
			return
		}
	}
}

func (m *memberList) refreshMembers(ctx context.Context, localMember Member) error {
	// refresh local member
	if err := m.putMember(ctx, localMember, int64(m.aliveTTL/time.Second)); err != nil {
		return err
	}
	// update member list
	members, err := m.getMembers(ctx)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.members = members
	m.mu.Unlock()
	return nil
}

func (m *memberList) putMember(ctx context.Context, member Member, ttlInSeconds int64) error {
	key := fmt.Sprintf("%s/%s", memberListPrefix, member.AllocationID)
	val := net.JoinHostPort(member.Host, member.Port)
	return m.kv.Put(ctx, key, val, ttlInSeconds)
}

func (m *memberList) getMembers(ctx context.Context) (Members, error) {
	kvs, err := m.kv.GetPrefix(ctx, memberListPrefix)
	if err != nil {
		return nil, err
	}
	var members Members
	for k, v := range kvs {
		var m Member

		host, port, err := net.SplitHostPort(v)
		if err != nil {
			return nil, err
		}
		m.AllocationID = k
		m.Host = host
		m.Port = port
		members = append(members, m)
	}
	return members, nil
}
