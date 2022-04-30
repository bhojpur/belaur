package assethelper

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
	rice "github.com/GeertJohan/go.rice"
)

const (
	rbacBuiltinPolicy = "rbac-policy.csv"
	rbacModel         = "rbac-model.conf"
	rbacAPIMappings   = "rbac-api-mappings.yml"
)

func loadStaticFile(filename string) (string, error) {
	box, err := rice.FindBox("../../static")
	if err != nil {
		return "", err
	}
	filestr, err := box.String(filename)
	if err != nil {
		return "", err
	}
	return filestr, nil
}

// LoadRBACBuiltinPolicy loads the builtin rbac-policy.csv
func LoadRBACBuiltinPolicy() (string, error) {
	return loadStaticFile(rbacBuiltinPolicy)
}

// LoadRBACAPIMappings loads the rbac-api-mappings.yml
func LoadRBACAPIMappings() (string, error) {
	return loadStaticFile(rbacAPIMappings)
}

// LoadRBACModel loads the rbac-model.conf
func LoadRBACModel() (string, error) {
	return loadStaticFile(rbacModel)
}
