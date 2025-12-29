// Package adapters provides concrete implementations of ports.
package adapters

import (
	"context"
)

// RepositoryImpl is an in-memory repository stub.
type RepositoryImpl struct {
	data map[string]interface{}
}

// NewRepositoryImpl returns a new in-memory repository.
func NewRepositoryImpl() *RepositoryImpl {
	return &RepositoryImpl{
		data: make(map[string]interface{}),
	}
}

// Save stores data in the repository.
func (r *RepositoryImpl) Save(_ context.Context, _ interface{}) error {
	return nil
}

// FindByID retrieves data by identifier.
func (r *RepositoryImpl) FindByID(_ context.Context, _ string) (interface{}, error) {
	return nil, nil
}
