package main

import (
	"context"
	"fmt"
	"io"
	"os"

	// "cloud.google.com/go/errorreporting"
	"cloud.google.com/go/storage"
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
	ListBooks(context.Context) ([]*Book, error)
	GetBook(context.Context, id string) (*Book, error)
	AddBook(context.Context, b *Book) (id string, error)
	DeleteBook(context.Context, id string) error
	UpdateBook(context.Context, b *Book) error
}

type Bookshelf struct {
	DB        BookDatabase
	StorageBucket     *storage.BucketHandle
	StorageBucketName string
	logWriter io.Writer
	// errorClient *errorreporting.Client
}

func NewBookshelf(db BookDatabase) (*Bookshelf, error) {
	ctx := context.Background()

	bucketName := projectID + "_bucket"
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	// errorClient, err := errorreporting.NewClient(ctx, projectID, errorreporting.Config{
	// 	ServiceName: "bookshelf", // need to do something about this, add in some kind of service on GCP
	// 	OnError: func(err error) {
	// 		fmt.Fprintf(os.Stderr, "Could not log error: %v", err)
	// 	},
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("errorreporting.NewClient: %v", err)
	// }
	b := &Bookshelf{
		logWriter:         os.Stderr,
		// errorClient:       errorClient,
		DB:                db,
		StorageBucketName: bucketName,
		StorageBucket:     storageClient.Bucket(bucketName),
	}
	return b, nil
}