package rolehelper

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

import belaur "github.com/bhojpur/belaur"

// NewUserRoleEndpoint is a constructor for creating new UserRoleEndpoints.
func NewUserRoleEndpoint(method string, path string) *belaur.UserRoleEndpoint {
	return &belaur.UserRoleEndpoint{Path: path, Method: method}
}

// FullUserRoleName returns a full user role name in the form {category}{role}.
func FullUserRoleName(category *belaur.UserRoleCategory, role *belaur.UserRole) string {
	return category.Name + role.Name
}

// FlattenUserCategoryRoles flattens the given user categories into a single slice with items in the form off
// {category}{role}s.
func FlattenUserCategoryRoles(cats []*belaur.UserRoleCategory) []string {
	var roles []string
	for _, category := range cats {
		for _, r := range category.Roles {
			roles = append(roles, FullUserRoleName(category, r))
		}
	}
	return roles
}

var (
	// DefaultUserRoles contains all the default user categories and roles.
	DefaultUserRoles = []*belaur.UserRoleCategory{
		{
			Name:        "Pipeline",
			Description: "Managing and initiating pipelines.",
			Roles: []*belaur.UserRole{
				{
					Name: "Create",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipeline"),
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/gitlsremote"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/name"),
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/githook"),
					},
					Description: "Create new pipelines.",
				},
				{
					Name: "List",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/created"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline"),
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/latest"),
					},
					Description: "List created pipelines.",
				},
				{
					Name: "Get",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipeline/:pipelineid"),
					},
					Description: "Get created pipelines.",
				},
				{
					Name: "Update",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/pipeline/:pipelineid"),
					},
					Description: "Update created pipelines.",
				},
				{
					Name: "Delete",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/pipeline/:pipelineid"),
					},
					Description: "Delete created pipelines.",
				},
				{
					Name: "Start",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipeline/:pipelineid/start"),
					},
					Description: "Start created pipelines.",
				},
			},
		},
		{
			Name:        "PipelineRun",
			Description: "Managing of pipeline runs.",
			Roles: []*belaur.UserRole{
				{
					Name: "Stop",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/pipelinerun/:pipelineid/:runid/stop"),
					},
					Description: "Stop running pipelines.",
				},
				{
					Name: "Get",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/:runid"),
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/latest"),
					},
					Description: "Get pipeline runs.",
				},
				{
					Name: "List",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "pipelinerun/:pipelineid"),
					},
					Description: "List pipeline runs.",
				},
				{
					Name: "Logs",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/pipelinerun/:pipelineid/:runid/latest"),
					},
					Description: "Get logs for pipeline runs.",
				},
			},
		},
		{
			Name:        "Secret",
			Description: "Managing of stored secrets used within pipelines.",
			Roles: []*belaur.UserRole{
				{
					Name: "List",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/secrets"),
					},
					Description: "List created secrets.",
				},
				{
					Name: "Delete",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/secret/:key"),
					},
					Description: "Delete created secrets.",
				},
				{
					Name: "Create",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/secret"),
					},
					Description: "Create new secrets.",
				},
				{
					Name: "Update",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/secret/update"),
					},
					Description: "Update created secrets.",
				},
			},
		},
		{
			Name:        "User",
			Description: "Managing of users.",
			Roles: []*belaur.UserRole{
				{
					Name: "Create",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/user"),
					},
					Description: "Create new users.",
				},
				{
					Name: "List",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/users"),
					},
					Description: "List created users.",
				},
				{
					Name: "ChangePassword",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/user/password"),
					},
					Description: "Change created users passwords.",
				},
				{
					Name: "Delete",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/user/:username"),
					},
					Description: "Delete created users.",
				},
			},
		},
		{
			Name:        "UserPermission",
			Description: "Managing of user permissions.",
			Roles: []*belaur.UserRole{
				{
					Name: "Get",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/user/:username/permissions"),
					},
					Description: "Get created users permissions.",
				},
				{
					Name: "Update",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("PUT", "/api/v1/user/:username/permissions"),
					},
					Description: "Update created users permissions.",
				},
			},
		},
		{
			Name:        "Worker",
			Description: "Managing of worker permissions.",
			Roles: []*belaur.UserRole{
				{
					Name: "GetRegistrationSecret",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/worker/secret"),
					},
					Description: "Get global worker registration secret.",
				},
				{
					Name: "GetOverview",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/worker/status"),
					},
					Description: "Get status overview of all workers.",
				},
				{
					Name: "GetWorker",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("GET", "/api/v1/worker"),
					},
					Description: "Get all worker for the worker overview table.",
				},
				{
					Name: "DeregisterWorker",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("DELETE", "/api/v1/worker/:workerid"),
					},
					Description: "Deregister a worker from the Bhojpur Belaur primary instance.",
				},
				{
					Name: "ResetWorkerRegisterSecret",
					APIEndpoint: []*belaur.UserRoleEndpoint{
						NewUserRoleEndpoint("POST", "/api/v1/worker/secret"),
					},
					Description: "Reset the global worker registration secret.",
				},
			},
		},
	}
)
