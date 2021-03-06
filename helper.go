package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
)

// UseString is safe way to use a string given a pointer to a string
func UseString(s *string) string {
	if s == nil {
		temp := ""
		s = &temp
	}
	return *s
}

// UsePointer allows for construction of pointer in one line
func UsePointer(s string) *string {
	return &s
}

func (b *Bookshelf) bookFromRequest(r *http.Request) (*Book, error) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	if id == "" {
		return nil, errors.New("no book with empty ID")
	}
	book, err := b.DB.GetBook(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not find book: %v", err)
	}
	return book, nil
}

func (b *Bookshelf) setCookieAndRedirect(w http.ResponseWriter, r *http.Request, a *Account, t time.Time) error {
	if a.ID > 0 {
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   UseString(a.Token),
			Expires: t,
		})
		b.userLoggedIn = true
		// probably should set the bookshelf to have a property pointing to the account for further use
		http.Redirect(w, r, "/books", http.StatusFound)
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
		return errors.New("account unsuccessful")
	}
	return nil
}

func (b *Bookshelf) removeCookie(w http.ResponseWriter, r *http.Request) error {
	c := http.Cookie{
		Name:   "token",
		MaxAge: -1}
	http.SetCookie(w, &c)

	cth, err := r.Cookie("token")
	fmt.Println(cth.Expires)
	fmt.Println(err)
	b.userLoggedIn = false
	return nil
}

func (b *Bookshelf) bookFromForm(r *http.Request) (*Book, error) {
	ctx := r.Context()
	imageURL, err := b.uploadFileFromForm(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("could not upload file: %v", err)
	}
	if imageURL == "" {
		imageURL = r.FormValue("imageURL")
	}

	book := &Book{
		Title:         UsePointer(r.FormValue("title")),
		Author:        UsePointer(r.FormValue("author")),
		Pages:         UsePointer(r.FormValue("pages")),
		PublishedDate: UsePointer(r.FormValue("publishedDate")),
		ImageURL:      UsePointer(imageURL),
		Description:   UsePointer(r.FormValue("description")),
	}

	return book, nil
}

func (b *Bookshelf) accountFromForm(r *http.Request) (*Account, error) {
	account := &Account{
		Email:    UsePointer(r.FormValue("email")),
		Password: UsePointer(r.FormValue("password")),
	}
	return account, nil
}

func (b *Bookshelf) tokenFromForm(r *http.Request) (string, error) {
	return r.FormValue("token"), nil
}

func (b *Bookshelf) emailFromForm(r *http.Request) (string, error) {
	return r.FormValue("email"), nil
}

func (b *Bookshelf) checkForToken(r *http.Request) (string, error) {
	tokenString := mux.Vars(r)["token"]
	if tokenString == "" {
		return "", errors.New("cannot reset with no identification")
	}
	return tokenString, nil
}

func (b *Bookshelf) uploadFileFromForm(ctx context.Context, r *http.Request) (url string, err error) {
	f, fh, err := r.FormFile("image")
	if err == http.ErrMissingFile {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	if b.StorageBucket == nil {
		return "", errors.New("storage bucket is missing: check bookshelf.go")
	}
	if _, err := b.StorageBucket.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			return "", fmt.Errorf("bucket %q does not exist: check bookshelf.go", b.StorageBucketName)
		}
		return "", fmt.Errorf("could not get bucket: %v", err)
	}

	name := uuid.Must(uuid.NewV4()).String() + path.Ext(fh.Filename)

	w := b.StorageBucket.Object(name).NewWriter(ctx)

	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.ContentType = fh.Header.Get("Content-Type")

	w.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(w, f); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	const publicURL = "https://storage.googleapis.com/%s/%s"
	return fmt.Sprintf(publicURL, b.StorageBucketName, name), nil
}
