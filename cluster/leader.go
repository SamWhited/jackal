/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

import _ "go.etcd.io/etcd/clientv3"

type Candidate interface {
	CheckLeadership() (isLeader bool, err error)

	Elect() error
	Resign() error
}
