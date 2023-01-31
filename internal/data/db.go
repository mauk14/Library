package data

import (
	"context"
)

type DB interface {
	Insert(ctx context.Context, query string, collection string, data interface{}) error
	Get(ctx context.Context, query string, id interface{}, collection string) (interface{}, error)
	GetAll(ctx context.Context, query string, collection string, opt map[string]string, genres []string, filters Filters) ([]*Book, Metadata, error)
	Update(ctx context.Context, query string, data interface{}) error
	Delete(ctx context.Context, query string, id int64, collection string) error
	GetLastId(ctx context.Context, query string, collection string) (int64, error)
}
