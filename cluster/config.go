/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"errors"
	"fmt"

	"github.com/ortuman/jackal/cluster/etcd"
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
	Type Type
	Etcd *etcd.Config
}

type clusterProxyType struct {
	Type string       `yaml:"type"`
	Etcd *etcd.Config `yaml:"etcd"`
}

// UnmarshalYAML satisfies Unmarshaler interface.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	p := clusterProxyType{}

	if err := unmarshal(&p); err != nil {
		return err
	}
	switch p.Type {
	case "etcd":
		if p.Etcd == nil {
			return errors.New("cluster.Config: couldn't read etcd configuration")
		}
		c.Type = Etcd
		c.Etcd = p.Etcd

	case "":
		return errors.New("cluster.Config: unspecified storage type")

	default:
		return fmt.Errorf("cluster.Config: unrecognized storage type: %s", p.Type)
	}
	return nil
}
