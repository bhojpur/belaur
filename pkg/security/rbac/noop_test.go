package rbac

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

	"github.com/stretchr/testify/assert"
)

func TestNoOpService_Enforce_AlwaysReturnsNoError(t *testing.T) {
	svc := NewNoOpService()
	err := svc.Enforce("", "", "", map[string]string{})
	assert.NoError(t, err)
}

func TestNoOpService_AddRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.AddRole("", []RoleRule{})
	assert.NoError(t, err)
}

func TestNoOpService_DeleteRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.DeleteRole("")
	assert.NoError(t, err)
}

func TestNoOpService_GetAllRoles_ReturnsEmptySlice(t *testing.T) {
	svc := NewNoOpService()
	roles := svc.GetAllRoles()
	assert.Equal(t, roles, []string{})
}

func TestNoOpService_GetUserAttachedRoles_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	_, err := svc.GetUserAttachedRoles("")
	assert.NoError(t, err)
}

func TestNoOpService_GetRoleAttachedUsers_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	_, err := svc.GetRoleAttachedUsers("")
	assert.NoError(t, err)
}

func TestNoOpService_AttachRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.AttachRole("", "")
	assert.NoError(t, err)
}

func TestNoOpService_DetachRole_ReturnsErrNotEnabled(t *testing.T) {
	svc := NewNoOpService()
	err := svc.DetachRole("", "")
	assert.NoError(t, err)
}
