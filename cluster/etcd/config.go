/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package etcd

type Config struct {
	Endpoints []string `yaml:"endpoints"`
}
