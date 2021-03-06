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

// CreatePipelinePut adds a pipeline which
// is not yet compiled but is about to.
func (s *BoltStore) CreatePipelinePut(p *belaur.CreatePipeline) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(createPipelineBucket)

		// Marshal pipeline object
		m, err := json.Marshal(p)
		if err != nil {
			return err
		}

		// Put pipeline
		return b.Put([]byte(p.ID), m)
	})
}

// CreatePipelineGet returns all available create pipeline
// objects in the store.
func (s *BoltStore) CreatePipelineGet() ([]belaur.CreatePipeline, error) {
	// create list
	var pipelineList []belaur.CreatePipeline

	return pipelineList, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(createPipelineBucket)

		// Iterate all created pipelines.
		// TODO: We might get a huge list here. It might be better
		// to just search for the last 20 elements.
		return b.ForEach(func(k, v []byte) error {
			// create single pipeline object
			p := &belaur.CreatePipeline{}

			// Unmarshal
			err := json.Unmarshal(v, p)
			if err != nil {
				return err
			}

			pipelineList = append(pipelineList, *p)
			return nil
		})
	})
}

// PipelinePut puts a pipeline into the store.
// On persist, the pipeline will get a unique id.
func (s *BoltStore) PipelinePut(p *belaur.Pipeline) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get pipeline bucket
		b := tx.Bucket(pipelineBucket)

		// Generate ID for the pipeline if its new.
		if p.ID == 0 {
			id, err := b.NextSequence()
			if err != nil {
				return err
			}
			p.ID = int(id)
		}

		// Marshal pipeline data into bytes.
		buf, err := json.Marshal(p)
		if err != nil {
			return err
		}

		// Persist bytes to pipelines bucket.
		return b.Put(itob(p.ID), buf)
	})
}

// PipelineGet gets a pipeline by given id.
func (s *BoltStore) PipelineGet(id int) (*belaur.Pipeline, error) {
	var pipeline *belaur.Pipeline

	return pipeline, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineBucket)

		// Get pipeline
		v := b.Get(itob(id))

		// Check if we found the pipeline
		if v == nil {
			return nil
		}

		// Unmarshal pipeline object
		pipeline = &belaur.Pipeline{}
		err := json.Unmarshal(v, pipeline)
		if err != nil {
			return err
		}

		return nil
	})
}

// PipelineGetByName looks up a pipeline by the given name.
// Returns nil if pipeline was not found.
func (s *BoltStore) PipelineGetByName(n string) (*belaur.Pipeline, error) {
	var pipeline *belaur.Pipeline

	return pipeline, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineBucket)

		// Iterate all created pipelines.
		return b.ForEach(func(k, v []byte) error {
			// create single pipeline object
			p := &belaur.Pipeline{}

			// Unmarshal
			err := json.Unmarshal(v, p)
			if err != nil {
				return err
			}

			// Is this pipeline we are looking for?
			if p.Name == n {
				pipeline = p
			}

			return nil
		})
	})
}

// PipelineGetRunHighestID looks for the highest public id for the given pipeline.
func (s *BoltStore) PipelineGetRunHighestID(p *belaur.Pipeline) (int, error) {
	var highestID int

	return highestID, s.db.View(func(tx *bolt.Tx) error {
		// Get Bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipeline runs.
		return b.ForEach(func(k, v []byte) error {
			// create single run object
			r := &belaur.PipelineRun{}

			// Unmarshal
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Is this a run from our pipeline?
			if r.PipelineID == p.ID {
				// Check if the id is higher than what we found before?
				if r.ID > highestID {
					highestID = r.ID
				}
			}

			return nil
		})
	})
}

// PipelinePutRun takes the given pipeline run and puts it into the store.
// If a pipeline run already exists in the store it will be overwritten.
func (s *BoltStore) PipelinePutRun(r *belaur.PipelineRun) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineRunBucket)

		// Marshal data into bytes.
		buf, err := json.Marshal(r)
		if err != nil {
			return err
		}

		// Persist bytes into bucket.
		return b.Put([]byte(r.UniqueID), buf)
	})
}

