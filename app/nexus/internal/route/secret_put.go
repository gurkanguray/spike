package route

//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

import (
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// routePutSecret handles HTTP requests to create or update secrets at a
// specified path.
//
// This endpoint requires a valid admin JWT token for authentication. It accepts
// a PUT request with a JSON body containing the secret path and values to store.
// The function performs an upsert operation, creating a new secret if it
// doesn't exist or updating an existing one.
//
// Request body format:
//
//	{
//	    "path": string,           // Path where the secret should be stored
//	    "values": map[string]any // Key-value pairs representing the secret data
//	}
//
// Responses:
//   - 200 OK: Secret successfully created or updated
//   - 400 Bad Request: Invalid request body or parameters
//   - 401 Unauthorized: Invalid or missing JWT token
//
// The function logs its progress at various stages using structured logging.
func routePutSecret(w http.ResponseWriter, r *http.Request) {
	log.Log().Info("routeGetSecret", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)

	validJwt := net.ValidateJwt(w, r, state.AdminToken())
	if !validJwt {
		return
	}

	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return
	}

	request := net.HandleRequest[
		reqres.SecretPutRequest, reqres.SecretPutResponse](
		requestBody, w,
		reqres.SecretPutResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return
	}

	values := request.Values
	path := request.Path

	state.UpsertSecret(path, values)
	log.Log().Info("routePutSecret", "msg", "Secret upserted")

	responseBody := net.MarshalBody(reqres.SecretPutResponse{}, w)
	if responseBody == nil {
		return
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routePutSecret", "msg", "OK")
}
