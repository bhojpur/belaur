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

const (
	configSettings = "belaur_config_settings"
)

// SettingsPut puts settings into the store.
func (s *BoltStore) SettingsPut(c *belaur.StoreConfig) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get settings bucket
		b := tx.Bucket(settingsBucket)

		// Marshal pipeline data into bytes.
		buf, err := json.Marshal(c)
		if err != nil {
			return err
		}

		// Persist bytes to settings bucket.
		return b.Put([]byte(configSettings), buf)
	})
}

// SettingsGet gets a pipeline by given id.
func (s *BoltStore) SettingsGet() (*belaur.StoreConfig, error) {
	var config = &belaur.StoreConfig{}

	return config, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(settingsBucket)

		// Get pipeline
		v := b.Get([]byte(configSettings))

		// Check if we found the pipeline
		if v == nil {
			return nil
		}

		// Unmarshal pipeline object
		err := json.Unmarshal(v, config)
		if err != nil {
			return err
		}

		return nil
	})
}
