package ports

import "context"

type Repository interface {
	Save(ctx context.Context, data interface{}) error
	FindByID(ctx context.Context, id string) (interface{}, error)
}
