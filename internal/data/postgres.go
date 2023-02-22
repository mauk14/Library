package data

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	DB *pgxpool.Pool
}

func (m *Postgres) Insert(ctx context.Context, query string, collection string, data interface{}) error {
	if collection == "books" {
		book := data.(*Book)
		args := []any{book.Title, book.Author, book.Year, book.Size, book.Genres, book.Version}
		return m.DB.QueryRow(ctx, query, args).Scan(&book.ID)
	}
	return errors.New("No collections in database")
}
func (m *Postgres) Get(ctx context.Context, query string, id interface{}, collection string, scope string) (interface{}, error) {
	return nil, nil
}
func (m *Postgres) GetAll(ctx context.Context, query string, collection string, opt map[string]string, genres []string, filters Filters) ([]*Book, Metadata, error) {
	return nil, Metadata{}, nil
}
func (m *Postgres) Update(ctx context.Context, query string, data interface{}) error {
	return nil
}
func (m *Postgres) Delete(ctx context.Context, query string, id int64, collection string, scope string) error {
	return nil
}
func (m *Postgres) GetLastId(ctx context.Context, query string, collection string) (int64, error) {
	return 0, nil
}
