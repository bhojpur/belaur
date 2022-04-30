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

	"github.com/labstack/echo/v4"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/store"
)

const msgSomethingWentWrong = "Something went wrong while retrieving settings information."

type settingsHandler struct {
	store store.SettingsStore
}

func newSettingsHandler(store store.SettingsStore) *settingsHandler {
	return &settingsHandler{store: store}
}

type rbacPutRequest struct {
	Enabled bool `json:"enabled"`
}

// @Summary Put RBAC settings
// @Description Save the given RBAC settings.
// @Tags settings
// @Accept json
// @Produce plain
// @Security ApiKeyAuth
// @Param RbacPutRequest body rbacPutRequest true "RBAC setting details."
// @Success 200 {string} string "Settings have been updated."
// @Failure 400 {string} string "Invalid body."
// @Failure 500 {string} string "Something went wrong while saving or retrieving rbac settings."
// @Router /settings/rbac [put]
func (h *settingsHandler) rbacPut(c echo.Context) error {
	var request rbacPutRequest
	if err := c.Bind(&request); err != nil {
		belaur.Cfg.Logger.Error("failed to bind body", "error", err.Error())
		return c.String(http.StatusBadRequest, "Invalid body provided.")
	}

	settings, err := h.store.SettingsGet()
	if err != nil {
		belaur.Cfg.Logger.Error("failed to get store settings", "error", err.Error())
		return c.String(http.StatusInternalServerError, msgSomethingWentWrong)
	}

	settings.RBACEnabled = request.Enabled

	if err := h.store.SettingsPut(settings); err != nil {
		belaur.Cfg.Logger.Error("failed to put store settings", "error", err.Error())
		return c.String(http.StatusInternalServerError, "An error occurred while saving the settings.")
	}

	return c.String(http.StatusOK, "Settings have been updated.")
}

type rbacGetResponse struct {
	Enabled bool `json:"enabled"`
}

// @Summary Get RBAC settings
// @Description Get the given RBAC settings.
// @Tags settings
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} rbacGetResponse
// @Failure 500 {string} string "Something went wrong while saving or retrieving rbac settings."
// @Router /settings/rbac [get]
func (h *settingsHandler) rbacGet(c echo.Context) error {
	settings, err := h.store.SettingsGet()
	if err != nil {
		belaur.Cfg.Logger.Error("failed to get store settings", "error", err.Error())
		return c.String(http.StatusInternalServerError, msgSomethingWentWrong)
	}

	response := rbacGetResponse{}
	// If RBAC is applied via config it takes priority.
	if belaur.Cfg.RBACEnabled {
		response.Enabled = true
	} else {
		response.Enabled = settings.RBACEnabled
	}

	return c.JSON(http.StatusOK, rbacGetResponse{Enabled: settings.RBACEnabled})
}
