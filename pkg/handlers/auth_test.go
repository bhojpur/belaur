package handlers

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
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/rolehelper"
	"github.com/bhojpur/belaur/pkg/security/rbac"
)

type mockEchoEnforcer struct{}

func (m *mockEchoEnforcer) Enforce(username, method, path string, params map[string]string) error {
	if username == "enforcer-perms-err" {
		return rbac.NewErrPermissionDenied("namespace", "action", "thing")
	}
	if username == "enforcer-err" {
		return errors.New("error")
	}
	return nil
}

var mockRoleAuth = &AuthConfig{
	RoleCategories: []*belaur.UserRoleCategory{
		{
			Name: "CatOne",
			Roles: []*belaur.UserRole{
				{
					Name: "GetSingle",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						rolehelper.NewUserRoleEndpoint("GET", "/catone/:id"),
						rolehelper.NewUserRoleEndpoint("GET", "/catone/latest"),
					},
				},
				{
					Name: "PostSingle",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						rolehelper.NewUserRoleEndpoint("POST", "/catone"),
					},
				},
			},
		},
		{
			Name: "CatTwo",
			Roles: []*belaur.UserRole{
				{
					Name: "GetSingle",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						rolehelper.NewUserRoleEndpoint("GET", "/cattwo/:first/:second"),
					},
				},
				{
					Name: "PostSingle",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						rolehelper.NewUserRoleEndpoint("POST", "/cattwo/:first/:second/start"),
					},
				},
			},
		},
	},
	rbacEnforcer: &mockEchoEnforcer{},
}

func makeAuthBarrierRouter() *echo.Echo {
	e := echo.New()
	e.Use(authMiddleware(mockRoleAuth))

	success := func(c echo.Context) error {
		return c.NoContent(200)
	}

	e.GET("/auth", success)
	e.GET("/catone/:test", success)
	e.GET("/catone/latest", success)
	e.POST("/catone", success)
	e.POST("/enforcer/test", success)

	return e
}

func TestAuthBarrierNoToken(t *testing.T) {
	e := makeAuthBarrierRouter()

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected response code %v got %v", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthBarrierBadHeader(t *testing.T) {
	e := makeAuthBarrierRouter()

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "my-token")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected response code %v got %v", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthBarrierHMACTokenWithHMACKey(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}

	claims := belaur.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + belaur.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Belaur Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(belaur.Cfg.JWTKey)

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}
}

func TestAuthBarrierRSATokenWithRSAKey(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		belaur.Cfg = nil
	}()

	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	belaur.Cfg = &belaur.Config{
		JWTKey: key,
	}

	claims := belaur.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + belaur.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Belaur Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	tokenstring, _ := token.SignedString(belaur.Cfg.JWTKey)

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}
}

func TestAuthBarrierHMACTokenWithRSAKey(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		JWTKey: &rsa.PrivateKey{},
	}

	claims := belaur.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + belaur.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Belaur Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString([]byte("hmac-jwt-key"))

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected response code %v got %v", http.StatusUnauthorized, rec.Code)
	}

	bodyBytes, _ := ioutil.ReadAll(rec.Body)
	body := string(bodyBytes)

	signingMethodError := fmt.Sprintf("unexpected signing method: %v", token.Header["alg"])
	if body != signingMethodError {
		t.Fatalf("expected body '%v' got '%v'", signingMethodError, body)
	}
}

func TestAuthBarrierRSATokenWithHMACKey(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}

	claims := belaur.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + belaur.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Belaur Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	tokenstring, _ := token.SignedString(key)

	req := httptest.NewRequest(echo.GET, "/auth", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected response code %v got %v", http.StatusUnauthorized, rec.Code)
	}

	bodyBytes, _ := ioutil.ReadAll(rec.Body)
	body := string(bodyBytes)

	signingMethodError := fmt.Sprintf("unexpected signing method: %v", token.Header["alg"])
	if body != signingMethodError {
		t.Fatalf("expected body '%v' got '%v'", signingMethodError, body)
	}
}

var roleTests = []struct {
	perm   string
	method string
	path   string
}{
	{"CatOneGetSingle", "GET", "/catone/:id"},
	{"CatOneGetSingle", "GET", "/catone/latest"},
	{"CatOnePostSingle", "POST", "/catone"},
	{"CatTwoGetSingle", "POST", "/cattwo/:first/:second"},
	{"CatTwoPostSingle", "POST", "/cattwo/:first/:second/start"},
}

func TestAuthBarrierNoPerms(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}

	claims := belaur.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + belaur.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Belaur Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(belaur.Cfg.JWTKey)

	for _, tt := range roleTests {
		t.Run(tt.perm, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(echo.POST, "/catone", nil)
			req.Header.Set("Authorization", "Bearer "+tokenstring)
			e.ServeHTTP(rec, req)
			testPermFailed(t, tt.perm, rec.Code, rec.Body.String())
		})
	}
}

func TestAuthBarrierAllPerms(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}

	claims := belaur.JwtCustomClaims{
		Username: "test-user",
		Roles:    []string{"CatOneGetSingle", "CatOnePostSingle", "CatTwoGetSingle", "CatTwoPostSingle"},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + belaur.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Belaur Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(belaur.Cfg.JWTKey)

	for _, tt := range roleTests {
		t.Run(tt.perm, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(echo.POST, "/catone", nil)
			req.Header.Set("Authorization", "Bearer "+tokenstring)
			e.ServeHTTP(rec, req)
			testPermSuccess(t, rec.Code, rec.Body.String())
		})
	}
}

func testPermFailed(t *testing.T, perm string, statusCode int, body string) {
	if body == "" {
		t.Fatalf("expected response body %v got %v", "Permission denied for user "+perm+". Required permission "+perm, body)
	}
	if statusCode != http.StatusForbidden {
		t.Fatalf("expected response code %v got %v", http.StatusForbidden, statusCode)
	}
}

func testPermSuccess(t *testing.T, statusCode int, body string) {
	if body != "" {
		t.Fatalf("expected response body %v got %v", "", body)
	}
	if statusCode != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, statusCode)
	}
}

func Test_AuthMiddleware_Enforcer_PermissionDenied(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}

	claims := belaur.JwtCustomClaims{
		Username: "enforcer-perms-err",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + belaur.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Belaur Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(belaur.Cfg.JWTKey)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.POST, "/enforcer/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)
	e.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, http.StatusForbidden)
	assert.Equal(t, rec.Body.String(), "Permission denied. Must have namespace/action thing")
}

func Test_AuthMiddleware_Enforcer_UnknownError(t *testing.T) {
	e := makeAuthBarrierRouter()

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		JWTKey: []byte("hmac-jwt-key"),
	}
	belaur.Cfg.Logger = hclog.NewNullLogger()

	claims := belaur.JwtCustomClaims{
		Username: "enforcer-err",
		Roles:    []string{},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + belaur.JwtExpiry,
			IssuedAt:  time.Now().Unix(),
			Subject:   "Belaur Session Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstring, _ := token.SignedString(belaur.Cfg.JWTKey)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(echo.POST, "/enforcer/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenstring)
	e.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, http.StatusInternalServerError)
	assert.Equal(t, rec.Body.String(), "Unknown error has occurred while validating permissions.")
}
