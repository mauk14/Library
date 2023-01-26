package data

import (
	"github.com/google/uuid"
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
