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
	"strconv"
	"sync"
	"testing"

	belaur "github.com/bhojpur/belaur"
)

func TestNewWorkload(t *testing.T) {
	mw := newManagedWorkloads()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			finished := make(chan bool)
			title := strconv.Itoa(j)
			wl := workload{
				done:        true,
				finishedSig: finished,
				job: &belaur.Job{
					Description: "Test job",
					ID:          uint32(j),
					Title:       "Test " + title,
				},
				started: true,
			}
			mw.Append(wl)
		}(i)
	}
	wg.Wait()
	if len(mw.workloads) != 10 {
		t.Fatal("workload len want: 10, was:", len(mw.workloads))
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			wl := mw.GetByID(uint32(j))
			if wl == nil {
				t.Error("failed to find a job that was created previously. failed id: ", j)
			}
		}(i)
	}
	if t.Failed() {
		t.Fatal("there were errors in the above function. ")
	}
	wg.Wait()
}

func TestReplaceWorkloadFlow(t *testing.T) {
	mw := newManagedWorkloads()
	finished := make(chan bool)
	wl := workload{
		done:        true,
		finishedSig: finished,
		job: &belaur.Job{
			Description: "Test job",
			ID:          1,
			Title:       "Test",
		},
		started: true,
	}
	mw.Append(wl)
	t.Run("replace works", func(t *testing.T) {
		replaceWl := workload{
			done:        true,
			finishedSig: finished,
			job: &belaur.Job{
				Description: "Test job replaced",
				ID:          1,
				Title:       "Test replaced",
			},
			started: true,
		}
		v := mw.Replace(replaceWl)
		if !v {
			t.Fatalf("return should be true. was false.")
		}
		l := mw.GetByID(1)
		if l.job.Title != "Test replaced" {
			t.Fatalf("got title: %s. wanted: 'Test replaced'", l.job.Title)
		}
	})

	t.Run("returns false if workload was not found", func(t *testing.T) {
		replaceWl := workload{
			done:        true,
			finishedSig: finished,
			job: &belaur.Job{
				Description: "Test job replaced",
				ID:          2,
				Title:       "Test replaced",
			},
			started: true,
		}
		v := mw.Replace(replaceWl)
		if v {
			t.Fatalf("return should be false. was true.")
		}
		l := mw.GetByID(2)
		if l != nil {
			t.Fatal("should have not found id 2 which was replaced:", l)
		}
	})
}
