package data

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"mauk14.library/internal/validator"
	"time"
)

type Book struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Year      int32     `json:"year,omitempty"`
	Size      Size      `json:"-"`
	Genres    []string  `json:"genres,omitempty"`
	Version   uuid.UUID `json:"version"`
}

type BookModel struct {
	DB DB
}

func (b *BookModel) Insert(book *Book) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	id, err := b.DB.GetLastId(ctx, "", "books")

	if err != nil {
		return err
	}

	book.ID = id + 1
	book.Version = uuid.New()
	book.CreatedAt = time.Now()

	return b.DB.Insert(ctx, "", "books", book)

}

func (m *BookModel) Get(id int64) (*Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var book Book
	var result interface{}

	result, err := m.DB.Get(ctx, "", id, "books")
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	switch result.(type) {
	case Book:
		book = result.(Book)
	}

	return &book, nil

}

func (m *BookModel) Update(book *Book) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.Update(ctx, "", book)

	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m *BookModel) Delete(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if id < 1 {
		return ErrRecordNotFound
	}

	err := m.DB.Delete(ctx, "", id, "books")
	return err
}

func (m *BookModel) GetAll(title string, author string, genres []string, filters Filters) ([]*Book, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	opt := make(map[string]string)
	opt["title"] = title
	opt["author"] = author

	return m.DB.GetAll(ctx, "", "books", opt, genres, filters)

}

func ValidateBook(v *validator.Validator, Book *Book) {
	v.Check(Book.Title != "", "title", "must be provided")
	v.Check(len(Book.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(Book.Author != "", "author", "must be provided")
	v.Check(len(Book.Author) <= 500, "author", "must not be more than 500 bytes long")

	v.Check(Book.Year != 0, "year", "must be provided")
	v.Check(Book.Year >= 1888, "year", "must be greater than 1888")
	v.Check(Book.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(Book.Size != 0, "size", "must be provided")
	v.Check(Book.Size > 0, "size", "must be a positive integer")

	v.Check(Book.Genres != nil, "genres", "must be provided")
	v.Check(len(Book.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(Book.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(Book.Genres), "genres", "must not contain duplicate values")
}
