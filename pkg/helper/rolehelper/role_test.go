package rolehelper

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

	belaur "github.com/bhojpur/belaur"
)

var mockData = []*belaur.UserRoleCategory{
	{
		Name: "CategoryA",
		Roles: []*belaur.UserRole{
			{
				Name: "RoleA",
			},
			{
				Name: "RoleB",
			},
		},
	},
	{
		Name: "CategoryB",
		Roles: []*belaur.UserRole{
			{
				Name: "RoleA",
			},
			{
				Name: "RoleB",
			},
		},
	},
}

func TestFlatRoleName(t *testing.T) {
	value := FullUserRoleName(mockData[0], mockData[0].Roles[0])
	expect := "CategoryARoleA"
	if value != expect {
		t.Fatalf("value %s should equal: %s", value, expect)
	}
}

func TestFlattenUserCategoryRoles(t *testing.T) {
	value := FlattenUserCategoryRoles(mockData)
	expect := []string{"CategoryARoleA", "CategoryARoleB", "CategoryBRoleA", "CategoryBRoleB"}

	for i := range expect {
		if expect[i] != value[i] {
			t.Fatalf("value %s should exist: %s", expect[i], value[i])
		}
	}
}
