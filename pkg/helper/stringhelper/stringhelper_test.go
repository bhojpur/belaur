package stringhelper

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

import "testing"

func TestIsContainedInSlice(t *testing.T) {
	slice := []string{"ITEM1", "item1", "ITEM2"}

	// Simple contains check
	if !IsContainedInSlice(slice, "item1", false) {
		t.Fatal("item1 not contained in slice")
	}

	// Check without case insensitive
	if IsContainedInSlice(slice, "Item2", false) {
		t.Fatal("Item2 is contained but should not")
	}

	// Check with case insensitive
	if !IsContainedInSlice(slice, "item2", true) {
		t.Fatal("item2 case insensitive check failed")
	}

	// Should also work for duplicates
	if !IsContainedInSlice(slice, "item1", true) {
		t.Fatal("item1 not contained")
	}
}

func TestDiffSlices(t *testing.T) {
	a := []string{"ITEM1", "item1", "item3", "item2"}
	b := []string{"item1"}

	// Simple case insensitive check
	out := DiffSlices(a, b, true)
	for _, item := range DiffSlices(a, b, true) {
		if item == "ITEM1" || item == "item1" {
			t.Fatalf("item1 should be non-existend: %s", item)
		}
	}

	// Check if it is sorted
	if len(out) != 2 {
		t.Fatalf("expected 2 but got %d", len(out))
	}
	if out[1] != "item3" {
		t.Fatalf("expected '%s' but got '%s'", "item3", out[1])
	}

	// Multiple different values
	b = append(b, "nonexistend")
	b = append(b, "item2")
	out = DiffSlices(a, b, false)

	// Check
	if len(out) != 2 {
		t.Fatalf("expected 2 but got %d", len(out))
	}
	if out[0] != "ITEM1" {
		t.Fatalf("expected '%s' but got '%s'", "ITEM1", out[0])
	}
	if out[1] != "item3" {
		t.Fatalf("expected '%s' but got '%s'", "item3", out[1])
	}
}
