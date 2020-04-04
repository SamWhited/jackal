/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package clusterrouter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/sony/gobreaker"

	"github.com/ortuman/jackal/cluster"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/storage"
	"github.com/ortuman/jackal/xmpp"
	"golang.org/x/net/http2"
)

const houseKeepingInterval = time.Second * 3

type clusterRouter struct {
	leader      cluster.Leader
	memberList  cluster.MemberList
	presencesSt storage.Presences
	hClient     *http.Client
	cb          *gobreaker.CircuitBreaker
}

func New(cluster *cluster.Cluster, presencesSt storage.Presences) (router.ClusterRouter, error) {
	h2cTransport := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}
	r := &clusterRouter{
		leader:      cluster,
		memberList:  cluster,
		presencesSt: presencesSt,
		hClient:     &http.Client{Transport: h2cTransport},
		cb:          gobreaker.NewCircuitBreaker(gobreaker.Settings{}),
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

func (r *clusterRouter) Route(ctx context.Context, stanza xmpp.Stanza) error {
	allocID, err := r.presencesSt.FetchPresenceAllocationID(ctx, stanza.ToJID())
	if err != nil {
		return err
	}
	if len(allocID) == 0 {
		return router.ErrNotAuthenticated
	}
	m := r.memberList.Members().Member(allocID)
	if m == nil {
		log.Warnf(fmt.Sprintf("allocation %s not found", allocID))
		return nil
	}
	// TODO(ortuman): route cluster message
	return nil
}

func (r *clusterRouter) loop() {
	tc := time.NewTicker(houseKeepingInterval)
	defer tc.Stop()

	for range tc.C {
		if err := r.houseKeeping(); err != nil {
			log.Warnf("housekeeping task error: %v", err)
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
