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
	"fmt"
	"strings"

	belaur "github.com/bhojpur/belaur"
)

// EndpointEnforcer represents the interface for enforcing RBAC using the echo.Context.
type EndpointEnforcer interface {
	Enforce(username, method, path string, params map[string]string) error
}

// Enforce uses the echo.Context to enforce RBAC. Uses the APILookup to apply policies to specific endpoints.
func (e *enforcerService) Enforce(username, method, path string, params map[string]string) error {
	group := e.rbacAPILookup

	endpoint, ok := group[path]
	if !ok {
		belaur.Cfg.Logger.Warn("path not mapped to api group", "method", method, "path", path)
		return nil
	}

	perm, ok := endpoint.Methods[method]
	if !ok {
		belaur.Cfg.Logger.Warn("method not mapped to api group path", "path", path, "method", method)
		return nil
	}

	splitAction := strings.Split(perm, "/")
	namespace := splitAction[0]
	action := splitAction[1]

	fullResource := "*"
	if endpoint.Param != "" {
		param := params[endpoint.Param]
		if param == "" {
			return fmt.Errorf("error param %s missing", endpoint.Param)
		}
		fullResource = param
	}

	allow, err := e.enforcer.Enforce(username, namespace, action, fullResource)
	if err != nil {
		return fmt.Errorf("error enforcing rbac: %w", err)
	}
	if !allow {
		return NewErrPermissionDenied(namespace, action, fullResource)
	}

	return nil
}

// ErrPermissionDenied is for when the RBAC enforcement check fails.
type ErrPermissionDenied struct {
	namespace string
	action    string
	resource  string
}

// NewErrPermissionDenied creates a new ErrPermissionDenied.
func NewErrPermissionDenied(namespace string, action string, resource string) *ErrPermissionDenied {
	return &ErrPermissionDenied{namespace: namespace, action: action, resource: resource}
}

func (e *ErrPermissionDenied) Error() string {
	msg := fmt.Sprintf("Permission denied. Must have %s/%s", e.namespace, e.action)
	if e.resource != "*" {
		msg = fmt.Sprintf("%s %s", msg, e.resource)
	}
	return msg
}
