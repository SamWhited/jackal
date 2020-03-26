/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

type Candidate interface {
	IsLeader() bool

	Elect() error
	Resign() error

	Close() error
}
