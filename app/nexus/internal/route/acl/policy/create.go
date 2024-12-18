//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"
	"net/http"
	"time"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/data"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// RoutePutPolicy handles HTTP PUT requests for creating new policies.
// It processes the request body to create a policy with the specified name,
// SPIFFE ID pattern, path pattern, and permissions.
//
// The function expects a JSON request body containing:
//   - Name: policy name
//   - SpiffeIdPattern: SPIFFE ID matching pattern
//   - PathPattern: path matching pattern
//   - Permissions: set of allowed permissions
//
// On success, it returns a JSON response with the created policy's ID.
// On failure, it returns an appropriate error response with status code.
//
// Parameters:
//   - w: HTTP response writer for sending the response
//   - r: HTTP request containing the policy creation data
//   - audit: Audit entry for logging the policy creation action
//
// Returns:
//   - error: nil on successful policy creation, error otherwise
//
// Example request body:
//
//	{
//	    "name": "example-policy",
//	    "spiffe_id_pattern": "spiffe://example.org/*/service",
//	    "path_pattern": "/api/*",
//	    "permissions": ["read", "write"]
//	}
//
// Example success response:
//
//	{
//	    "id": "policy-123"
//	}
//
// Example error response:
//
//	{
//	    "err": "Internal server error"
//	}
func RoutePutPolicy(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routePutPolicy", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditCreate

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.PolicyCreateRequest, reqres.PolicyCreateResponse](
		requestBody, w,
		reqres.PolicyCreateResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	// TODO: sanitize

	name := request.Name
	spiffeIdPattern := request.SpiffeIdPattern
	pathPattern := request.PathPattern
	permissions := request.Permissions

	policy, err := state.CreatePolicy(data.Policy{
		Id:              "",
		Name:            name,
		SpiffeIdPattern: spiffeIdPattern,
		PathPattern:     pathPattern,
		Permissions:     permissions,
		CreatedAt:       time.Time{},
		CreatedBy:       "",
	})
	if err != nil {
		log.Log().Info("routePutPolicy",
			"msg", "Failed to create policy", "err", err)

		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: "Internal server error",
		}, w)

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Error("routePutPolicy", "msg", "internal server error")

		return err
	}

	responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
		Id: policy.Id,
	}, w)

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routePutPolicy", "msg", "OK")

	return nil
}
