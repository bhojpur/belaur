package api

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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	belaur "github.com/bhojpur/belaur"
)

// RegisterResponse represents a response from API registration
// call.
type RegisterResponse struct {
	UniqueID string `json:"uniqueid"`
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	CACert   string `json:"cacert"`
}

// RegisterWorker registers a new worker at a Bhojpur Belaur instance.
// It uses the given secret for authentication and returns certs
// which can be used for a future mTLS connection.
func RegisterWorker(host, secret, name string, tags []string) (*RegisterResponse, error) {
	fullURL := fmt.Sprintf("%s/api/%s/worker/register", host, belaur.APIVersion)
	resp, err := http.PostForm(fullURL,
		url.Values{
			"secret": {secret},
			"tags":   tags,
			"name":   {name},
		})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the content
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check the return code first
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to register worker. Return code was '%d' and message was: %s", resp.StatusCode, string(body))
	}

	// Unmarshal the json response
	regResp := RegisterResponse{}
	if err = json.Unmarshal(body, &regResp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal registration response: %s", string(body))
	}

	return &regResp, nil
}
