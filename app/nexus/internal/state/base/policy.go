//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/spiffe/spike/internal/entity/data"
)

var (
	ErrPolicyNotFound = errors.New("policy not found")
	ErrPolicyExists   = errors.New("policy already exists")
	ErrInvalidPolicy  = errors.New("invalid policy")
)

// CreatePolicy creates a new policy with an auto-generated ID.
func CreatePolicy(policy data.Policy) (data.Policy, error) {
	if policy.Name == "" {
		return data.Policy{}, ErrInvalidPolicy
	}

	// Generate ID and set creation time
	policy.Id = uuid.New().String()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}

	policies.Store(policy.Id, policy)
	return policy, nil
}

// GetPolicy retrieves a policy by ID. Returns ErrPolicyNotFound if the policy
// doesn't exist.
func GetPolicy(id string) (data.Policy, error) {
	if value, exists := policies.Load(id); exists {
		return value.(data.Policy), nil
	}
	return data.Policy{}, ErrPolicyNotFound
}

// UpdatePolicy updates an existing policy. Returns ErrPolicyNotFound if
// the policy doesn't exist.
func UpdatePolicy(policy data.Policy) error {
	if policy.Id == "" || policy.Name == "" {
		return ErrInvalidPolicy
	}

	// Check if policy exists
	if _, exists := policies.Load(policy.Id); !exists {
		return ErrPolicyNotFound
	}

	// Preserve original creation timestamp and creator
	original, _ := GetPolicy(policy.Id)
	policy.CreatedAt = original.CreatedAt
	policy.CreatedBy = original.CreatedBy

	policies.Store(policy.Id, policy)
	return nil
}

// DeletePolicy removes a policy by ID. Returns ErrPolicyNotFound if the policy
// doesn't exist.
func DeletePolicy(id string) error {
	if _, exists := policies.Load(id); !exists {
		return ErrPolicyNotFound
	}

	policies.Delete(id)
	return nil
}

// ListPolicies returns all policies as a slice.
func ListPolicies() []data.Policy {
	var result []data.Policy

	policies.Range(func(key, value interface{}) bool {
		result = append(result, value.(data.Policy))
		return true
	})

	return result
}

// ListPoliciesByPath returns all policies that match a given path pattern.
func ListPoliciesByPath(pathPattern string) []data.Policy {
	var result []data.Policy

	policies.Range(func(key, value interface{}) bool {
		policy := value.(data.Policy)
		if policy.PathPattern == pathPattern {
			result = append(result, policy)
		}
		return true
	})

	return result
}

// ListPoliciesBySpiffeId returns all policies that match a given SPIFFE ID pattern.
func ListPoliciesBySpiffeId(spiffeIdPattern string) []data.Policy {
	var result []data.Policy

	policies.Range(func(key, value interface{}) bool {
		policy := value.(data.Policy)
		if policy.SpiffeIdPattern == spiffeIdPattern {
			result = append(result, policy)
		}
		return true
	})

	return result
}
