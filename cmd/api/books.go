package main

import (
	"fmt"
	"github.com/google/uuid"
	"mauk14.library/internal/data"
	"mauk14.library/internal/validator"
	"net/http"
	"time"
)

func (app *application) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string    `json:"title"`
		Author string    `json:"author"`
		Year   int32     `json:"year"`
		Size   data.Size `json:"size"`
		Genres []string  `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	book := &data.Book{
		Title:  input.Title,
		Year:   input.Year,
		Author: input.Author,
		Size:   input.Size,
		Genres: input.Genres,
	}

	v := validator.New()

	if data.ValidateBook(v, book); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)

}

func (app *application) showBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	book := data.Book{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "It",
		Size:      102,
		Author:    "Stiven Hoking",
		Genres:    []string{"horror"},
		Version:   uuid.New(),
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"book": book}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)

	}
}
