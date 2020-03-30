/*
 * Copyright (c) 2019 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package memorystorage

import (
	"context"

	"github.com/ortuman/jackal/log"
)

type Storage struct {
	User      *User
	Roster    *Roster
	Presences *Presences
	VCard     *VCard
	Private   *Private
	BlockList *BlockList
	PubSub    *PubSub
	Offline   *Offline
}

// New initializes in-memory storage and returns associated container.
func New() (*Storage, error) {
	return &Storage{
		User:      NewUser(),
		Roster:    NewRoster(),
		Presences: NewPresences(),
		VCard:     NewVCard(),
		Private:   NewPrivate(),
		BlockList: NewBlockList(),
		PubSub:    NewPubSub(),
		Offline:   NewOffline(),
	}, nil
}

func (c *Storage) Shutdown(_ context.Context) error {
	log.Infof("memorystorage: shutted down successfully")
	return nil
}

func (c *Storage) IsClusterCompatible() bool { return false }
