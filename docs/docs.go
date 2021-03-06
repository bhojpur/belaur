package docs

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
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "termsOfService": "https://github.com/bhojpur/belaur/blob/master/LICENSE",
        "contact": {
            "name": "API Support",
            "url": "https://github.com/bhojpur/belaur"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/login": {
            "post": {
                "description": "Returns an authenticated user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "User Login",
                "parameters": [
                    {
                        "description": "UserLogin request",
                        "name": "UserLoginRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/belaur.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/belaur.User"
                        }
                    },
                    "400": {
                        "description": "error reading json",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "403": {
                        "description": "credentials provided",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "{creating jwt token|signing jwt token}",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/permission": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns a list of all the roles available.",
                "tags": [
                    "rbac"
                ],
                "summary": "Returns a list of default roles.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/belaur.UserRoleCategory"
                            }
                        }
                    }
                }
            }
        },
        "/pipeline": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Starts creating a pipeline given all the data asynchronously.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Create pipeline.",
                "parameters": [
                    {
                        "description": "Create pipeline details",
                        "name": "CreatePipelineRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/belaur.CreatePipeline"
                        }
                    }
                ],
                "responses": {
                    "200": {},
                    "400": {
                        "description": "Failed to bind, validation error and invalid details",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal error while saving create pipeline run",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipeline/created": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get a list of all pipelines which are about to be compiled and which have been compiled.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Get all create pipelines.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/belaur.CreatePipeline"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal error while retrieving create pipeline data.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipeline/githook": {
            "post": {
                "description": "This is the global endpoint which will handle all github webhook callbacks.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Handle github webhook callbacks.",
                "parameters": [
                    {
                        "description": "A github webhook payload",
                        "name": "payload",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/pipelines.Payload"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "successfully processed event",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bind error and schedule errors",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Various internal errors running and triggering and rebuilding pipelines. Please check the logs for more information.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipeline/gitlsremote": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Checks for available git remote branches which in turn verifies repository access.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Check for repository access.",
                "parameters": [
                    {
                        "description": "The repository details",
                        "name": "PipelineGitLSRemoteRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/belaur.GitRepo"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Available branches",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Failed to bind body",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "403": {
                        "description": "No access",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipeline/latest": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns the latest of all registered pipelines included with the latest run.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Returns the latest run.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/pipelines.getAllWithLatestRun"
                        }
                    },
                    "500": {
                        "description": "Internal error while getting latest run",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipeline/name": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns all registered pipelines.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Returns all registered pipelines.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/belaur.Pipeline"
                            }
                        }
                    }
                }
            }
        },
        "/pipeline/periodicschedules": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns the latest of all registered pipelines included with the latest run.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Returns the latest run.",
                "parameters": [
                    {
                        "description": "A list of valid cronjob specs",
                        "name": "schedules",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {},
                    "400": {
                        "description": "Bind error and schedule errors",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipeline/{pipelineid}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get pipeline information based on ID.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Get pipeline information.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The ID of the pipeline",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/belaur.Pipeline"
                        }
                    },
                    "400": {
                        "description": "The given pipeline id is not valid",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Pipeline not found with the given id",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Update a pipeline by its ID.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Update pipeline.",
                "parameters": [
                    {
                        "description": "PipelineUpdate request",
                        "name": "PipelineUpdateRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/belaur.Pipeline"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Pipeline has been updated",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Error while updating the pipeline",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "The pipeline with the given ID was not found",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal error while updating and building the new pipeline information",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Accepts a pipeline id and deletes it from the store. It also removes the binary inside the pipeline folder.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Delete a pipeline.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The ID of the pipeline.",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Pipeline has been deleted",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Error while deleting the pipeline",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "The pipeline with the given ID was not found",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal error while deleting and removing the pipeline from store and disk",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipeline/{pipelineid}/pull": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Pull new code using the repository of the pipeline.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Update the underlying repository of the pipeline.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The ID of the pipeline.",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {},
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipeline/{pipelineid}/reset-trigger-token": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Generates a new remote trigger token for a given pipeline.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Reset trigger token.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The ID of the pipeline.",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Trigger successful for pipeline: {pipelinename}",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid pipeline id",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Pipeline not found",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal storage error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipeline/{pipelineid}/start": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Starts a pipeline with a given ID and arguments for that pipeline and returns created/scheduled status.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Start a pipeline.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The ID of the pipeline.",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    },
                    {
                        "description": "Optional arguments of the pipeline.",
                        "name": "args",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/belaur.Argument"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/belaur.PipelineRun"
                        }
                    },
                    "400": {
                        "description": "Various failures regarding starting the pipeline like: invalid id, invalid docker value and schedule errors",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Pipeline not found",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipeline/{pipelineid}/{pipelinetoken}/trigger": {
            "post": {
                "description": "Using a trigger token, start a pipeline run. This endpoint does not require authentication.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "pipelines"
                ],
                "summary": "Trigger a pipeline.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The ID of the pipeline.",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "The trigger token for this pipeline.",
                        "name": "pipelinetoken",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Trigger successful for pipeline: {pipelinename}",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Error while triggering pipeline",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "403": {
                        "description": "Invalid trigger token",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipelinerun/{pipelineid}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns all runs about the given pipeline.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelinerun"
                ],
                "summary": "Get all pipeline runs.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID of the pipeline",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "a list of pipeline runes",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/belaur.PipelineRun"
                            }
                        }
                    },
                    "400": {
                        "description": "Invalid pipeline id",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Error retrieving all pipeline runs.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipelinerun/{pipelineid}/latest": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns the latest run of a pipeline, given by id.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelinerun"
                ],
                "summary": "Get latest pipeline runs.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID of the pipeline",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "the latest pipeline run",
                        "schema": {
                            "$ref": "#/definitions/belaur.PipelineRun"
                        }
                    },
                    "400": {
                        "description": "Invalid pipeline id",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "error getting latest run or cannot read pipeline run log file",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipelinerun/{pipelineid}/{runid}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns details about a specific pipeline run.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelinerun"
                ],
                "summary": "Get Pipeline run.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID of the pipeline",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "ID of the pipeline run",
                        "name": "runid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/belaur.PipelineRun"
                        }
                    },
                    "400": {
                        "description": "Invalid pipeline or pipeline not found.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Pipeline Run not found.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Something went wrong while getting pipeline run.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipelinerun/{pipelineid}/{runid}/log": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns logs from a pipeline run.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pipelinerun"
                ],
                "summary": "Get logs for pipeline run.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID of the pipeline",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "ID of the run",
                        "name": "runid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "logs",
                        "schema": {
                            "$ref": "#/definitions/pipelines.jobLogs"
                        }
                    },
                    "400": {
                        "description": "Invalid pipeline id or run id or pipeline not found",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "cannot read pipeline run log file",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/pipelinerun/{pipelineid}/{runid}/stop": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Stops a pipeline run.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "pipelinerun"
                ],
                "summary": "Stop a pipeline run.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID of the pipeline",
                        "name": "pipelineid",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "ID of the pipeline run",
                        "name": "runid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "pipeline successfully stopped",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid pipeline id or run id",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Pipeline Run not found.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/rbac/roles": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Gets all RBAC roles.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "rbac"
                ],
                "summary": "Gets all RBAC roles.",
                "responses": {
                    "200": {
                        "description": "All the roles.",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    }
                }
            }
        },
        "/rbac/roles/{role}": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Adds an RBAC role using the RBAC service.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "rbac"
                ],
                "summary": "Adds an RBAC role.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Name of the role",
                        "name": "role",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Role created successfully.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Must provide role.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "An error occurred while adding the role.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "description": "Deletes an RBAC role using the RBAC service.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "rbac"
                ],
                "summary": "Delete an RBAC role.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The name of the rule",
                        "name": "role",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Role deleted successfully.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Must provide role.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "An error occurred while deleting the role.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/rbac/roles/{role}/attach/{username}": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Attach role to user.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "rbac"
                ],
                "summary": "Attach role to user.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The role",
                        "name": "role",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "The username of the user",
                        "name": "username",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Role attached successfully.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Must provide role or username.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "An error occurred while attaching the role.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Detach role to user.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "rbac"
                ],
                "summary": "Detach role to user.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The role",
                        "name": "role",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "The username of the user",
                        "name": "username",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Role detached successfully.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Must provide role or username.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "An error occurred while detaching the role.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/rbac/roles/{role}/attached": {
            "get": {
                "description": "Gets a user attached to a role.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "rbac"
                ],
                "summary": "Gets a user attached to a role.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The role for the user",
                        "name": "role",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Attached users for the role",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Must provide role.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "An error occurred while getting the user.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/secret": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Creates a secret.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "secrets"
                ],
                "summary": "Create a secret.",
                "parameters": [
                    {
                        "description": "The secret to create",
                        "name": "secret",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.addSecret"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "secret successfully set",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Error binding or key is reserved.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Cannot get or load secrets",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/secret/:key": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Removes a secret from the vault.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "secrets"
                ],
                "summary": "Removes a secret from the vault..",
                "parameters": [
                    {
                        "description": "Key",
                        "name": "key",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "secret successfully deleted",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "key is reserved and cannot be deleted",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Cannot get or load secrets",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/secrets": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieves all secrets from the vault.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "secrets"
                ],
                "summary": "List all secrets.",
                "responses": {
                    "200": {
                        "description": "Secrets",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/handlers.addSecret"
                            }
                        }
                    },
                    "500": {
                        "description": "Cannot get or load secrets",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/settings/poll": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Gets the status of the poll setting.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "settings"
                ],
                "summary": "Get the status of the poll setting.",
                "responses": {
                    "200": {
                        "description": "Poll status",
                        "schema": {
                            "$ref": "#/definitions/pipelines.pollStatus"
                        }
                    },
                    "500": {
                        "description": "Internal server error while getting setting.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/settings/poll/off": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Turns off the polling functionality for Bhojpur Belaur which periodically checks if there is new code to deploy for all pipelines.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "settings"
                ],
                "summary": "Turn off polling functionality.",
                "responses": {
                    "200": {
                        "description": "Polling is turned off.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Error while toggling poll setting.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal server error while getting setting.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/settings/poll/on": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Turns on the polling functionality for Bhojpur Belaur which periodically checks if there is new code to deploy for all pipelines.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "settings"
                ],
                "summary": "Turn on polling functionality.",
                "responses": {
                    "200": {
                        "description": "Polling is turned on.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Error while toggling poll setting.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal server error while getting setting.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/settings/rbac": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get the given RBAC settings.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "settings"
                ],
                "summary": "Get RBAC settings",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlers.rbacGetResponse"
                        }
                    },
                    "500": {
                        "description": "Something went wrong while saving or retrieving rbac settings.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Save the given RBAC settings.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "settings"
                ],
                "summary": "Put RBAC settings",
                "parameters": [
                    {
                        "description": "RBAC setting details.",
                        "name": "RbacPutRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/handlers.rbacPutRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Settings have been updated.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid body.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Something went wrong while saving or retrieving rbac settings.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Adds a new user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Add user.",
                "parameters": [
                    {
                        "description": "UserAdd request",
                        "name": "UserAddRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/belaur.User"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "User has been added",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid parameters given for add user request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "{User put failed|User permission put error}",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/password": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Changes the password of the given user.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Change password for user.",
                "parameters": [
                    {
                        "description": "UserChangePassword request",
                        "name": "UserChangePasswordRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.changePasswordRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "{Invalid parameters given for password change request|Cannot find user with the given username|New password does not match new password confirmation}",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "412": {
                        "description": "Precondition Failed",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/{username}/delete": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Delete a given user.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Delete user.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The username to delete",
                        "name": "username",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "{Invalid username given|Auto user cannot be deleted}",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "{User not found|Permission not found|Rbac not found}",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/{username}/permissions": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get permissions of the user.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Get permission of the user.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The username to get permission for",
                        "name": "username",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/belaur.UserPermission"
                        }
                    },
                    "400": {
                        "description": "Failed to get permission",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Adds or updates permissions for a user..",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Adds or updates permissions for a user.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The username to get permission for",
                        "name": "username",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Permissions have been updated",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "{Invalid parameters given for request|Permissions put failed}",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/user/{username}/reset-trigger-token": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Generates and saves a new remote trigger token for a given user.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Generate new remote trigger token.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The username to reset the token for",
                        "name": "username",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "{Invalid username given|Only auto user can have a token reset|User not found|Error retrieving user}",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/users": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns a list of registered users.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "users"
                ],
                "summary": "Get all users",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/belaur.User"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/users/{username}/rbac/roles": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Gets all roles for a user.",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "rbac"
                ],
                "summary": "Gets all roles for a user.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The username of the user",
                        "name": "username",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Attached roles to a user",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Must provide username.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "An error occurred while getting the roles.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/worker": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Gets all workers.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "workers"
                ],
                "summary": "Get all workers.",
                "responses": {
                    "200": {
                        "description": "A list of workers.",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/belaur.Worker"
                            }
                        }
                    },
                    "500": {
                        "description": "Cannot get memdb service from service store.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/worker/register": {
            "post": {
                "description": "Allows new workers to register themself at this Bhojpur Belaur instance.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "workers"
                ],
                "summary": "Register a new worker.",
                "parameters": [
                    {
                        "description": "Worker details",
                        "name": "RegisterWorkerRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/workers.registerWorker"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Details of the registered worker.",
                        "schema": {
                            "$ref": "#/definitions/workers.registerResponse"
                        }
                    },
                    "400": {
                        "description": "Invalid arguments of the worker.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "403": {
                        "description": "Wrong global worker secret provided.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Various internal services like, certs, vault and generating new secrets.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/worker/secret": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns the global secret for registering new worker.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "workers"
                ],
                "summary": "Get worker register secret.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Cannot get worker secret from vault.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Generates a new global worker registration secret.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "workers"
                ],
                "summary": "Reset worker register secret.",
                "responses": {
                    "200": {
                        "description": "global worker registration secret has been successfully reset",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Vault related internal problems.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/worker/status": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Returns general status information about all workers.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "workers"
                ],
                "summary": "Get worker status overview.",
                "responses": {
                    "200": {
                        "description": "The worker status overview response.",
                        "schema": {
                            "$ref": "#/definitions/workers.workerStatusOverviewResponse"
                        }
                    },
                    "500": {
                        "description": "Cannot get memdb service from service store.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/worker/{workerid}": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Deregister an existing worker.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "workers"
                ],
                "summary": "Deregister and existing worker.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "The id of the worker to deregister.",
                        "name": "workerid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Worker has been successfully deregistered.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Worker id is missing or worker not registered.",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Cannot get memdb service from service store or failed to delete worker.",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "belaur.Argument": {
            "type": "object",
            "properties": {
                "desc": {
                    "type": "string"
                },
                "key": {
                    "type": "string"
                },
                "type": {
                    "type": "string"
                },
                "value": {
                    "type": "string"
                }
            }
        },
        "belaur.CreatePipeline": {
            "type": "object",
            "properties": {
                "created": {
                    "type": "string"
                },
                "githubtoken": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "output": {
                    "type": "string"
                },
                "pipeline": {
                    "type": "object",
                    "$ref": "#/definitions/belaur.Pipeline"
                },
                "status": {
                    "type": "integer"
                },
                "statustype": {
                    "type": "string"
                }
            }
        },
        "belaur.GitRepo": {
            "type": "object",
            "properties": {
                "branches": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "password": {
                    "type": "string"
                },
                "privatekey": {
                    "type": "object",
                    "$ref": "#/definitions/belaur.PrivateKey"
                },
                "selectedbranch": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                },
                "user": {
                    "type": "string"
                }
            }
        },
        "belaur.Job": {
            "type": "object",
            "properties": {
                "args": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/belaur.Argument"
                    }
                },
                "dependson": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/belaur.Job"
                    }
                },
                "desc": {
                    "type": "string"
                },
                "failpipeline": {
                    "type": "boolean"
                },
                "id": {
                    "type": "integer"
                },
                "status": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "belaur.Pipeline": {
            "type": "object",
            "properties": {
                "created": {
                    "type": "string"
                },
                "docker": {
                    "type": "boolean"
                },
                "execpath": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "jobs": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/belaur.Job"
                    }
                },
                "name": {
                    "type": "string"
                },
                "notvalid": {
                    "type": "boolean"
                },
                "periodicschedules": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "repo": {
                    "type": "object",
                    "$ref": "#/definitions/belaur.GitRepo"
                },
                "sha256sum": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "trigger_token": {
                    "type": "string"
                },
                "type": {
                    "type": "string"
                },
                "uuid": {
                    "type": "string"
                }
            }
        },
        "belaur.PipelineRun": {
            "type": "object",
            "properties": {
                "docker": {
                    "type": "boolean"
                },
                "dockerworkerid": {
                    "type": "string"
                },
                "finishdate": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "jobs": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/belaur.Job"
                    }
                },
                "pipelineid": {
                    "type": "integer"
                },
                "pipelinetags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "pipelinetype": {
                    "type": "string"
                },
                "scheduledate": {
                    "type": "string"
                },
                "startdate": {
                    "type": "string"
                },
                "started_reason": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                },
                "uniqueid": {
                    "type": "string"
                }
            }
        },
        "belaur.PrivateKey": {
            "type": "object",
            "properties": {
                "key": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "belaur.User": {
            "type": "object",
            "properties": {
                "display_name": {
                    "type": "string"
                },
                "jwtexpiry": {
                    "type": "integer"
                },
                "lastlogin": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                },
                "tokenstring": {
                    "type": "string"
                },
                "trigger_token": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "belaur.UserPermission": {
            "type": "object",
            "properties": {
                "groups": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "roles": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "belaur.UserRole": {
            "type": "object",
            "properties": {
                "api_endpoints": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/belaur.UserRoleEndpoint"
                    }
                },
                "description": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "belaur.UserRoleCategory": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "roles": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/belaur.UserRole"
                    }
                }
            }
        },
        "belaur.UserRoleEndpoint": {
            "type": "object",
            "properties": {
                "method": {
                    "type": "string"
                },
                "path": {
                    "type": "string"
                }
            }
        },
        "belaur.Worker": {
            "type": "object",
            "properties": {
                "finishedruns": {
                    "type": "integer"
                },
                "lastcontact": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "registerdate": {
                    "type": "string"
                },
                "slots": {
                    "type": "integer"
                },
                "status": {
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "uniqueid": {
                    "type": "string"
                }
            }
        },
        "handlers.addSecret": {
            "type": "object",
            "properties": {
                "key": {
                    "type": "string"
                },
                "value": {
                    "type": "string"
                }
            }
        },
        "handlers.rbacGetResponse": {
            "type": "object",
            "properties": {
                "enabled": {
                    "type": "boolean"
                }
            }
        },
        "handlers.rbacPutRequest": {
            "type": "object",
            "properties": {
                "enabled": {
                    "type": "boolean"
                }
            }
        },
        "handlers.updateSecret": {
            "type": "object",
            "properties": {
                "key": {
                    "type": "string"
                },
                "newvalue": {
                    "type": "string"
                }
            }
        },
        "pipelines.Payload": {
            "type": "object",
            "properties": {
                "repository": {
                    "type": "object",
                    "$ref": "#/definitions/pipelines.Repository"
                }
            }
        },
        "pipelines.Repository": {
            "type": "object",
            "properties": {
                "git_url": {
                    "type": "string"
                },
                "html_url": {
                    "type": "string"
                },
                "ssh_url": {
                    "type": "string"
                }
            }
        },
        "pipelines.getAllWithLatestRun": {
            "type": "object",
            "properties": {
                "p": {
                    "type": "object",
                    "$ref": "#/definitions/belaur.Pipeline"
                },
                "r": {
                    "type": "object",
                    "$ref": "#/definitions/belaur.PipelineRun"
                }
            }
        },
        "pipelines.jobLogs": {
            "type": "object",
            "properties": {
                "finished": {
                    "type": "boolean"
                },
                "log": {
                    "type": "string"
                }
            }
        },
        "pipelines.pollStatus": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "boolean"
                }
            }
        },
        "user.changePasswordRequest": {
            "type": "object",
            "properties": {
                "newpassword": {
                    "type": "string"
                },
                "newpasswordconf": {
                    "type": "string"
                },
                "oldpassword": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "workers.registerResponse": {
            "type": "object",
            "properties": {
                "cacert": {
                    "type": "string"
                },
                "cert": {
                    "type": "string"
                },
                "key": {
                    "type": "string"
                },
                "uniqueid": {
                    "type": "string"
                }
            }
        },
        "workers.registerWorker": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "secret": {
                    "type": "string"
                },
                "tags": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "workers.workerStatusOverviewResponse": {
            "type": "object",
            "properties": {
                "activeworker": {
                    "type": "integer"
                },
                "finishedruns": {
                    "type": "integer"
                },
                "inactiveworker": {
                    "type": "integer"
                },
                "queuesize": {
                    "type": "integer"
                },
                "suspendedworker": {
                    "type": "integer"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "",
	BasePath:    "/api/v1",
	Schemes:     []string{},
	Title:       "Belaur API",
	Description: "This is the API that the Bhojpur Belaur Admin UI uses.",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
