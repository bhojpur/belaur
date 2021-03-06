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
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/security/rbac"
)

type mockRBACSvc struct {
	rbac.Service
}

func (e *mockRBACSvc) GetAllRoles() []string {
	return []string{"role-a", "role-b"}
}

func (e *mockRBACSvc) AddRole(role string, roleRules []rbac.RoleRule) error {
	if role == "success" {
		return nil
	}
	return errors.New("add error")
}

func (e *mockRBACSvc) DeleteRole(role string) error {
	if role == "delme" {
		return nil
	}
	return errors.New("delete error")
}

func (e *mockRBACSvc) GetUserAttachedRoles(username string) ([]string, error) {
	if username == "test" {
		return []string{"role-a", "role-b"}, nil
	}
	return nil, errors.New("an error")
}

func (e *mockRBACSvc) GetRoleAttachedUsers(role string) ([]string, error) {
	if role == "test" {
		return []string{"user-a", "user-b"}, nil
	}
	return nil, errors.New("an error")
}

func (e *mockRBACSvc) AttachRole(username string, role string) error {
	if role == "test-role" && username == "test-user" {
		return nil
	}
	return errors.New("an error")
}

func (e *mockRBACSvc) DetachRole(username string, role string) error {
	if role == "test-role" && username == "test-user" {
		return nil
	}
	return errors.New("an error")
}

func Test_rbacHandler_AddRole(t *testing.T) {
	handler := Provider{
		svc: &mockRBACSvc{},
	}

	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.NewNullLogger()
	defer func() {
		belaur.Cfg = nil
	}()

	e := echo.New()

	t.Run("success (200) if add is successful", func(t *testing.T) {
		body := `[
	{
		"namespace": "secrets",
		"action": "delete",
		"resource": "*",
		"effect": "deny"
	},
	{
		"namespace": "secrets",
		"action": "get",
		"resource": "*",
		"effect": "allow"
	}
]`
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")
		c.SetParamNames("role")
		c.SetParamValues("success")

		err := handler.AddRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, rec.Body.String(), "Role created successfully.")
	})

	t.Run("error (400) if role is not provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")

		err := handler.AddRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, rec.Body.String(), "Must provide role.")
	})

	t.Run("error (400) if body is invalid", func(t *testing.T) {
		body := `{}`
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")
		c.SetParamNames("role")
		c.SetParamValues("success")

		err := handler.AddRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, rec.Body.String(), "Invalid body provided.")
	})

	t.Run("error (500) if error occurs adding the role", func(t *testing.T) {
		body := `[]`
		req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer([]byte(body)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")
		c.SetParamNames("role")
		c.SetParamValues("error")

		err := handler.AddRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusInternalServerError)
		assert.Equal(t, rec.Body.String(), "An error occurred while adding the role.")
	})
}

func Test_Provider_DeleteRole(t *testing.T) {
	handler := Provider{
		svc: &mockRBACSvc{},
	}

	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.NewNullLogger()
	defer func() {
		belaur.Cfg = nil
	}()

	e := echo.New()

	t.Run("success (200) if role is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")
		c.SetParamNames("role")
		c.SetParamValues("delme")

		err := handler.DeleteRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, rec.Body.String(), "Role deleted successfully.")
	})

	t.Run("error (400) if no role is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")

		err := handler.DeleteRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, rec.Body.String(), "Must provide role.")
	})

	t.Run("error (400) if error occurs deleting role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role")
		c.SetParamNames("role")
		c.SetParamValues("error")

		err := handler.DeleteRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusInternalServerError)
		assert.Equal(t, rec.Body.String(), "An error occurred while deleting the role.")
	})
}

func Test_Provider_getAllRoles(t *testing.T) {
	handler := Provider{
		svc: &mockRBACSvc{},
	}

	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.NewNullLogger()
	defer func() {
		belaur.Cfg = nil
	}()

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/rbac/roles")

	err := handler.GetAllRoles(c)
	assert.NoError(t, err)
	assert.Equal(t, rec.Code, http.StatusOK)
	assert.Equal(t, rec.Body.String(), "[\"role-a\",\"role-b\"]\n")
}