// PipelineGetScheduled returns the scheduled pipelines with a return limit.
func (s *BoltStore) PipelineGetScheduled(limit int) ([]*belaur.PipelineRun, error) {
	// create returning list
	var runList []*belaur.PipelineRun

	return runList, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipelines.
		return b.ForEach(func(k, v []byte) error {
			// Check if we already reached the limit
			if len(runList) >= limit {
				return nil
			}

			// Unmarshal
			r := &belaur.PipelineRun{}
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Check if the run is still scheduled not executed yet
			if r.Status == belaur.RunNotScheduled {
				// Append to our list
				runList = append(runList, r)
			}

			return nil
		})
	})
}

// PipelineGetRunByPipelineIDAndID looks for pipeline run by given pipeline id and run id.
func (s *BoltStore) PipelineGetRunByPipelineIDAndID(pipelineid int, runid int) (*belaur.PipelineRun, error) {
	var pipelineRun *belaur.PipelineRun

	return pipelineRun, s.db.View(func(tx *bolt.Tx) error {
		// Get Bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipeline runs.
		return b.ForEach(func(k, v []byte) error {
			// create single run object
			r := &belaur.PipelineRun{}

			// Unmarshal
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Is this a run from our pipeline?
			if r.PipelineID == pipelineid && r.ID == runid {
				pipelineRun = r
			}

			return nil
		})
	})
}

// PipelineGetRunByID returns the pipeline run by internal unique id.
func (s *BoltStore) PipelineGetRunByID(runID string) (*belaur.PipelineRun, error) {
	var pipelineRun *belaur.PipelineRun

	return pipelineRun, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineRunBucket)

		// Get pipeline run
		v := b.Get([]byte(runID))

		// Check if we found the pipeline
		if v == nil {
			return nil
		}

		// Unmarshal pipeline object
		pipelineRun = &belaur.PipelineRun{}
		err := json.Unmarshal(v, pipelineRun)
		if err != nil {
			return err
		}

		return nil
	})
}

// PipelineGetAllRunsByPipelineID looks for all pipeline runs by the given pipeline id.
func (s *BoltStore) PipelineGetAllRunsByPipelineID(pipelineID int) ([]belaur.PipelineRun, error) {
	var runs []belaur.PipelineRun

	return runs, s.db.View(func(tx *bolt.Tx) error {
		// Get Bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipeline runs.
		return b.ForEach(func(k, v []byte) error {
			// create single run object
			r := &belaur.PipelineRun{}

			// Unmarshal
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Is this a run from our pipeline?
			if r.PipelineID == pipelineID {
				// add this to our list
				runs = append(runs, *r)
			}

			return nil
		})
	})
}

// PipelineGetAllRuns loads all existing pipeline runs.
func (s *BoltStore) PipelineGetAllRuns() ([]belaur.PipelineRun, error) {
	var runs []belaur.PipelineRun

	return runs, s.db.View(func(tx *bolt.Tx) error {
		// Get Bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipeline runs.
		return b.ForEach(func(k, v []byte) error {
			// create single run object
			r := &belaur.PipelineRun{}

			// Unmarshal
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Append this run
			runs = append(runs, *r)

			return nil
		})
	})
}

// PipelineGetLatestRun returns the latest run by the given pipeline id.
func (s *BoltStore) PipelineGetLatestRun(pipelineID int) (*belaur.PipelineRun, error) {
	var run *belaur.PipelineRun

	return run, s.db.View(func(tx *bolt.Tx) error {
		// Get Bucket
		b := tx.Bucket(pipelineRunBucket)

		// Iterate all pipeline runs.
		return b.ForEach(func(k, v []byte) error {
			// create single run object
			r := &belaur.PipelineRun{}

			// Unmarshal
			err := json.Unmarshal(v, r)
			if err != nil {
				return err
			}

			// Is this a run from our pipeline?
			if r.PipelineID == pipelineID {
				// Check if this is the latest run
				if run == nil || run.StartDate.Before(r.StartDate) {
					// set it
					run = r
				}
			}

			return nil
		})
	})
}

// PipelineDelete deletes the pipeline with the given id.
func (s *BoltStore) PipelineDelete(id int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineBucket)

		// Delete pipeline
		return b.Delete(itob(id))
	})
}

// PipelineRunDelete deletes the pipeline run with the given id.
func (s *BoltStore) PipelineRunDelete(uniqueID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(pipelineRunBucket)

		// Delete pipeline
		return b.Delete([]byte(uniqueID))
	})
}
