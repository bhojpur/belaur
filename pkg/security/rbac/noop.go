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

type noOpService struct{}

// NewNoOpService is used to instantiated a noOpService for when rbac enabled=false.
func NewNoOpService() Service {
	return &noOpService{}
}

// Enforce no-op enforcement. Allows everything.
func (n noOpService) Enforce(username, method, path string, params map[string]string) error {
	// Allow all
	return nil
}

// AddRole that errors since rbac is not enabled.
func (n noOpService) AddRole(role string, roleRules []RoleRule) error {
	return nil
}

// DeleteRole that errors since rbac is not enabled.
func (n noOpService) DeleteRole(role string) error {
	return nil
}

// GetAllRoles that returns nothing since rbac is not enabled.
func (n noOpService) GetAllRoles() []string {
	return []string{}
}

// GetUserAttachedRoles that errors since rbac is not enabled.
func (n noOpService) GetUserAttachedRoles(username string) ([]string, error) {
	return nil, nil
}

// GetRoleAttachedUsers that errors since rbac is not enabled.
func (n noOpService) GetRoleAttachedUsers(role string) ([]string, error) {
	return nil, nil
}

// AttachRole that errors since rbac is not enabled.
func (n noOpService) AttachRole(username string, role string) error {
	return nil
}

// DetachRole that errors since rbac is not enabled.
func (n noOpService) DetachRole(username string, role string) error {
	return nil
}

func (n noOpService) DeleteUser(username string) error {
	return nil
}