func Test_Provider_getUserAttachedRoles(t *testing.T) {
	handler := Provider{
		svc: &mockRBACSvc{},
	}

	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.NewNullLogger()
	defer func() {
		belaur.Cfg = nil
	}()

	e := echo.New()

	t.Run("success (200) if user is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/users/:username/rbac/roles")
		c.SetParamNames("username")
		c.SetParamValues("test")

		err := handler.GetUserAttachedRoles(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, rec.Body.String(), "[\"role-a\",\"role-b\"]\n")
	})

	t.Run("error (400) if no username is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/users/:username/rbac/roles")

		err := handler.GetUserAttachedRoles(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, rec.Body.String(), "Must provide username.")
	})

	t.Run("error (500) if error occurs getting roles", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/users/:username/rbac/roles")
		c.SetParamNames("username")
		c.SetParamValues("error")

		err := handler.GetUserAttachedRoles(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusInternalServerError)
		assert.Equal(t, rec.Body.String(), "An error occurred while getting the roles.")
	})
}

func Test_Provider_GetRolesAttachedUsers(t *testing.T) {
	handler := Provider{
		svc: &mockRBACSvc{},
	}

	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.NewNullLogger()
	defer func() {
		belaur.Cfg = nil
	}()

	e := echo.New()

	t.Run("success (200) if role is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attached")
		c.SetParamNames("role")
		c.SetParamValues("test")

		err := handler.GetRoleAttachedUsers(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, rec.Body.String(), "[\"user-a\",\"user-b\"]\n")
	})

	t.Run("error (400) if no role is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attached")

		err := handler.GetRoleAttachedUsers(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, rec.Body.String(), "Must provide role.")
	})

	t.Run("error (500) if an error occurs getting users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attached")
		c.SetParamNames("role")
		c.SetParamValues("error")

		err := handler.GetRoleAttachedUsers(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusInternalServerError)
		assert.Equal(t, rec.Body.String(), "An error occurred while getting the users.")
	})
}

func Test_Provider_attachRole(t *testing.T) {
	provider := Provider{
		svc: &mockRBACSvc{},
	}

	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.NewNullLogger()
	defer func() {
		belaur.Cfg = nil
	}()

	e := echo.New()

	t.Run("success (200) if role is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role", "username")
		c.SetParamValues("test-role", "test-user")

		err := provider.AttachRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, rec.Body.String(), "Role attached successfully.")
	})

	t.Run("error (400) if no role is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")

		err := provider.AttachRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, rec.Body.String(), "Must provide role.")
	})

	t.Run("error (400) if no username is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role")
		c.SetParamValues("test-role")

		err := provider.AttachRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, rec.Body.String(), "Must provide username.")
	})

	t.Run("error (500) if error occurs attaching the role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role", "username")
		c.SetParamValues("error", "error")

		err := provider.AttachRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusInternalServerError)
		assert.Equal(t, rec.Body.String(), "An error occurred while attaching the role.")
	})
}

func Test_Provider_detachRole(t *testing.T) {
	handler := Provider{
		svc: &mockRBACSvc{},
	}

	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.NewNullLogger()
	defer func() {
		belaur.Cfg = nil
	}()

	e := echo.New()

	t.Run("success (200) if role is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role", "username")
		c.SetParamValues("test-role", "test-user")

		err := handler.DetachRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, rec.Body.String(), "Role detached successfully.")
	})

	t.Run("error (400) if no role is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")

		err := handler.DetachRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, rec.Body.String(), "Must provide role.")
	})

	t.Run("error (400) if no username is provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role")
		c.SetParamValues("test-role")

		err := handler.DetachRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, rec.Body.String(), "Must provide username.")
	})

	t.Run("error (500) if error occurs detaching the role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/rbac/roles/:role/attach/:username")
		c.SetParamNames("role", "username")
		c.SetParamValues("error", "error")

		err := handler.DetachRole(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusInternalServerError)
		assert.Equal(t, rec.Body.String(), "An error occurred while detaching the role.")
	})
}
