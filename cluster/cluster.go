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

	"github.com/ortuman/jackal/cluster/etcd"
	"github.com/ortuman/jackal/log"
)

const defaultClusterPort = "14369"

var interfaceAddrs = net.InterfaceAddrs

var initEtcd = func(config *etcd.Config) (leader Leader, kv KV, err error) { return etcd.New(config) }

// Cluster groups leader and memberlist cluster interfaces.
type Cluster struct {
	Leader
	MemberList
	srv *server
}

// New returns a new cluster subsystem instance.
func New(config *Config, allocationID string) (*Cluster, error) {
	var leader Leader
	var kv KV
	var err error

	switch config.Type {
	case Etcd:
		leader, kv, err = initEtcd(config.Etcd)
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
		Port:         defaultClusterPort,
	}
	cl := &Cluster{
		Leader:     leader,
		MemberList: newMemberList(kv, localMember),
		srv:        newServer(),
	}
	go cl.serve()

	return cl, nil
}

// RegisterStanzaHandler registers cluster stanza handler callback.
func (c *Cluster) RegisterStanzaHandler(hnd StanzaHandler) {
	c.srv.registerStanzaHandler(hnd)
}

// Shutdown shuts down cluster subsystem.
func (c *Cluster) Shutdown(ctx context.Context) error {
	ch := make(chan error)
	go func() {
		ch <- c.shutdown(ctx)
	}()
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Cluster) shutdown(ctx context.Context) error {
	if err := c.MemberList.Leave(); err != nil {
		return err
	}
	if err := c.Leader.Resign(); err != nil {
		return err
	}
	if err := c.srv.shutdown(ctx); err != nil {
		return err
	}
	log.Infof("successfully shutted down")
	return nil
}

func (c *Cluster) serve() {
	if err := c.srv.start(); err != nil {
		log.Fatal(err)
	}
}

func getLocalIP() (string, error) {
	addrs, err := interfaceAddrs()
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
