/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"errors"
	"fmt"
	"time"

	"github.com/ortuman/jackal/cluster/etcd"
)

const (
	defaultAliveTTL = time.Second * 3
)

// Type represents a cluster type.
type Type int

const (
	// Etcd represents a Etcd cluster type.
	Etcd Type = iota
)

var typeStringMap = map[Type]string{
	Etcd: "etcd",
}

func (t Type) String() string { return typeStringMap[t] }

// Config represents a cluster configuration.
type Config struct {
	Type     Type
	Etcd     *etcd.Config
	Port     int
	AliveTTL time.Duration
}

type clusterProxyType struct {
	Type     string       `yaml:"type"`
	Etcd     *etcd.Config `yaml:"etcd"`
	Port     int          `yaml:"port"`
	AliveTTL int          `yaml:"alive_ttl"`
}

// UnmarshalYAML satisfies Unmarshaler interface.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	p := clusterProxyType{}

	if err := unmarshal(&p); err != nil {
		return err
	}
	if p.Port == 0 {
		return errors.New("cluster.Config: port zero value")
	}

	switch p.Type {
	case "etcd":
		if p.Etcd == nil {
			return errors.New("cluster.Config: couldn't read etcd configuration")
		}
		c.Type = Etcd
		c.Etcd = p.Etcd
		c.Port = p.Port

		c.AliveTTL = time.Duration(p.AliveTTL) * time.Second
		if c.AliveTTL == 0 {
			c.AliveTTL = defaultAliveTTL
		}

	case "":
		return errors.New("cluster.Config: unspecified storage type")

	default:
		return fmt.Errorf("cluster.Config: unrecognized storage type: %s", p.Type)
	}
	return nil
}
