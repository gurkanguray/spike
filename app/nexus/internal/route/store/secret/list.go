//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"errors"
	"net/http"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// RouteListPaths handles requests to retrieve all available secret paths.
//
// This endpoint requires a valid admin JWT token for authentication.
// The function returns a list of all paths where secrets are stored, regardless
// of their version or deletion status.
//
// The function follows these steps:
//  1. Validates the JWT token
//  2. Validates the request body format
//  3. Retrieves all secret paths from the state
//  4. Returns the list of paths
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *log.AuditEntry for logging audit information
//
// Returns:
//   - error: if an error occurs during request processing.
//
// Request body format:
//
//	{} // Empty request body expected
//
// Response format on success (200 OK):
//
//	{
//	    "keys": []string   // Array of all secret paths
//	}
//
// Error responses:
//   - 401 Unauthorized: Invalid or missing JWT token
//   - 400 Bad Request: Invalid request body format
//
// All operations are logged using structured logging. This endpoint only
// returns the paths to secrets and not their contents; use RouteGetSecret to
// retrieve actual secret values.
func RouteListPaths(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeListPaths", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditList

	requestBody := net.ReadRequestBody(w, r)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.SecretListRequest, reqres.SecretListResponse](
		requestBody, w,
		reqres.SecretListResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	keys := state.ListKeys()

	responseBody := net.MarshalBody(reqres.SecretListResponse{Keys: keys}, w)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeListPaths", "msg", "OK")
	return nil
}
