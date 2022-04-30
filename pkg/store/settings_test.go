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
	"io/ioutil"
	"os"
	"testing"

	belaur "github.com/bhojpur/belaur"
	"github.com/stretchr/testify/assert"
)

func TestBoltStore_Settings(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestBoltStore_SettingsGet")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	store := NewBoltStore()
	belaur.Cfg.Bolt.Mode = 0600
	err = store.Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	empty := &belaur.StoreConfig{
		ID:          0,
		Poll:        false,
		RBACEnabled: false,
	}

	config, err := store.SettingsGet()
	assert.NoError(t, err)
	assert.EqualValues(t, empty, config)

	cfg := &belaur.StoreConfig{
		ID:          1,
		Poll:        true,
		RBACEnabled: true,
	}

	err = store.SettingsPut(cfg)
	assert.NoError(t, err)

	config, err = store.SettingsGet()
	assert.NoError(t, err)
	assert.EqualValues(t, config, cfg)
}
