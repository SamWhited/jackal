/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package clusterrouter

import (
	"context"

	"github.com/ortuman/jackal/cluster"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/xmpp"
)

type clusterRouter struct {
	cluster *cluster.Cluster
}

func New(cluster *cluster.Cluster) router.ClusterRouter {
	return &clusterRouter{
		cluster: cluster,
	}
}

func (r *clusterRouter) Route(ctx context.Context, stanza xmpp.Stanza) error {
	return nil
}
