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
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	belaur "github.com/bhojpur/belaur"
)

func TestGetSHAPair(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestGetSHAPAir")
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

	pair := belaur.SHAPair{}
	pair.PipelineID = 1
	pair.Original = []byte("original")
	pair.Worker = []byte("worker")
	err = store.UpsertSHAPair(pair)
	if err != nil {
		t.Fatal(err)
	}

	ok, p, err := store.GetSHAPair(1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("sha pair not found")
	}

	if p.PipelineID != pair.PipelineID {
		t.Fatalf("pipeline id match error. want %d got %d", pair.PipelineID, p.PipelineID)
	}
	if !bytes.Equal(p.Worker, pair.Worker) {
		t.Fatalf("worker sha match error. want %s got %s", pair.Worker, p.Worker)
	}
	if !bytes.Equal(p.Original, pair.Original) {
		t.Fatalf("original sha match error. want %s got %s", pair.Original, p.Original)
	}
}

func TestUpsertSHAPair(t *testing.T) {
	// Create tmp folder
	tmp, err := ioutil.TempDir("", "TestUpsertSHAPair")
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

	pair := belaur.SHAPair{}
	pair.PipelineID = 1
	pair.Original = []byte("original")
	pair.Worker = []byte("worker")
	err = store.UpsertSHAPair(pair)
	if err != nil {
		t.Fatal(err)
	}
	// Test is upsert overwrites existing records.
	pair.Original = []byte("original2")
	err = store.UpsertSHAPair(pair)
	if err != nil {
		t.Fatal(err)
	}

	ok, p, err := store.GetSHAPair(1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("sha pair not found")
	}

	if p.PipelineID != pair.PipelineID {
		t.Fatalf("pipeline id match error. want %d got %d", pair.PipelineID, p.PipelineID)
	}
	if !bytes.Equal(p.Worker, pair.Worker) {
		t.Fatalf("worker sha match error. want %s got %s", pair.Worker, p.Worker)
	}
	if !bytes.Equal(p.Original, pair.Original) {
		t.Fatalf("original sha match error. want %s got %s", pair.Original, p.Original)
	}
}
