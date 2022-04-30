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
	"encoding/json"

	belaur "github.com/bhojpur/belaur"
	bolt "go.etcd.io/bbolt"
)

// UpsertSHAPair creates or updates a record for a SHA pair of the original SHA and the
// rebuilt Worker SHA for a pipeline.
func (s *BoltStore) UpsertSHAPair(pair belaur.SHAPair) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(shaPairBucket)

		// Marshal SHAPair struct
		m, err := json.Marshal(pair)
		if err != nil {
			return err
		}

		// Put SHAPair
		return b.Put(itob(pair.PipelineID), m)
	})
}

// GetSHAPair returns a pair of shas for this pipeline run.
func (s *BoltStore) GetSHAPair(pipelineID int) (ok bool, pair belaur.SHAPair, err error) {
	return ok, pair, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(shaPairBucket)

		// Get SHAPair
		v := b.Get(itob(pipelineID))

		// Check if we found the SHAPair
		if v == nil {
			ok = false
			return nil
		}

		// Unmarshal SHAPair struct
		err := json.Unmarshal(v, &pair)
		if err != nil {
			return err
		}
		ok = true
		return nil
	})
}
