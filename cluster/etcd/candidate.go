/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package etcd

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/pborman/uuid"
)

const candidateOpTimeout = time.Second * 3

type Candidate struct {
	ss       *concurrency.Session
	election *concurrency.Election
	isLeader int32
	elected  int32
}

func newCandidate(cli *v3.Client) (*Candidate, error) {
	c := &Candidate{}
	ss, err := concurrency.NewSession(cli)
	if err != nil {
		return nil, err
	}
	c.ss = ss
	c.election = concurrency.NewElection(c.ss, "/jackal-leader/")
	return c, nil
}

func (c *Candidate) Elect() error {
	if atomic.CompareAndSwapInt32(&c.elected, 0, 1) {
		return nil // already electe
	}
	go c.elect()
	return nil
}

func (c *Candidate) Resign() error {
	if atomic.LoadInt32(&c.elected) == 0 {
		return nil // nothing to do here
	}
	ctx, cancel := context.WithTimeout(context.Background(), candidateOpTimeout)
	defer cancel()

	if err := c.election.Resign(ctx); err != nil {
		return err
	}
	return c.ss.Close()
}

func (c *Candidate) IsLeader() bool {
	return atomic.LoadInt32(&c.isLeader) == 1
}

func (c *Candidate) elect() {
	val := uuid.New()
	if err := c.election.Campaign(context.Background(), val); err != nil {
		log.Fatalf("%v", err)
		return
	}
	atomic.StoreInt32(&c.isLeader, 1)
}
