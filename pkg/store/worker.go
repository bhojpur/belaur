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

// WorkerPut stores the given worker in the bolt database.
// Worker object will be overwritten in case it already exist.
func (s *BoltStore) WorkerPut(w *belaur.Worker) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(workerBucket)

		// Marshal worker object
		m, err := json.Marshal(*w)
		if err != nil {
			return err
		}

		// Put worker
		return b.Put([]byte(w.UniqueID), m)
	})
}

// WorkerGetAll returns all existing worker objects from the store.
// It returns an error when the action failed.
func (s *BoltStore) WorkerGetAll() ([]*belaur.Worker, error) {
	var worker []*belaur.Worker

	return worker, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(workerBucket)

		// Iterate all worker.
		return b.ForEach(func(k, v []byte) error {
			// Unmarshal
			w := &belaur.Worker{}
			err := json.Unmarshal(v, w)
			if err != nil {
				return err
			}

			// Append to our list
			worker = append(worker, w)

			return nil
		})
	})
}

// WorkerDeleteAll deletes all worker objects in the bucket.
func (s *BoltStore) WorkerDeleteAll() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Delete bucket
		if err := tx.DeleteBucket(workerBucket); err != nil {
			return err
		}

		_, err := tx.CreateBucket(workerBucket)
		return err
	})
}

// WorkerDelete deletes a worker by the given identifier.
func (s *BoltStore) WorkerDelete(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(workerBucket)

		// Delete entry
		return b.Delete([]byte(id))
	})
}

// WorkerGet gets a worker by the given identifier.
func (s *BoltStore) WorkerGet(id string) (*belaur.Worker, error) {
	var worker *belaur.Worker

	return worker, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(workerBucket)

		// Get worker
		v := b.Get([]byte(id))

		// Check if we found the worker
		if v == nil {
			return nil
		}

		// Unmarshal pipeline object
		worker = &belaur.Worker{}
		err := json.Unmarshal(v, worker)
		if err != nil {
			return err
		}

		return nil
	})
}
