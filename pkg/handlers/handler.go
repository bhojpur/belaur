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

	echoSwagger "github.com/swaggo/echo-swagger"

	rice "github.com/GeertJohan/go.rice"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/rolehelper"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	// List of secret keys which cannot be modified via the normal Vault API.
	ignoredVaultKeys []string
)

// InitHandlers initializes(registers) all handlers.
func (s *BelaurHandler) InitHandlers(e *echo.Echo) error {
	// Define prefix
	p := "/api/" + belaur.APIVersion + "/"

	// --- Register handlers at echo instance ---

	// API router group.
	apiGrp := e.Group(p)

	// API router group with auth middleware.
	apiAuthGrp := e.Group(p, authMiddleware(&AuthConfig{
		RoleCategories: rolehelper.DefaultUserRoles,
		rbacEnforcer:   s.deps.RBACService,
	}))

	// Endpoints for Bhojpur Belaur primary instance
	if belaur.Cfg.Mode == belaur.ModeServer {
		apiGrp.POST("login", s.deps.UserProvider.UserLogin)
		apiAuthGrp.GET("users", s.deps.UserProvider.UserGetAll)
		apiAuthGrp.POST("user/password", s.deps.UserProvider.UserChangePassword)
		apiAuthGrp.DELETE("user/:username", s.deps.UserProvider.UserDelete)
		apiAuthGrp.GET("user/:username/permissions", s.deps.UserProvider.UserGetPermissions)
		apiAuthGrp.PUT("user/:username/permissions", s.deps.UserProvider.UserPutPermissions)
		apiAuthGrp.POST("user", s.deps.UserProvider.UserAdd)
		apiAuthGrp.PUT("user/:username/reset-trigger-token", s.deps.UserProvider.UserResetTriggerToken)
		apiAuthGrp.GET("permission", PermissionGetAll)

		// Pipelines
		// Create pipeline provider
		apiAuthGrp.POST("pipeline", s.deps.PipelineProvider.CreatePipeline)
		apiAuthGrp.POST("pipeline/gitlsremote", s.deps.PipelineProvider.PipelineGitLSRemote)
		apiAuthGrp.GET("pipeline/name", s.deps.PipelineProvider.PipelineNameAvailable)
		apiAuthGrp.GET("pipeline/created", s.deps.PipelineProvider.CreatePipelineGetAll)
		apiAuthGrp.GET("pipeline", s.deps.PipelineProvider.PipelineGetAll)
		apiAuthGrp.GET("pipeline/:pipelineid", s.deps.PipelineProvider.PipelineGet)
		apiAuthGrp.PUT("pipeline/:pipelineid", s.deps.PipelineProvider.PipelineUpdate)
		apiAuthGrp.DELETE("pipeline/:pipelineid", s.deps.PipelineProvider.PipelineDelete)
		apiAuthGrp.POST("pipeline/:pipelineid/start", s.deps.PipelineProvider.PipelineStart)
		apiAuthGrp.PUT("pipeline/:pipelineid/reset-trigger-token", s.deps.PipelineProvider.PipelineResetToken)
		apiAuthGrp.POST("pipeline/:pipelineid/pull", s.deps.PipelineProvider.PipelinePull)
		apiAuthGrp.GET("pipeline/latest", s.deps.PipelineProvider.PipelineGetAllWithLatestRun)
		apiAuthGrp.POST("pipeline/periodicschedules", s.deps.PipelineProvider.PipelineCheckPeriodicSchedules)
		apiGrp.POST("pipeline/githook", s.deps.PipelineProvider.GitWebHook)
		apiGrp.POST("pipeline/:pipelineid/:pipelinetoken/trigger", s.deps.PipelineProvider.PipelineTrigger)

		// Settings
		settingsHandler := newSettingsHandler(s.deps.Store)
		apiAuthGrp.POST("settings/poll/on", s.deps.PipelineProvider.SettingsPollOn)
		apiAuthGrp.POST("settings/poll/off", s.deps.PipelineProvider.SettingsPollOff)
		apiAuthGrp.GET("settings/poll", s.deps.PipelineProvider.SettingsPollGet)
		apiAuthGrp.GET("settings/rbac", settingsHandler.rbacGet)
		apiAuthGrp.PUT("settings/rbac", settingsHandler.rbacPut)

		// PipelineRun
		apiAuthGrp.POST("pipelinerun/:pipelineid/:runid/stop", s.deps.PipelineProvider.PipelineStop)
		apiAuthGrp.GET("pipelinerun/:pipelineid/:runid", s.deps.PipelineProvider.PipelineRunGet)
		apiAuthGrp.GET("pipelinerun/:pipelineid", s.deps.PipelineProvider.PipelineGetAllRuns)
		apiAuthGrp.GET("pipelinerun/:pipelineid/latest", s.deps.PipelineProvider.PipelineGetLatestRun)
		apiAuthGrp.GET("pipelinerun/:pipelineid/:runid/log", s.deps.PipelineProvider.GetJobLogs)

		// Secrets
		apiAuthGrp.GET("secrets", ListSecrets)
		apiAuthGrp.DELETE("secret/:key", RemoveSecret)
		apiAuthGrp.POST("secret", CreateSecret)
		apiAuthGrp.PUT("secret/update", UpdateSecret)

		// RBAC - Management
		apiAuthGrp.GET("rbac/roles", s.deps.RBACProvider.GetAllRoles)
		apiAuthGrp.PUT("rbac/roles/:role", s.deps.RBACProvider.AddRole)
		apiAuthGrp.DELETE("rbac/roles/:role", s.deps.RBACProvider.DeleteRole)
		apiAuthGrp.PUT("rbac/roles/:role/attach/:username", s.deps.RBACProvider.AttachRole)
		apiAuthGrp.DELETE("rbac/roles/:role/attach/:username", s.deps.RBACProvider.DetachRole)
		apiAuthGrp.GET("rbac/roles/:role/attached", s.deps.RBACProvider.GetRoleAttachedUsers)
		// RBAC - Users
		apiAuthGrp.GET("users/:username/rbac/roles", s.deps.RBACProvider.GetUserAttachedRoles)

		// Swagger
		apiGrp.GET("swagger/*", echoSwagger.WrapHandler)
	}

	// Worker
	apiAuthGrp.GET("worker/secret", s.deps.WorkerProvider.GetWorkerRegisterSecret)
	apiAuthGrp.GET("worker/status", s.deps.WorkerProvider.GetWorkerStatusOverview)
	apiAuthGrp.GET("worker", s.deps.WorkerProvider.GetWorker)
	apiAuthGrp.DELETE("worker/:workerid", s.deps.WorkerProvider.DeregisterWorker)
	apiAuthGrp.POST("worker/secret", s.deps.WorkerProvider.ResetWorkerRegisterSecret)
	apiGrp.POST("worker/register", s.deps.WorkerProvider.RegisterWorker)

	// Middleware
	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())
	e.Use(middleware.BodyLimit("32M"))

	// Extra options
	e.HideBanner = true

	// Are we in production mode?
	if !belaur.Cfg.DevMode {
		staticAssets, err := rice.FindBox("../webui/dist")
		if err != nil {
			belaur.Cfg.Logger.Error("Cannot find assets in production mode.")
			return err
		}

		// Register handler for static assets
		assetHandler := http.FileServer(staticAssets.HTTPBox())
		e.GET("/", echo.WrapHandler(assetHandler))
		e.GET("/favicon.ico", echo.WrapHandler(assetHandler))
		e.GET("/css/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/js/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/fonts/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
		e.GET("/img/*", echo.WrapHandler(http.StripPrefix("/", assetHandler)))
	}

	// Setup ignored vault keys which cannot be modified directly via the Vault API
	ignoredVaultKeys = make([]string, 0, 1)
	ignoredVaultKeys = append(ignoredVaultKeys, belaur.WorkerRegisterKey)

	return nil
}
