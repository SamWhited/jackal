/*
 * Copyright (c) 2020 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package cluster

// Leader defines cluster leader interface.
type Leader interface {
	// IsLeader tells whether current allocation was elected as leader
	IsLeader() bool

	// Elect starts cluster leader election
	Elect() error

	// Resign resigns current allocation as leader in case it was previously elected.
	Resign() error
}
