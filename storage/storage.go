/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package storage

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/storage/model"
	"github.com/ortuman/jackal/xml"
)

// ErrMockedError represents a storage mocked error value.
var ErrMockedError = errors.New("storage mocked error")

// Storage represents an entity storage interface.
type Storage interface {
	Shutdown()

	InsertOrUpdateUser(user *model.User) error
	DeleteUser(username string) error
	FetchUser(username string) (*model.User, error)
	UserExists(username string) (bool, error)

	InsertOrUpdateRosterItem(ri *model.RosterItem) (model.RosterVersion, error)
	DeleteRosterItem(username, jid string) (model.RosterVersion, error)
	FetchRosterItems(username string) ([]model.RosterItem, model.RosterVersion, error)
	FetchRosterItem(username, jid string) (*model.RosterItem, error)

	InsertOrUpdateRosterNotification(rn *model.RosterNotification) error
	DeleteRosterNotification(contact, jid string) error
	FetchRosterNotifications(contact string) ([]model.RosterNotification, error)

	InsertOrUpdateVCard(vCard xml.XElement, username string) error
	FetchVCard(username string) (xml.XElement, error)

	FetchPrivateXML(namespace string, username string) ([]xml.XElement, error)
	InsertOrUpdatePrivateXML(privateXML []xml.XElement, namespace string, username string) error

	InsertOfflineMessage(message xml.XElement, username string) error
	CountOfflineMessages(username string) (int, error)
	FetchOfflineMessages(username string) ([]xml.XElement, error)
	DeleteOfflineMessages(username string) error

	InsertOrUpdateBlockListItems(items []model.BlockListItem) error
	DeleteBlockListItems(items []model.BlockListItem) error

	FetchBlockListItems(username string) ([]model.BlockListItem, error)
}

var (
	inst        Storage
	instMu      sync.RWMutex
	initialized uint32
)

// Initialize initializes storage sub system.
func Initialize(cfg *Config) {
	if atomic.CompareAndSwapUint32(&initialized, 0, 1) {
		instMu.Lock()
		defer instMu.Unlock()

		switch cfg.Type {
		case BadgerDB:
			inst = newBadgerDB(cfg.BadgerDB)
		case MySQL:
			inst = newSQLStorage(cfg.MySQL)
		case Mock:
			inst = newMockStorage()
		default:
			// should not be reached
			break
		}
	}
}

// Instance returns global storage sub system.
func Instance() Storage {
	instMu.RLock()
	defer instMu.RUnlock()

	if inst == nil {
		log.Fatalf("storage subsystem not initialized")
	}
	return inst
}

// Shutdown shuts down storage sub system.
// This method should be used only for testing purposes.
func Shutdown() {
	if atomic.CompareAndSwapUint32(&initialized, 1, 0) {
		instMu.Lock()
		defer instMu.Unlock()

		inst.Shutdown()
		inst = nil
	}
}

// ActivateMockedError forces the return of ErrMockedError from current storage manager.
// This method should only be used for testing purposes.
func ActivateMockedError() {
	instMu.Lock()
	defer instMu.Unlock()

	switch inst := inst.(type) {
	case *mockStorage:
		inst.activateMockedError()
	}
}

// DeactivateMockedError disables mocked storage error from a previous activation.
// This method should only be used for testing purposes.
func DeactivateMockedError() {
	instMu.Lock()
	defer instMu.Unlock()

	switch inst := inst.(type) {
	case *mockStorage:
		inst.deactivateMockedError()
	}
}
