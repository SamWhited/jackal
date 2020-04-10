/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package c2srouter

import (
	"context"
	"sync"

	"github.com/ortuman/jackal/cluster"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/storage"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

type c2sRouter struct {
	mu            sync.RWMutex
	tbl           map[string]*resources
	userSt        storage.User
	blockListSt   storage.BlockList
	presencesSt   storage.Presences
	cluster       *cluster.Cluster
	clusterRouter *clusterRouter
}

func New(userSt storage.User, blockListSt storage.BlockList, presencesSt storage.Presences, cluster *cluster.Cluster) (router.C2SRouter, error) {
	r := &c2sRouter{
		tbl:         make(map[string]*resources),
		userSt:      userSt,
		blockListSt: blockListSt,
		presencesSt: presencesSt,
	}
	if cluster != nil {
		clusterRouter, err := newClusterRouter(cluster, presencesSt)
		if err != nil {
			return nil, err
		}
		r.cluster = cluster
		r.clusterRouter = clusterRouter
	}
	return r, nil
}

func (r *c2sRouter) Route(ctx context.Context, stanza xmpp.Stanza, validations router.C2SRoutingValidations) error {
	fromJID := stanza.FromJID()
	toJID := stanza.ToJID()

	// apply validations
	username := stanza.ToJID().Node()
	if (validations & router.UserExistence) > 0 {
		exists, err := r.userSt.UserExists(ctx, username) // user exists?
		if err != nil {
			return err
		}
		if !exists {
			return router.ErrNotExistingAccount
		}
	}
	if (validations & router.BlockedDestinationJID) > 0 {
		if r.isBlockedJID(ctx, toJID, fromJID.Node()) { // check whether destination JID is blocked
			return router.ErrBlockedJID
		}
	}
	return r.route(ctx, stanza)
}

func (r *c2sRouter) Bind(stm stream.C2S) {
	user := stm.Username()
	r.mu.RLock()
	rs := r.tbl[user]
	r.mu.RUnlock()

	if rs == nil {
		r.mu.Lock()
		rs = r.tbl[user] // avoid double initialization
		if rs == nil {
			rs = &resources{}
			r.tbl[user] = rs
		}
		r.mu.Unlock()
	}
	rs.bind(stm)

	log.Infof("bound c2s stream... (%s/%s)", stm.Username(), stm.Resource())
}

func (r *c2sRouter) Unbind(user, resource string) {
	r.mu.RLock()
	rs := r.tbl[user]
	r.mu.RUnlock()

	if rs == nil {
		return
	}
	r.mu.Lock()
	rs.unbind(resource)
	if rs.len() == 0 {
		delete(r.tbl, user)
	}
	r.mu.Unlock()

	log.Infof("unbound c2s stream... (%s/%s)", user, resource)
}

func (r *c2sRouter) Stream(username, resource string) stream.C2S {
	r.mu.RLock()
	rs := r.tbl[username]
	r.mu.RUnlock()

	if rs == nil {
		return nil
	}
	return rs.stream(resource)
}

func (r *c2sRouter) Streams(username string) []stream.C2S {
	r.mu.RLock()
	rs := r.tbl[username]
	r.mu.RUnlock()

	if rs == nil {
		return nil
	}
	return rs.allStreams()
}

func (r *c2sRouter) route(ctx context.Context, stanza xmpp.Stanza) error {
	toJID := stanza.ToJID()
	if toJID.IsFullWithUser() {
		return r.routeToResource(ctx, stanza)
	}
	switch stanza.(type) {
	case *xmpp.Message:
		routed, err := r.routeToPrioritaryResource(ctx, stanza)
		if err != nil {
			return err
		}
		if !routed {
			goto route2all
		}
		return nil
	}
route2all:
	return r.routeToAllResources(ctx, stanza)
}

func (r *c2sRouter) routeToResource(ctx context.Context, stanza xmpp.Stanza) error {
	allocID, err := r.presencesSt.FetchPresenceAllocationID(ctx, stanza.ToJID())
	if err != nil {
		return err
	}
	if len(allocID) == 0 {
		return router.ErrResourceNotFound
	}
	if r.clusterRouter == nil || r.cluster.IsLocalAllocationID(allocID) {

	}
	return r.clusterRouter.route(ctx, stanza, allocID)
}

func (r *c2sRouter) routeToPrioritaryResource(ctx context.Context, stanza xmpp.Stanza) (routed bool, err error) {
	return false, nil
}

func (r *c2sRouter) routeToAllResources(ctx context.Context, stanza xmpp.Stanza) error {
	return nil
}

func (r *c2sRouter) isBlockedJID(ctx context.Context, j *jid.JID, username string) bool {
	blockList, err := r.blockListSt.FetchBlockListItems(ctx, username)
	if err != nil {
		log.Error(err)
		return false
	}
	if len(blockList) == 0 {
		return false
	}
	blockListJIDs := make([]jid.JID, len(blockList))
	for i, listItem := range blockList {
		j, _ := jid.NewWithString(listItem.JID, true)
		blockListJIDs[i] = *j
	}
	for _, blockedJID := range blockListJIDs {
		if blockedJID.Matches(j) {
			return true
		}
	}
	return false
}
