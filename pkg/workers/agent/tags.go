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
	"fmt"
	"os/exec"
	"strings"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/stringhelper"
)

var (
	supportedBinaries = map[belaur.PipelineType]string{
		belaur.PTypePython: "python",
		belaur.PTypeJava:   "mvn",
		belaur.PTypeCpp:    "make",
		belaur.PTypeGolang: "go",
		belaur.PTypeRuby:   "gem",
	}
)

// findLocalBinaries finds all supported local binaries for local execution
// or build of pipelines in the local "path" variable. If the tags list contains
// a negative language value, the language check is skipped.
func findLocalBinaries(tags []string) []string {
	var foundSuppBinary []string

	// Iterate all supported binary names
	for key, binName := range supportedBinaries {
		// Check if negative tags value has been set
		if stringhelper.IsContainedInSlice(tags, fmt.Sprintf("-%s", key.String()), true) {
			continue
		}

		// Check if the binary name is available
		if _, err := exec.LookPath(binName); err == nil {
			// It is available. Add the tag to the list.
			foundSuppBinary = append(foundSuppBinary, key.String())
		}
	}

	// Add given tags
	for _, tag := range tags {
		if !strings.HasPrefix(tag, "-") {
			foundSuppBinary = append(foundSuppBinary, tag)
		}
	}

	return foundSuppBinary
}
