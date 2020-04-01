/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package clusterrouter

import (
	"context"
	"time"

	"github.com/ortuman/jackal/cluster"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/storage"
	"github.com/ortuman/jackal/xmpp"
)

const houseKeepingInterval = time.Second * 3

type clusterRouter struct {
	leader      cluster.Leader
	memberList  cluster.MemberList
	presencesSt storage.Presences
}

func New(cluster *cluster.Cluster, presencesSt storage.Presences) (router.ClusterRouter, error) {
	r := &clusterRouter{
		leader:      cluster,
		memberList:  cluster,
		presencesSt: presencesSt,
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
