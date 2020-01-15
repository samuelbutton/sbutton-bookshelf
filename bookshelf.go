package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/errorreporting"
	"cloud.google.com/go/storage"
)

// Book is a structure that holds metadata about a book
// all metadata is saved as strings
type Book struct {
	ID            *string
	Title         *string
	Author        *string
	Pages         *string
	PublishedDate *string
	ImageURL      *string
	Description   *string
}

// BookDatabase provides a thread safe interface
type BookDatabase interface {
	ListBooks(context.Context) ([]*Book, error)
	GetBook(ctx context.Context, id string) (*Book, error)
	AddBook(ctx context.Context, b *Book) (id string, err error)
	DeleteBook(ctx context.Context, id string) error
	UpdateBook(ctx context.Context, b *Book) error
	// GetUser(ctx context.Context, id string) (*Account, error)
	LoginAccount(ctx context.Context, email string, password string) (*Account, time.Time, error)
	ValidateAccount(ctx context.Context, a *Account) error
	CreateAccount(ctx context.Context, a *Account) (*Account, time.Time, error)
	GetResetToken(ctx context.Context, email string) (string, error)
	CheckTokenValidity(ctx context.Context, tokenString string, a *Account) error
	RemoveToken(ctx context.Context, a *Account) error
	UpdatePassword(ctx context.Context, a *Account) error
}

// Bookshelf with storage for book information (relational) and
// image (files in bucket)
type Bookshelf struct {
	DB                BookDatabase
	Messages          []string
	account           *Account
	userLoggedIn      bool
	StorageBucket     *storage.BucketHandle
	StorageBucketName string
	logWriter         io.Writer
	errorClient       *errorreporting.Client
}

// NewBookshelf creates new storage and structure
func NewBookshelf(db BookDatabase) (*Bookshelf, error) {
	ctx := context.Background()
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT must be set")
	}

	bucketName := projectID + "_bucket"
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	errorClient, err := errorreporting.NewClient(ctx, projectID, errorreporting.Config{
		ServiceName: "sbutton-bookshelf",
		OnError: func(err error) {
			fmt.Fprintf(os.Stderr, "Could not log error: %v", err)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("errorreporting.NewClient: %v", err)
	}
	b := &Bookshelf{
		logWriter:         os.Stderr,
		errorClient:       errorClient,
		DB:                db,
		Messages:          nil,
		userLoggedIn:      false,
		StorageBucketName: bucketName,
		StorageBucket:     storageClient.Bucket(bucketName),
	}
	return b, nil
}
