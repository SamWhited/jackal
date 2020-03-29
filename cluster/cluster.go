/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/ortuman/jackal/cluster/etcd"
)

type Cluster struct {
	Candidate
	MemberList

	localMember Member
}

func New(config *Config) (*Cluster, error) {
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
	return &Cluster{
		localMember: Member{Host: localIP, Port: strconv.Itoa(config.Port)},
		Candidate:   candidate,
		MemberList:  newMemberList(kv, config.AliveTTL),
	}, nil
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
