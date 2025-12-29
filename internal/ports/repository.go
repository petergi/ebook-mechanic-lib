package ports

import "context"

// Repository defines a basic persistence port.
type Repository interface {
	Save(ctx context.Context, data interface{}) error
	FindByID(ctx context.Context, id string) (interface{}, error)
}
