package belaurscheduler

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
	"sync"

	belaur "github.com/bhojpur/belaur"
)

// workload is a wrapper around a single job object.
type workload struct {
	finishedSig chan bool
	done        bool
	started     bool
	job         *belaur.Job
}

// managedWorkloads holds workloads.
// managedWorkloads can be safely shared between goroutines.
type managedWorkloads struct {
	sync.RWMutex

	workloads []workload
}

// newManagedWorkloads creates a new instance of managedWorkloads.
func newManagedWorkloads() *managedWorkloads {
	mw := &managedWorkloads{
		workloads: make([]workload, 0),
	}

	return mw
}

// Append appends a new workload to managedWorkloads.
func (mw *managedWorkloads) Append(wl workload) {
	mw.Lock()
	defer mw.Unlock()

	mw.workloads = append(mw.workloads, wl)
}

// GetByID looks up the workload by the given id.
func (mw *managedWorkloads) GetByID(id uint32) *workload {
	var foundWorkload *workload
	for wl := range mw.Iter() {
		if wl.job.ID == id {
			copyWL := wl
			foundWorkload = &copyWL
		}
	}

	return foundWorkload
}

// Replace takes the given workload and replaces it in the managedWorkloads
// slice. Return true when success otherwise false.
func (mw *managedWorkloads) Replace(wl workload) bool {
	mw.Lock()
	defer mw.Unlock()

	// Search for the id
	i := -1
	for id, currWL := range mw.workloads {
		if currWL.job.ID == wl.job.ID {
			i = id
			break
		}
	}

	// We got it?
	if i == -1 {
		return false
	}

	// Yes
	mw.workloads[i] = wl
	return true
}

// Iter iterates over the workloads in the concurrent slice.
func (mw *managedWorkloads) Iter() <-chan workload {
	c := make(chan workload)

	go func() {
		mw.RLock()
		defer mw.RUnlock()
		for _, mw := range mw.workloads {
			c <- mw
		}
		close(c)
	}()

	return c
}
