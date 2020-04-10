/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package c2srouter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ortuman/jackal/cluster"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/storage"
	"github.com/ortuman/jackal/util/pool"
	"github.com/ortuman/jackal/xmpp"
	"github.com/sony/gobreaker"
	"golang.org/x/net/http2"
)

const houseKeepingInterval = time.Second * 3

type clusterRouter struct {
	hClient     *http.Client
	cb          *gobreaker.CircuitBreaker
	pool        *pool.BufferPool
	leader      cluster.Leader
	memberList  cluster.MemberList
	presencesSt storage.Presences
	closeCh     chan chan struct{}
}

func newClusterRouter(cluster *cluster.Cluster, presencesSt storage.Presences) (*clusterRouter, error) {
	h2cTransport := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}
	r := &clusterRouter{
		hClient:     &http.Client{Transport: h2cTransport},
		cb:          gobreaker.NewCircuitBreaker(gobreaker.Settings{}),
		pool:        pool.NewBufferPool(),
		leader:      cluster.Leader,
		memberList:  cluster.MemberList,
		presencesSt: presencesSt,
		closeCh:     make(chan chan struct{}, 1),
	}
	if err := r.leader.Elect(); err != nil {
		return nil, err
	}
	if err := r.memberList.Join(); err != nil {
		return nil, err
	}
	go r.loop()

	return r, nil
}

func (r *clusterRouter) route(ctx context.Context, stanza xmpp.Stanza, allocationID string) error {
	member := r.memberList.Members().Member(allocationID)
	if member == nil {
		log.Warnf(fmt.Sprintf("allocation %s not found", allocationID))
		return nil
	}
	buf := r.pool.Get()
	defer r.pool.Put(buf)

	if err := stanza.ToXML(buf, true); err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s:%s/route", member.Host, member.Port)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/xml")

	_, err = r.cb.Execute(func() (i interface{}, e error) {
		resp, err := r.hClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("response status code: %d", resp.StatusCode)
		}
		return nil, nil
	})
	return err
}

func (r *clusterRouter) shutdown(ctx context.Context) error {
	ch := make(chan struct{})
	r.closeCh <- ch
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *clusterRouter) loop() {
	tc := time.NewTicker(houseKeepingInterval)
	defer tc.Stop()

	for {
		select {
		case <-tc.C:
			if err := r.houseKeeping(); err != nil {
				log.Warnf("housekeeping task error: %v", err)
			}
		case ch := <-r.closeCh:
			close(ch)
			return
		}
	}
}

func (r *clusterRouter) houseKeeping() error {
	if !r.leader.IsLeader() {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), (houseKeepingInterval*5)/10)
	defer cancel()

	allocIDs, err := r.presencesSt.FetchAllocationIDs(ctx)
	if err != nil {
		return err
	}
	members := r.memberList.Members()
	for _, allocID := range allocIDs {
		if m := members.Member(allocID); m != nil {
			continue
		}
		// clear inactive allocation presences
		if err := r.presencesSt.DeleteAllocationPresences(ctx, allocID); err != nil {
			return err
		}
	}
	return nil
}
