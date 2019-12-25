package main

import (
	"os"
	"io"
)

// Book is a structure that holds metadata about a book
// all metadata is saved as strings
type Book struct {
	ID            string
	Title         string
	Author        string
	Pages         string
	PublishedDate string
	ImageURL      string
	Description   string
}

// BookDatabase provides a thread safe interface
type BookDatabase interface {
	// takes a context variable (for canceling and deadlines)
	// returns list of books, ordered by title
	ListBooks(context.Context) ([]*Book, error)

	// get an entry as a reference
	GetBook(context.Context, id string) (*Book, error)

	// add entry
	AddBook(context.Context, b *Book) (id string, error)

	// delete entry
	DeleteBook(context.Context, id string) error

	// update entry
	UpdateBook(context.Context, b *Book) error

	// [the above to be added back in once actions are considered]
}

type Bookshelf struct {
	DB        BookDatabase
	logWriter io.Writer
	// potentially add in storage bucket from GCP
	// missing some kind of storage infrastructure
}

func NewBookshelf(db BookDatabase) (*Bookshelf, error) {
	ctx := context.Background()

	b := &Bookshelf{
		DB: db,
		logWriter: os.Stderr
	}
	return b, nil
}