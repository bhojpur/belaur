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
	"time"

	belaur "github.com/bhojpur/belaur"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
)

// UserPut takes the given user and saves it
// to the bolt database. User will be overwritten
// if it already exists.
// It also clears the password field afterwards.
func (s *BoltStore) UserPut(u *belaur.User, encryptPassword bool) error {
	// Encrypt password before we save it
	if encryptPassword {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)
		if err != nil {
			return err
		}
		u.Password = string(hash)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(userBucket)

		// Marshal user object
		m, err := json.Marshal(u)
		if err != nil {
			return err
		}

		// Clear password from origin object
		u.Password = ""

		// Put user
		return b.Put([]byte(u.Username), m)
	})
}

// UserAuth looks up a user by given username.
// Then it compares passwords and returns user obj if
// given password is valid. Returns nil if password was
// wrong or user not found.
func (s *BoltStore) UserAuth(u *belaur.User, updateLastLogin bool) (*belaur.User, error) {
	// Look up user
	user, err := s.UserGet(u.Username)

	// Error occurred and/or user not found
	if err != nil || user == nil {
		return nil, err
	}

	// Check if password is valid
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)); err != nil {
		return nil, err
	}

	// Update last login field
	if updateLastLogin {
		user.LastLogin = time.Now()
		err = s.UserPut(user, false)
		if err != nil {
			return nil, err
		}
	}

	// We will use the user object later.
	// But we don't need the password anymore.
	user.Password = ""

	// Return user
	return user, nil
}

// UserGet looks up a user by given username.
// Returns nil if user was not found.
func (s *BoltStore) UserGet(username string) (*belaur.User, error) {
	user := &belaur.User{}
	err := s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(userBucket)

		// Lookup user
		userRaw := b.Get([]byte(username))

		// User found?
		if userRaw == nil {
			// Nope. That is not an error so just leave
			user = nil
			return nil
		}

		// Unmarshal
		return json.Unmarshal(userRaw, user)
	})

	return user, err
}

// UserGetAll returns all stored users.
func (s *BoltStore) UserGetAll() ([]belaur.User, error) {
	var users []belaur.User

	return users, s.db.View(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(userBucket)

		// Iterate all users and add them to slice
		return b.ForEach(func(k, v []byte) error {
			// create single user object
			u := &belaur.User{}

			// Unmarshal
			err := json.Unmarshal(v, u)
			if err != nil {
				return err
			}

			// Remove password for security reasons
			u.Password = ""

			users = append(users, *u)
			return nil
		})
	})
}

// UserDelete deletes the given user.
func (s *BoltStore) UserDelete(u string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get bucket
		b := tx.Bucket(userBucket)

		// Delete user
		return b.Delete([]byte(u))
	})
}
