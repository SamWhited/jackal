/*
 * Copyright (c) 2019 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package storage

import (
	"context"
	"fmt"

	memorystorage "github.com/ortuman/jackal/storage/memory"
	"github.com/ortuman/jackal/storage/mysql"
	"github.com/ortuman/jackal/storage/pgsql"
)

type Storage struct {
	User
	Roster
	Presences
	VCard
	Private
	BlockList
	PubSub
	Offline

	shutdownFn            func(ctx context.Context) error
	isClusterCompatibleFn func() bool
}

// Shutdown
func (s *Storage) Shutdown(ctx context.Context) error { return s.shutdownFn(ctx) }

// IsClusterCompatible tells whether undelying storage is cluster compatible.
func (s *Storage) IsClusterCompatible() bool { return s.isClusterCompatibleFn() }

// New initializes configured storage type.
func New(config *Config) (*Storage, error) {
	switch config.Type {
	case MySQL:
		return initMySQL(config.MySQL)
	case PostgreSQL:
		return initPgSQL(config.PostgreSQL)
	case Memory:
		return initMemoryStorage()
	default:
		return nil, fmt.Errorf("storage: unrecognized storage type: %d", config.Type)
	}
}

func initMySQL(config *mysql.Config) (*Storage, error) {
	mySQLStorage, err := mysql.New(config)
	if err != nil {
		return nil, err
	}
	return &Storage{
		User:                  mySQLStorage.User,
		Roster:                mySQLStorage.Roster,
		Presences:             mySQLStorage.Presences,
		VCard:                 mySQLStorage.VCard,
		Private:               mySQLStorage.Private,
		BlockList:             mySQLStorage.BlockList,
		PubSub:                mySQLStorage.PubSub,
		Offline:               mySQLStorage.Offline,
		shutdownFn:            mySQLStorage.Shutdown,
		isClusterCompatibleFn: mySQLStorage.IsClusterCompatible,
	}, nil
}

func initPgSQL(config *pgsql.Config) (*Storage, error) {
	pgSQLStorage, err := pgsql.New(config)
	if err != nil {
		return nil, err
	}
	return &Storage{
		User:                  pgSQLStorage.User,
		Roster:                pgSQLStorage.Roster,
		Presences:             pgSQLStorage.Presences,
		VCard:                 pgSQLStorage.VCard,
		Private:               pgSQLStorage.Private,
		BlockList:             pgSQLStorage.BlockList,
		PubSub:                pgSQLStorage.PubSub,
		Offline:               pgSQLStorage.Offline,
		shutdownFn:            pgSQLStorage.Shutdown,
		isClusterCompatibleFn: pgSQLStorage.IsClusterCompatible,
	}, nil
}

func initMemoryStorage() (*Storage, error) {
	memStorage, err := memorystorage.New()
	if err != nil {
		return nil, err
	}
	return &Storage{
		User:                  memStorage.User,
		Roster:                memStorage.Roster,
		Presences:             memStorage.Presences,
		VCard:                 memStorage.VCard,
		Private:               memStorage.Private,
		BlockList:             memStorage.BlockList,
		PubSub:                memStorage.PubSub,
		Offline:               memStorage.Offline,
		shutdownFn:            memStorage.Shutdown,
		isClusterCompatibleFn: memStorage.IsClusterCompatible,
	}, nil
}
