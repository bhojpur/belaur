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
	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/rolehelper"
)

// CreatePermissionsIfNotExisting iterates any existing users and creates default permissions if they don't exist.
// This is most probably when they have upgraded to the Bhojpur Belaur version where permissions was added.
func (s *BoltStore) CreatePermissionsIfNotExisting() error {
	users, _ := s.UserGetAll()
	for _, user := range users {
		perms, err := s.UserPermissionsGet(user.Username)
		if err != nil {
			return err
		}
		if perms == nil {
			perms := &belaur.UserPermission{
				Username: user.Username,
				Roles:    rolehelper.FlattenUserCategoryRoles(rolehelper.DefaultUserRoles),
				Groups:   []string{},
			}
			err := s.UserPermissionsPut(perms)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
