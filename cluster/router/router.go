/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package clusterrouter

import (
	"context"
	"time"

	"github.com/ortuman/jackal/cluster"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/xmpp"
)

const houseKeepingInterval = time.Second * 3

type clusterRouter struct {
	leader     cluster.Leader
	memberList cluster.MemberList
}

func New(cluster *cluster.Cluster) (router.ClusterRouter, error) {
	r := &clusterRouter{
		leader:     cluster,
		memberList: cluster,
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
	return nil
}

func (r *clusterRouter) loop() {
	tc := time.NewTicker(houseKeepingInterval)
	defer tc.Stop()

	for range tc.C {
		r.houseKeeping()
	}
}

func (r *clusterRouter) houseKeeping() {
	if !r.leader.IsLeader() {
		return
	}
}
