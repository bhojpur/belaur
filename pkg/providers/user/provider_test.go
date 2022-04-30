package user

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
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	belaur "github.com/bhojpur/belaur"
	gStore "github.com/bhojpur/belaur/pkg/store"
)

type mockUserStorageService struct {
	gStore.BelaurStore
	user *belaur.User
	err  error
}

func (m *mockUserStorageService) UserAuth(u *belaur.User, updateLastLogin bool) (*belaur.User, error) {
	return m.user, m.err
}

func (m *mockUserStorageService) UserGet(username string) (*belaur.User, error) {
	return m.user, m.err
}

func (m *mockUserStorageService) UserPut(u *belaur.User, encryptPassword bool) error {
	return nil
}

func (m *mockUserStorageService) UserPermissionsGet(username string) (*belaur.UserPermission, error) {
	return &belaur.UserPermission{}, nil
}

func TestUserLoginHMACKey(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestUserLoginHMACKey")
	dataDir := tmp

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		JWTKey: []byte("hmac-jwt-key"),
		Logger: hclog.New(&hclog.LoggerOptions{
			Level:  hclog.Trace,
			Output: hclog.DefaultOutput,
			Name:   "Belaur",
		}),
		DataPath: dataDir,
		Mode:     belaur.ModeServer,
	}

	e := echo.New()

	body := map[string]string{
		"username": "admin",
		"password": "admin",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ms := &mockUserStorageService{user: &belaur.User{
		Username: "username",
		Password: "password",
	}, err: nil}

	provider := NewProvider(ms, nil)
	if err := provider.UserLogin(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}

	data, err := ioutil.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	user := &belaur.User{}
	err = json.Unmarshal(data, user)
	if err != nil {
		t.Fatalf("error unmarshaling response %v", err.Error())
	}
	token, _, err := new(jwt.Parser).ParseUnverified(user.Tokenstring, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("error parsing the token %v", err.Error())
	}
	alg := "HS256"
	if token.Header["alg"] != alg {
		t.Fatalf("expected token alg %v got %v", alg, token.Header["alg"])
	}
}

func TestDeleteUserNotAllowedForAutoUser(t *testing.T) {
	dataDir, _ := ioutil.TempDir("", "TestDeleteUserNotAllowedForAutoUser")

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		Logger:    hclog.NewNullLogger(),
		DataPath:  dataDir,
		CAPath:    dataDir,
		VaultPath: dataDir,
	}

	e := echo.New()
	req := httptest.NewRequest(echo.DELETE, "/api/"+belaur.APIVersion+"/user/auto", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	provider := NewProvider(nil, nil)
	_ = provider.UserDelete(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
	}
}

func TestResetAutoUserTriggerToken(t *testing.T) {
	dataDir, _ := ioutil.TempDir("", "TestResetAutoUserTriggerToken")

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		Logger:    hclog.NewNullLogger(),
		DataPath:  dataDir,
		CAPath:    dataDir,
		VaultPath: dataDir,
	}

	t.Run("reset auto user token", func(t *testing.T) {
		user := belaur.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		ms := &mockUserStorageService{user: &user, err: nil}
		e := echo.New()
		req := httptest.NewRequest(echo.PUT, "/api/"+belaur.APIVersion+"/user/auto/reset-trigger-token", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("username")
		c.SetParamValues("auto")

		provider := NewProvider(ms, nil)
		_ = provider.UserResetTriggerToken(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v; error: %s", http.StatusOK, rec.Code, rec.Body.String())
		}

		if user.TriggerToken == "triggerToken" {
			t.Fatal("user's trigger token should have been reset")
		}
	})
	t.Run("only auto user can reset trigger token", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.PUT, "/api/"+belaur.APIVersion+"/user/auto2/reset-trigger-token", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("username")
		c.SetParamValues("auto2")

		provider := NewProvider(nil, nil)
		_ = provider.UserResetTriggerToken(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v; error: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
		}
	})
}

