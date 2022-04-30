package services

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"reflect"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/security"
	"github.com/bhojpur/belaur/pkg/store"
	"github.com/bhojpur/belaur/pkg/store/memdb"
)

// storeService is an instance of store.
// Use this to talk to the store.
var storeService store.BelaurStore

// vaultService is an instance of the internal Vault.
var vaultService security.BelaurVault

// memDBService is an instance of the internal memdb.
var memDBService memdb.BelaurMemDB

// StorageService initializes and keeps track of a storage service.
// If the internal storage service is a singleton. This function retruns an error
// but most of the times we don't care about it, because it's only ever
// initialized once in the main.go. If it wouldn't work, main would
// os.Exit(1) and the rest of the application would just stop.
func StorageService() (store.BelaurStore, error) {
	if storeService != nil && !reflect.ValueOf(storeService).IsNil() {
		return storeService, nil
	}
	storeService = store.NewBoltStore()
	err := storeService.Init(belaur.Cfg.DataPath)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot initialize store", "error", err.Error())
		return storeService, err
	}
	return storeService, nil
}

// MockStorageService sets the internal store singleton to the give
// mock implementation. A mock needs to be created in the test. The
// provider will make sure that everything that would use the store
// will use the mock instead.
func MockStorageService(store store.BelaurStore) {
	storeService = store
}

// DefaultVaultService provides a vault with a FileStorer backend.
func DefaultVaultService() (security.BelaurVault, error) {
	return VaultService(&security.FileVaultStorer{})
}

// VaultService creates a vault manager service.
func VaultService(vaultStore security.VaultStorer) (security.BelaurVault, error) {
	if vaultService != nil && !reflect.ValueOf(vaultService).IsNil() {
		return vaultService, nil
	}

	// TODO: For now use this to keep the refactor of certificate out of the refactor of Vault Service.
	ca, err := security.InitCA()
	if err != nil {
		belaur.Cfg.Logger.Error("cannot initialize certificate:", "error", err.Error())
		return nil, err
	}
	v, err := security.NewVault(ca, vaultStore)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot initialize vault manager:", "error", err.Error())
		return nil, err
	}
	vaultService = v
	return vaultService, nil
}

// MockVaultService provides a way to create and set a mock
// for the internal vault service manager.
func MockVaultService(service security.BelaurVault) {
	vaultService = service
}

// DefaultMemDBService provides a default memDBService with an underlying storer.
func DefaultMemDBService() (memdb.BelaurMemDB, error) {
	s, err := StorageService()
	if err != nil {
		return nil, err
	}
	return MemDBService(s)
}

// MemDBService creates a memdb service instance.
func MemDBService(store store.BelaurStore) (memdb.BelaurMemDB, error) {
	if memDBService != nil && !reflect.ValueOf(memDBService).IsNil() {
		return memDBService, nil
	}

	db, err := memdb.InitMemDB(store)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot initialize memdb service", "error", err.Error())
		return nil, err
	}
	memDBService = db
	return memDBService, nil
}

// MockMemDBService provides a way to create and set a mock
// for the internal memdb service manager.
func MockMemDBService(db memdb.BelaurMemDB) {
	memDBService = db
}
