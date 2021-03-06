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
	"net/http"

	"github.com/bhojpur/belaur/pkg/helper/stringhelper"

	"github.com/bhojpur/belaur/pkg/services"
	"github.com/labstack/echo/v4"
)

type addSecret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type updateSecret struct {
	Key   string `json:"key"`
	Value string `json:"newvalue"`
}

// CreateSecret creates a secret
// @Summary Create a secret.
// @Description Creates a secret.
// @Tags secrets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param secret body addSecret true "The secret to create"
// @Success 201 {string} string "secret successfully set"
// @Failure 400 {string} string "Error binding or key is reserved."
// @Failure 500 {string} string "Cannot get or load secrets"
// @Router /secret [post]
func CreateSecret(c echo.Context) error {
	var key, value string
	s := new(addSecret)
	err := c.Bind(s)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	key = s.Key
	value = s.Value

	return upsertSecret(c, key, value)
}

// UpdateSecret updates a given secret
// @Summary Update a secret.
// @Description Update a secret.
// @Tags secrets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param secret body updateSecret true "The secret to update with the new value"
// @Success 201 {string} string "secret successfully set"
// @Failure 400 {string} string "Error binding or key is reserved."
// @Failure 500 {string} string "Cannot get or load secrets"
func UpdateSecret(c echo.Context) error {
	var key, value string
	s := new(updateSecret)
	err := c.Bind(s)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	key = s.Key
	value = s.Value
	return upsertSecret(c, key, value)
}

// updates or creates a secret
func upsertSecret(c echo.Context, key string, value string) error {
	// Handle ignored special keys
	if stringhelper.IsContainedInSlice(ignoredVaultKeys, key, true) {
		return c.String(http.StatusBadRequest, "key is reserved and cannot be set/changed")
	}

	v, err := services.DefaultVaultService()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v.Add(key, []byte(value))
	err = v.SaveSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusCreated, "secret successfully set")
}

// ListSecrets retrieves all secrets from the vault.
// @Summary List all secrets.
// @Description Retrieves all secrets from the vault.
// @Tags secrets
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} addSecret "Secrets"
// @Failure 500 {string} string "Cannot get or load secrets"
// @Router /secrets [get]
func ListSecrets(c echo.Context) error {
	secrets := make([]addSecret, 0)
	v, err := services.DefaultVaultService()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	kvs := v.GetAll()
	for _, k := range kvs {
		// Handle ignored special keys
		if stringhelper.IsContainedInSlice(ignoredVaultKeys, k, true) {
			continue
		}

		s := addSecret{Key: k, Value: "**********"}
		secrets = append(secrets, s)
	}
	return c.JSON(http.StatusOK, secrets)
}

// RemoveSecret removes a secret from the vault.
// @Summary Removes a secret from the vault..
// @Description Removes a secret from the vault.
// @Tags secrets
// @Produce plain
// @Security ApiKeyAuth
// @Param key body string true "Key"
// @Success 200 {string} string "secret successfully deleted"
// @Failure 400 {string} string "key is reserved and cannot be deleted"
// @Failure 500 {string} string "Cannot get or load secrets"
// @Router /secret/:key [delete]
func RemoveSecret(c echo.Context) error {
	key := c.Param("key")
	if key == "" {
		return c.String(http.StatusBadRequest, "invalid key given")
	}

	// Handle ignored special keys
	if stringhelper.IsContainedInSlice(ignoredVaultKeys, key, true) {
		return c.String(http.StatusBadRequest, "key is reserved and cannot be deleted")
	}

	v, err := services.DefaultVaultService()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	err = v.LoadSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	v.Remove(key)
	err = v.SaveSecrets()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "secret successfully deleted")
}
