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
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/rolehelper"
	"github.com/bhojpur/belaur/pkg/security/rbac"
)

var (
	// errNotAuthorized is thrown when user wants to access resource which is protected
	errNotAuthorized = errors.New("no or invalid jwt token provided. You are not authorized")
)

func authMiddleware(authCfg *AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := getToken(c)
			if err != nil {
				return c.String(http.StatusUnauthorized, err.Error())
			}

			// Validate token
			if claims, okClaims := token.Claims.(jwt.MapClaims); okClaims && token.Valid {
				// All ok, continue
				username, hasUsername := claims["username"]
				roles, hasRoles := claims["roles"]
				if hasUsername && hasRoles && roles != nil {
					// Look through the perms until we find that the user has this permission
					if err := authCfg.checkRole(roles, c.Request().Method, c.Path()); err != nil {
						return c.String(http.StatusForbidden, "Permission denied for user.")
					}

					username, okUsername := username.(string)
					if !okUsername {
						belaur.Cfg.Logger.Error("username is not type string")
						return c.String(http.StatusInternalServerError, "Unknown error has occurred.")
					}

					// Currently this lives inside the existing auth middleware. Ideally we would have independent
					// middleware for enforcing RBAC. For now I will leave this here so we avoid parsing the token
					// and claims multiple times.
					params := map[string]string{}
					for i, n := range c.ParamNames() {
						params[n] = c.ParamValues()[i]
					}
					err := authCfg.rbacEnforcer.Enforce(username, c.Request().Method, c.Path(), params)
					if err != nil {
						if _, permDenied := err.(*rbac.ErrPermissionDenied); permDenied {
							return c.String(http.StatusForbidden, err.Error())
						}
						belaur.Cfg.Logger.Error("rbacEnforcer error", "error", err.Error())
						return c.String(http.StatusInternalServerError, "Unknown error has occurred while validating permissions.")
					}
				}
				return next(c)
			}
			return c.String(http.StatusUnauthorized, errNotAuthorized.Error())
		}
	}
}

// AuthConfig is a simple config struct to be passed into AuthMiddleware. Currently allows the ability to specify
// the permission roles required for each echo endpoint.
type AuthConfig struct {
	RoleCategories []*belaur.UserRoleCategory
	rbacEnforcer   rbac.EndpointEnforcer
}

// Finds the required role for the metho & path specified. If it exists we validate that the provided user roles have
// the permission role. If not, error specifying the required role.
func (ra *AuthConfig) checkRole(userRoles interface{}, method, path string) error {
	perm := ra.getRequiredRole(method, path)
	if perm == "" {
		return nil
	}
	for _, role := range userRoles.([]interface{}) {
		if role.(string) == perm {
			return nil
		}
	}
	return fmt.Errorf("required permission role %s", perm)
}

// Iterate over each category to find a permission (if existing) for this API endpoint.
func (ra *AuthConfig) getRequiredRole(method, path string) string {
	for _, category := range ra.RoleCategories {
		for _, role := range category.Roles {
			for _, endpoint := range role.APIEndpoint {
				// If the http method & path match then return the role required for this endpoint
				if method == endpoint.Method && path == endpoint.Path {
					return rolehelper.FullUserRoleName(category, role)
				}
			}
		}
	}
	return ""
}

// Get the JWT token from the echo context
func getToken(c echo.Context) (*jwt.Token, error) {
	// Get the token
	jwtRaw := c.Request().Header.Get("Authorization")
	split := strings.Split(jwtRaw, " ")
	if len(split) != 2 {
		return nil, errNotAuthorized
	}
	jwtString := split[1]

	// Parse token
	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		signingMethodError := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		switch token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			if _, ok := belaur.Cfg.JWTKey.([]byte); !ok {
				return nil, signingMethodError
			}
			return belaur.Cfg.JWTKey, nil
		case *jwt.SigningMethodRSA:
			if _, ok := belaur.Cfg.JWTKey.(*rsa.PrivateKey); !ok {
				return nil, signingMethodError
			}
			return belaur.Cfg.JWTKey.(*rsa.PrivateKey).Public(), nil
		default:
			return nil, signingMethodError
		}
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}
