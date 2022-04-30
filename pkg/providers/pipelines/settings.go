package pipelines

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

	belaur "github.com/bhojpur/belaur"
	"github.com/labstack/echo/v4"
)

// SettingsPollOn turn on polling functionality
// @Summary Turn on polling functionality.
// @Description Turns on the polling functionality for Bhojpur Belaur which periodically checks if there is new code to deploy for all pipelines.
// @Tags settings
// @Produce plain
// @Security ApiKeyAuth
// @Success 200 {string} string "Polling is turned on."
// @Failure 400 {string} string "Error while toggling poll setting."
// @Failure 500 {string} string "Internal server error while getting setting."
// @Router /settings/poll/on [post]
func (pp *PipelineProvider) SettingsPollOn(c echo.Context) error {
	settingsStore := pp.deps.SettingsStore

	configStore, err := settingsStore.SettingsGet()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Something went wrong while retrieving settings information.")
	}
	if configStore == nil {
		configStore = &belaur.StoreConfig{}
	}

	belaur.Cfg.Poll = true
	err = pp.deps.PipelineService.StartPoller()
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	configStore.Poll = true
	err = settingsStore.SettingsPut(configStore)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned on.")
}

// SettingsPollOff turn off polling functionality.
// @Summary Turn off polling functionality.
// @Description Turns off the polling functionality for Bhojpur Belaur which periodically checks if there is new code to deploy for all pipelines.
// @Tags settings
// @Produce plain
// @Security ApiKeyAuth
// @Success 200 {string} string "Polling is turned off."
// @Failure 400 {string} string "Error while toggling poll setting."
// @Failure 500 {string} string "Internal server error while getting setting."
// @Router /settings/poll/off [post]
func (pp *PipelineProvider) SettingsPollOff(c echo.Context) error {
	settingsStore := pp.deps.SettingsStore

	configStore, err := settingsStore.SettingsGet()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Something went wrong while retrieving settings information.")
	}
	if configStore == nil {
		configStore = &belaur.StoreConfig{}
	}
	belaur.Cfg.Poll = false
	err = pp.deps.PipelineService.StopPoller()
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	configStore.Poll = true
	err = settingsStore.SettingsPut(configStore)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.String(http.StatusOK, "Polling is turned off.")
}

type pollStatus struct {
	Status bool
}

// SettingsPollGet get status of polling functionality.
// @Summary Get the status of the poll setting.
// @Description Gets the status of the poll setting.
// @Tags settings
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} pollStatus "Poll status"
// @Failure 500 {string} string "Internal server error while getting setting."
// @Router /settings/poll [get]
func (pp *PipelineProvider) SettingsPollGet(c echo.Context) error {
	settingsStore := pp.deps.SettingsStore

	configStore, err := settingsStore.SettingsGet()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Something went wrong while retrieving settings information.")
	}
	var ps pollStatus
	if configStore == nil {
		ps.Status = belaur.Cfg.Poll
	} else {
		ps.Status = configStore.Poll
	}
	return c.JSON(http.StatusOK, ps)
}
