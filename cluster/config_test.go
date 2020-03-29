/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestClusterConfig(t *testing.T) {
	cfg := Config{}

	etcdCfg := `
  type: etcd
  etcd:
    endpoints: [http://127.0.0.1:2378, http://127.0.0.1:2379]
  port: 7898
`

	err := yaml.Unmarshal([]byte(etcdCfg), &cfg)
	require.Nil(t, err)
	require.Equal(t, Etcd, cfg.Type)
	require.NotNil(t, cfg.Etcd)
	require.Len(t, cfg.Etcd.Endpoints, 2)

	invalidMySQLCfg := `
  type: etcd
`
	err = yaml.Unmarshal([]byte(invalidMySQLCfg), &cfg)
	require.NotNil(t, err)

	invalidCfg := `
  type: invalid
`
	err = yaml.Unmarshal([]byte(invalidCfg), &cfg)
	require.NotNil(t, err)
}

func TestClusterBadConfig(t *testing.T) {
	cfg := Config{}

	memCfg := `
  type
`
	err := yaml.Unmarshal([]byte(memCfg), &cfg)
	require.NotNil(t, err)
}
