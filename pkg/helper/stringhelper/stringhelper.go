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

import (
	"sort"
	"strings"
)

// IsContainedInSlice checks if the given string is contained in the slice.
// If case insensitive is enabled, checks are case insensitive.
func IsContainedInSlice(s []string, c string, caseIns bool) bool {
	for _, curr := range s {
		if caseIns && strings.EqualFold(curr, c) {
			return true
		} else if curr == c {
			return true
		}
	}
	return false
}

// DiffSlices returns A - B set difference of the two given slices.
// It also sorts the returned slice.
func DiffSlices(a []string, b []string, caseIns bool) []string {
	m := map[string]bool{}
	for _, aItem := range a {
		if caseIns {
			aItem = strings.ToLower(aItem)
		}

		m[aItem] = true
	}

	// Check if equal
	for _, bItem := range b {
		if _, ok := m[bItem]; ok {
			m[bItem] = false
		}
	}

	var items []string
	for item, exists := range m {
		if exists {
			items = append(items, item)
		}
	}
	sort.Strings(items)
	return items
}
