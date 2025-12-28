package adapters

import (
	"context"
)

type RepositoryImpl struct {
	data map[string]interface{}
}

func NewRepositoryImpl() *RepositoryImpl {
	return &RepositoryImpl{
		data: make(map[string]interface{}),
	}
}

func (r *RepositoryImpl) Save(ctx context.Context, data interface{}) error {
	return nil
}

func (r *RepositoryImpl) FindByID(ctx context.Context, id string) (interface{}, error) {
	return nil, nil
}
