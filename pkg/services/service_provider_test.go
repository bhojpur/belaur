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
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/security"
	"github.com/bhojpur/belaur/pkg/store"
	"github.com/bhojpur/belaur/pkg/store/memdb"
	"github.com/hashicorp/go-hclog"
)

func TestStorageService(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestStorageService")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	belaur.Cfg.DataPath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	if storeService != nil {
		t.Fatal("initial service should be nil. was: ", storeService)
	}
	_, _ = StorageService()
	defer func() { storeService = nil }()
	if storeService == nil {
		t.Fatal("storage service should not be nil")
	}
}

func TestVaultService(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestVaultService")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.CAPath = tmp
	belaur.Cfg.VaultPath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	if vaultService != nil {
		t.Fatal("initial service should be nil. was: ", vaultService)
	}
	_, _ = DefaultVaultService()
	defer func() {
		vaultService = nil
	}()

	if vaultService == nil {
		t.Fatal("service should not be nil")
	}
}

func TestMemDBService(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestMemDBService")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.CAPath = tmp
	belaur.Cfg.VaultPath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	if memDBService != nil {
		t.Fatal("initial service should be nil. was: ", memDBService)
	}
	if _, err := StorageService(); err != nil {
		t.Fatal(err)
	}
	if _, err := MemDBService(storeService); err != nil {
		t.Fatal(err)
	}
	defer func() {
		memDBService = nil
		storeService = nil
	}()

	if memDBService == nil {
		t.Fatal("service should not be nil")
	}
}

type testMockStorageService struct {
	store.BelaurStore
}

type testMockVaultService struct {
	security.BelaurVault
}

type testMockMemDBService struct {
	memdb.BelaurMemDB
}

func TestCanMockServiceToNil(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCanMockServiceToNil")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.CAPath = tmp
	belaur.Cfg.VaultPath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})

	t.Run("can mock storage to nil", func(t *testing.T) {
		mcp := new(testMockStorageService)
		MockStorageService(mcp)
		s1, _ := StorageService()
		if _, ok := s1.(*testMockStorageService); !ok {
			t.Fatalf("want type: '%s' got: '%s'", "testMockStorageService", reflect.TypeOf(s1).String())
		}
		MockStorageService(nil)
		s2, _ := StorageService()
		if reflect.TypeOf(s2).String() == "*services.testMockStorageService" {
			t.Fatalf("want type: '%s' got: '%s'", "BoltStorage", reflect.TypeOf(s2).String())
		}
	})

	t.Run("can mock vault to nil", func(t *testing.T) {
		mcp := new(testMockVaultService)
		MockVaultService(mcp)
		s1, _ := DefaultVaultService()
		if _, ok := s1.(*testMockVaultService); !ok {
			t.Fatalf("want type: '%s' got: '%s'", "testMockVaultService", reflect.TypeOf(s1).String())
		}
		MockVaultService(nil)
		s2, _ := DefaultVaultService()
		if reflect.TypeOf(s2).String() == "*services.testMockVaultService" {
			t.Fatalf("got: '%s'", reflect.TypeOf(s2).String())
		}
	})

	t.Run("can mock memdb to nil", func(t *testing.T) {
		mcp := new(testMockMemDBService)
		MockMemDBService(mcp)
		s1, _ := DefaultMemDBService()
		if _, ok := s1.(*testMockMemDBService); !ok {
			t.Fatalf("want type: '%s' got: '%s'", "testMockMemDBService", reflect.TypeOf(s1).String())
		}
		MockMemDBService(nil)
		msp := new(testMockStorageService)
		MockStorageService(msp)
		s2, _ := MemDBService(storeService)
		if reflect.TypeOf(s2).String() == "*services.testMockMemDBService" {
			t.Fatalf("got: '%s'", reflect.TypeOf(s2).String())
		}
	})
}

func TestDefaultVaultStorer(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestDefaultVaultStorer")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.CAPath = tmp
	belaur.Cfg.VaultPath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	v, err := DefaultVaultService()
	if err != nil {
		t.Fatal(err)
	}
	if va, ok := v.(security.BelaurVault); !ok {
		t.Fatal("DefaultVaultService should have given back a BelaurVault. was instead: ", va)
	}
}

func TestDefaultMemDBService(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestDefaultMemDBService")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.CAPath = tmp
	belaur.Cfg.VaultPath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	_, _ = StorageService()
	v, err := DefaultMemDBService()
	if err != nil {
		t.Fatal(err)
	}
	if va, ok := v.(memdb.BelaurMemDB); !ok {
		t.Fatal("DefaultVaultService should have given back a BelaurVault. was instead: ", va)
	}
}
