/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/ortuman/jackal/cluster/etcd"
	"github.com/ortuman/jackal/log"
)

type Cluster struct {
	Candidate
	MemberList
}

func New(config *Config, allocationID string) (*Cluster, error) {
	var candidate Candidate
	var kv KV
	var err error

	switch config.Type {
	case Etcd:
		candidate, kv, err = etcd.New(config.Etcd)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("cluster: unrecognized cluster type: %d", config.Type)
	}
	localIP, err := getLocalIP()
	if err != nil {
		return nil, err
	}
	localMember := Member{
		AllocationID: allocationID,
		Host:         localIP,
		Port:         strconv.Itoa(config.Port),
	}
	return &Cluster{
		Candidate:  candidate,
		MemberList: newMemberList(kv, localMember, config.AliveTTL),
	}, nil
}

func (c *Cluster) Shutdown(ctx context.Context) error {
	ch := make(chan error)
	go func() {
		ch <- c.shutdown()
	}()
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Cluster) shutdown() error {
	if err := c.MemberList.Leave(); err != nil {
		return err
	}
	if err := c.Candidate.Resign(); err != nil {
		return err
	}
	log.Infof("cluster: shutted down successfully")
	return nil
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("failed to get local ip")
}
