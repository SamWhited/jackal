/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package etcd

import (
	"context"
	"log"
	"sync/atomic"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

type Candidate struct {
	allocID  string
	ss       *concurrency.Session
	election *concurrency.Election
	isLeader int32
}

func newCandidate(cli *v3.Client, allocationID string) (*Candidate, error) {
	c := &Candidate{allocID: allocationID}
	ss, err := concurrency.NewSession(cli)
	if err != nil {
		return nil, err
	}
	c.ss = ss
	c.election = concurrency.NewElection(c.ss, "/jackal-leader/")
	return c, nil
}

func (c *Candidate) Elect() error {
	go c.elect()
	return nil
}

func (c *Candidate) Resign() error {
	return c.election.Resign(context.Background())
}

func (c *Candidate) IsLeader() bool {
	return atomic.LoadInt32(&c.isLeader) == 1
}

func (c *Candidate) Close() error {
	return c.ss.Close()
}

func (c *Candidate) elect() {
	if err := c.election.Campaign(context.Background(), c.allocID); err != nil {
		log.Fatalf("%v", err)
		return
	}
	atomic.StoreInt32(&c.isLeader, 1)
}