func TestUserLoginRSAKey(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestUserLoginRSAKey")
	dataDir := tmp

	defer func() {
		belaur.Cfg = nil
	}()

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	belaur.Cfg = &belaur.Config{
		JWTKey: key,
		Logger: hclog.New(&hclog.LoggerOptions{
			Level:  hclog.Trace,
			Output: hclog.DefaultOutput,
			Name:   "Belaur",
		}),
		DataPath: dataDir,
		Mode:     belaur.ModeServer,
	}
	ms := &mockUserStorageService{user: &belaur.User{
		Username: "username",
		Password: "password",
	}, err: nil}

	e := echo.New()

	body := map[string]string{
		"username": "admin",
		"password": "admin",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	provider := NewProvider(ms, nil)
	if err := provider.UserLogin(c); err != nil {
		t.Fatal(err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}

	data, err := ioutil.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	user := &belaur.User{}
	err = json.Unmarshal(data, user)
	if err != nil {
		t.Fatalf("error unmarshaling response %v", err.Error())
	}
	token, _, err := new(jwt.Parser).ParseUnverified(user.Tokenstring, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("error parsing the token %v", err.Error())
	}
	alg := "RS512"
	if token.Header["alg"] != alg {
		t.Fatalf("expected token alg %v got %v", alg, token.Header["alg"])
	}
}

type mockStore struct {
	gStore.BelaurStore
	userPermissionsGetFunc func(username string) (*belaur.UserPermission, error)
	userPermissionsPutFunc func(perms *belaur.UserPermission) error
}

func (s *mockStore) UserPermissionsGet(username string) (*belaur.UserPermission, error) {
	return s.userPermissionsGetFunc(username)
}

func (s *mockStore) UserPermissionsPut(perms *belaur.UserPermission) error {
	return s.userPermissionsPutFunc(perms)
}

func TestUserPutPermissions(t *testing.T) {
	ms := &mockStore{
		userPermissionsPutFunc: func(perms *belaur.UserPermission) error {
			return nil
		},
	}

	bts, _ := json.Marshal(&belaur.UserPermission{
		Username: "test-user",
		Roles:    []string{"TestRole"},
		Groups:   []string{},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(bts))
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/user/:username/permissions")
	c.SetParamNames("username")
	c.SetParamValues("test-user")

	provider := NewProvider(ms, nil)
	_ = provider.UserPutPermissions(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("code is %d. expected %d", rec.Code, http.StatusOK)
	}
}

func TestUserPutPermissionsError(t *testing.T) {
	ms := &mockStore{
		userPermissionsPutFunc: func(perms *belaur.UserPermission) error {
			return errors.New("test error")
		},
	}

	bts, _ := json.Marshal(&belaur.UserPermission{
		Username: "test-user",
		Roles:    []string{"TestRole"},
		Groups:   []string{},
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(bts))
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/user/:username/permissions")
	c.SetParamNames("username")
	c.SetParamValues("test-user")

	provider := NewProvider(ms, nil)
	_ = provider.UserPutPermissions(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code is %d. expected %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUserGetPermissions(t *testing.T) {
	ms := &mockStore{
		userPermissionsGetFunc: func(username string) (*belaur.UserPermission, error) {
			return &belaur.UserPermission{
				Username: "test-user",
				Roles:    []string{"TestRole"},
				Groups:   []string{},
			}, nil
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/user/:username/permissions")
	c.SetParamNames("username")
	c.SetParamValues("test-user")

	provider := NewProvider(ms, nil)
	_ = provider.UserGetPermissions(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("code is %d. expected %d", rec.Code, http.StatusOK)
	}
}

func TestUserGetPermissionsErrors(t *testing.T) {
	ms := &mockStore{
		userPermissionsGetFunc: func(username string) (*belaur.UserPermission, error) {
			return nil, errors.New("test error")
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/user/:username/permissions")
	c.SetParamNames("username")
	c.SetParamValues("test-user")

	provider := NewProvider(ms, nil)
	_ = provider.UserGetPermissions(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code is %d. expected %d", rec.Code, http.StatusBadRequest)
	}
}
