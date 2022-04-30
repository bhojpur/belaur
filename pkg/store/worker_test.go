package store

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
	"io/ioutil"
	"os"
	"testing"

	belaur "github.com/bhojpur/belaur"
)

func TestWorkerPut(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreWorkerPut")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	belaur.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Create worker
	w := &belaur.Worker{
		UniqueID: "unique-id",
		Status:   belaur.WorkerActive,
		Tags:     []string{"tag1", "tag2"},
		Name:     "my-worker",
	}

	if err := store.WorkerPut(w); err != nil {
		t.Fatal(err)
	}

	gotWorker, err := store.WorkerGet("unique-id")
	if err != nil {
		t.Fatal(err)
	}

	if gotWorker == nil {
		t.Fatal("expected worker but got nil")
	}
	if gotWorker.Name != "my-worker" {
		t.Fatalf("expected '%s' but got %s", "my-worker", gotWorker.Name)
	}
	if gotWorker.Status != belaur.WorkerActive {
		t.Fatalf("expected '%s' but got %s", belaur.WorkerActive, gotWorker.Status)
	}
	if gotWorker.UniqueID != "unique-id" {
		t.Fatalf("expected '%s' but got '%s'", "unique-id", gotWorker.UniqueID)
	}
	if len(gotWorker.Tags) != 2 {
		t.Fatalf("expected '%d' but got '%d'", 2, len(gotWorker.Tags))
	}
	if gotWorker.Tags[0] != "tag1" {
		t.Fatalf("expected '%s' but got '%s'", "tag1", gotWorker.Tags[0])
	}
	if gotWorker.Tags[1] != "tag2" {
		t.Fatalf("expected '%s' but got '%s'", "tag2", gotWorker.Tags[1])
	}
}

func TestWorkerGetAll(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreWorkerGetAll")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	belaur.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Create worker
	w1 := &belaur.Worker{
		UniqueID: "unique-id1",
		Status:   belaur.WorkerActive,
		Tags:     []string{"tag1", "tag2"},
		Name:     "my-worker1",
	}
	w2 := &belaur.Worker{
		UniqueID: "unique-id2",
		Status:   belaur.WorkerInactive,
		Tags:     []string{"tag1", "tag2"},
		Name:     "my-worker2",
	}

	if err := store.WorkerPut(w1); err != nil {
		t.Fatal(err)
	}
	if err := store.WorkerPut(w2); err != nil {
		t.Fatal(err)
	}

	gotWorker, err := store.WorkerGetAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(gotWorker) != 2 {
		t.Fatalf("expected '%d' worker but got '%d'", 2, len(gotWorker))
	}
	if gotWorker[0].Name != "my-worker1" {
		t.Fatalf("expected '%s' but got '%s'", "my-worker1", gotWorker[0].Name)
	}
	if gotWorker[1].Name != "my-worker2" {
		t.Fatalf("expected '%s' but got '%s'", "my-worker2", gotWorker[1].Name)
	}
}

func TestWorkerDeleteAll(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreWorkerDeleteAll")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	belaur.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Create worker
	w1 := &belaur.Worker{
		UniqueID: "unique-id1",
		Status:   belaur.WorkerActive,
		Tags:     []string{"tag1", "tag2"},
		Name:     "my-worker1",
	}
	w2 := &belaur.Worker{
		UniqueID: "unique-id2",
		Status:   belaur.WorkerInactive,
		Tags:     []string{"tag1", "tag2"},
		Name:     "my-worker2",
	}

	if err := store.WorkerPut(w1); err != nil {
		t.Fatal(err)
	}
	if err := store.WorkerPut(w2); err != nil {
		t.Fatal(err)
	}

	if err := store.WorkerDeleteAll(); err != nil {
		t.Fatal(err)
	}

	gotWorker, err := store.WorkerGetAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(gotWorker) != 0 {
		t.Fatalf("expected '%d' but got '%d'", 0, len(gotWorker))
	}
}

func TestWorkerDelete(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreWorkerDelete")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	belaur.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Create worker
	w1 := &belaur.Worker{
		UniqueID: "unique-id1",
		Status:   belaur.WorkerActive,
		Tags:     []string{"tag1", "tag2"},
		Name:     "my-worker1",
	}
	w2 := &belaur.Worker{
		UniqueID: "unique-id2",
		Status:   belaur.WorkerInactive,
		Tags:     []string{"tag1", "tag2"},
		Name:     "my-worker2",
	}

	if err := store.WorkerPut(w1); err != nil {
		t.Fatal(err)
	}
	if err := store.WorkerPut(w2); err != nil {
		t.Fatal(err)
	}

	if err := store.WorkerDelete("unique-id2"); err != nil {
		t.Fatal(err)
	}

	gotWorker, err := store.WorkerGetAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(gotWorker) != 1 {
		t.Fatalf("expected '%d' but got '%d'", 1, len(gotWorker))
	}
	if gotWorker[0].UniqueID != "unique-id1" {
		t.Fatalf("expected '%s' but got '%s'", "unique-id1", gotWorker[0].UniqueID)
	}
}

func TestWorkerGet(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestStoreWorkerGet")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	belaur.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Create worker
	w := &belaur.Worker{
		UniqueID: "unique-id",
		Status:   belaur.WorkerActive,
		Tags:     []string{"tag1", "tag2"},
		Name:     "my-worker",
	}

	if err := store.WorkerPut(w); err != nil {
		t.Fatal(err)
	}

	gotWorker, err := store.WorkerGet("unique-id")
	if err != nil {
		t.Fatal(err)
	}

	if gotWorker == nil {
		t.Fatal("expected worker but got nil")
	}
	if gotWorker.Name != "my-worker" {
		t.Fatalf("expected '%s' but got %s", "my-worker", gotWorker.Name)
	}
	if gotWorker.Status != belaur.WorkerActive {
		t.Fatalf("expected '%s' but got %s", belaur.WorkerActive, gotWorker.Status)
	}
	if gotWorker.UniqueID != "unique-id" {
		t.Fatalf("expected '%s' but got '%s'", "unique-id", gotWorker.UniqueID)
	}
	if len(gotWorker.Tags) != 2 {
		t.Fatalf("expected '%d' but got '%d'", 2, len(gotWorker.Tags))
	}
	if gotWorker.Tags[0] != "tag1" {
		t.Fatalf("expected '%s' but got '%s'", "tag1", gotWorker.Tags[0])
	}
	if gotWorker.Tags[1] != "tag2" {
		t.Fatalf("expected '%s' but got '%s'", "tag2", gotWorker.Tags[1])
	}
}
