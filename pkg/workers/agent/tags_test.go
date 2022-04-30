package agent

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
	"testing"

	"github.com/bhojpur/belaur/pkg/helper/stringhelper"
)

func TestFindLocalBinaries(t *testing.T) {
	testTags := []string{"tag1", "tag2", "tag3"}

	// Simple lookup. Go binary should always exist in dev/test environments.
	tags := findLocalBinaries(testTags)

	// Check output
	if len(tags) < 4 {
		t.Fatalf("expected at least 4 tags but got %d", len(tags))
	}
	if len(stringhelper.DiffSlices(append(testTags, "golang"), tags, false)) != 0 {
		t.Fatalf("expected different output: %#v", tags)
	}

	// Negative language tag
	testTags = append(testTags, "-golang")
	tags = findLocalBinaries(testTags)

	// Check output
	if len(tags) < 3 {
		t.Fatalf("expected at least 3 tags but got %d", len(tags))
	}
	if stringhelper.IsContainedInSlice(tags, "golang", false) {
		t.Fatalf("golang should not be included: %#v", tags)
	}
}
