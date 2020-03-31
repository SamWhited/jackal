/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ortuman/jackal/log"
)

const (
	memberListOpTimeout = time.Second * 3

	memberListPrefix = "jackal/memberlist/"
)

type MemberList interface {
	Join() error
	Leave() error

	Members() Members
}

type memberList struct {
	kv          KV
	localMember Member
	aliveTTL    time.Duration
	closeCh     chan chan bool
	joined      int32
	left        int32
	mu          sync.RWMutex
	members     Members
}

func newMemberList(kv KV, localMember Member, aliveTTL time.Duration) *memberList {
	return &memberList{
		kv:          kv,
		localMember: localMember,
		aliveTTL:    aliveTTL,
		closeCh:     make(chan chan bool),
	}
}

func (m *memberList) Join() error {
	if !atomic.CompareAndSwapInt32(&m.joined, 0, 1) {
		return nil // already joined
	}
	ctx, cancel := context.WithTimeout(context.Background(), memberListOpTimeout)
	defer cancel()

	if err := m.refreshMembers(ctx); err != nil {
		atomic.StoreInt32(&m.joined, 0)
		return err
	}
	go m.loop()
	return nil
}

func (m *memberList) Leave() error {
	if atomic.LoadInt32(&m.joined) == 0 {
		return nil
	}
	if !atomic.CompareAndSwapInt32(&m.left, 0, 1) {
		return nil // already left
	}
	ctx, cancel := context.WithTimeout(context.Background(), memberListOpTimeout)
	defer cancel()

	if err := m.delMember(ctx, m.localMember.AllocationID); err != nil {
		return err
	}
	m.mu.Lock()
	m.members = nil
	m.mu.Unlock()

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

func (m *memberList) loop() {
	tc := time.NewTicker((m.aliveTTL * 5) / 10)
	defer tc.Stop()

	for {
		select {
		case <-tc.C:
			log.Debugf("refreshing member list...")

			ctx, cancel := context.WithTimeout(context.Background(), memberListOpTimeout)
			if err := m.refreshMembers(ctx); err != nil {
				log.Warnf("failed to refresh member list: %v", err)
			}
			cancel()

		case ch := <-m.closeCh:
			close(ch)
			return
		}
	}
}

func (m *memberList) refreshMembers(ctx context.Context) error {
	// refresh local member
	if err := m.putMember(ctx, m.localMember, int64(m.aliveTTL/time.Second)); err != nil {
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
	key := fmt.Sprintf("%s%s", memberListPrefix, member.AllocationID)
	val := net.JoinHostPort(member.Host, member.Port)
	return m.kv.Put(ctx, key, val, ttlInSeconds)
}

func (m *memberList) delMember(ctx context.Context, allocationID string) error {
	key := fmt.Sprintf("%s%s", memberListPrefix, allocationID)
	return m.kv.Del(ctx, key)
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
		m.AllocationID = strings.Replace(k, memberListPrefix, "", 1) // remove key prefix
		m.Host = host
		m.Port = port
		members = append(members, m)
	}
	return members, nil
}
