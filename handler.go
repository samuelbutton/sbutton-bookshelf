package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"runtime/debug"

	"cloud.google.com/go/storage"
	"github.com/gofrs/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	// See template.go for parseTemplate and execute (below)
	listTmpl   = parseTemplate("list.html")
	editTmpl   = parseTemplate("edit.html")
	detailTmpl = parseTemplate("detail.html")
)

// method off of the object because it's based on pointers
func (b *Bookshelf) registerHandlers() {
	r := mux.NewRouter()

	r.Handle("/", http.RedirectHandler("/books", http.StatusFound))

	r.Methods("GET").Path("/books").
		Handler(appHandler(b.listHandler))
	r.Methods("GET").Path("/books/add").
		Handler(appHandler(b.addFormHandler))
	r.Methods("GET").Path("/books/{id:[0-9a-zA-Z_\\-]+}").
		Handler(appHandler(b.detailHandler))
	r.Methods("GET").Path("/books/{id:[0-9a-zA-Z_\\-]+}/edit").
		Handler(appHandler(b.editFormHandler))

	r.Methods("POST").Path("/books").
		Handler(appHandler(b.createHandler))
	r.Methods("POST", "PUT").Path("/books/{id:[0-9a-zA-Z_\\-]+}").
		Handler(appHandler(b.updateHandler))
	r.Methods("POST").Path("/books/{id:[0-9a-zA-Z_\\-]+}:delete").
		Handler(appHandler(b.deleteHandler)).Name("delete")

	r.Methods("GET").Path("/_ah/health").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

	r.Methods("GET").Path("/logs").Handler(appHandler(b.sendLog))
	r.Methods("GET").Path("/errors").Handler(appHandler(b.sendError))

	http.Handle("/", handlers.CombinedLoggingHandler(b.logWriter, r))
}

// displays a list with summaries of books in the database
func (b *Bookshelf) listHandler(w http.ResponseWriter, r *http.Request) *appError {
	// pull in context from request
	// always non-nil, defaults to background context
	// if the request is cancelled, then everything related cancels as well
	ctx := r.Context()
	books, err := b.DB.ListBooks(ctx)
	if err != nil {
		return b.appErrorf(r, err, "could not list books: %v", err)
	}
	return listTmpl.Execute(b, w, r, books)
}

func (b *Bookshelf) detailHandler(w http.ResponseWriter, r *http.Request) *appError {
	book, err := b.bookFromRequest(r)
	if err != nil {
		return b.appErrorf(r, err, "%v", err)
	}
	return detailTmpl.Execute(b, w, r, book)
}

func (b *Bookshelf) addFormHandler(w http.ResponseWriter, r *http.Request) *appError {
	return editTmpl.Execute(b, w, r, nil)
}

func (b *Bookshelf) editFormHandler(w http.ResponseWriter, r *http.Request) *appError {
	book, err := b.bookFromRequest(r)
	if err != nil {
		return b.appErrorf(r, err, "%v", err)
	}
	return editTmpl.Execute(b, w, r, book)
}

// add a book to the database
func (b *Bookshelf) createHandler(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	book, err := b.bookFromForm(r)
	if err != nil {
		return b.appErrorf(r, err, "could not parse book from form: %w", err)
	}
	id, err := b.DB.AddBook(ctx, book)
	if err != nil {
		return b.appErrorf(r, err, "could not save book: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/books/%q", id), http.StatusFound)
	return nil
}

func (b *Bookshelf) updateHandler(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	if id == "" {
		return b.appErrorf(r, errors.New("no book with empty ID"), "no book with empty ID")
	}
	book, err := b.bookFromForm(r)
	if err != nil {
		return b.appErrorf(r, err, "could not parse book from form: %v", err)
	}
	book.ID = UsePointer(id)

	if err := b.DB.UpdateBook(ctx, book); err != nil {
		return b.appErrorf(r, err, "UpdateBook: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/books/%s", UseString(book.ID)), http.StatusFound)
	return nil
}

func (b *Bookshelf) deleteHandler(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	if err := b.DB.DeleteBook(ctx, id); err != nil {
		return b.appErrorf(r, err, "DeleteBook Error: %v", err)
	}
	http.Redirect(w, r, "/books", http.StatusFound)
	return nil
}

// sendLog logs a message.
//
// See https://cloud.google.com/logging/docs/setup/go for how to use the
// Stackdriver logging client. Output to stdout and stderr is automaticaly
// sent to Stackdriver when running on App Engine.
func (b *Bookshelf) sendLog(w http.ResponseWriter, r *http.Request) *appError {
	fmt.Fprintln(b.logWriter, "Hey, you triggered a custom log entry. Good job!")

	fmt.Fprintln(w, `<html>Log sent! Check the <a href="http://console.cloud.google.com/logs">logging section of the Cloud Console</a>.</html>`)

	return nil
}

// sendError triggers an error that is sent to Error Reporting.
func (b *Bookshelf) sendError(w http.ResponseWriter, r *http.Request) *appError {
	msg := `<html>Logging an error. Check <a href="http://console.cloud.google.com/errors">Error Reporting</a> (it may take a minute or two for the error to appear).</html>`
	err := errors.New("uh oh! an error occurred")
	return b.appErrorf(r, err, msg)
}

// we create the type "appHandler" to re-use write / request functions
// an "appHandler" is defined as a function that takes a writer, request and returns error
type appHandler func(http.ResponseWriter, *http.Request) *appError

// our error specific to the app is defined by an  error, message, code, request,
// bookshelf and stack ([]byte)
type appError struct {
	err     error
	message string
	code    int
	req     *http.Request
	b       *Bookshelf
	stack   []byte
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		fmt.Fprintf(e.b.logWriter, "Handler error (reported to Error Reporting): status code: %d, message: %s, underlying err: %+v\n", e.code, e.message, e.err)
		w.WriteHeader(e.code)
		fmt.Fprint(w, e.message)

		// e.b.errorClient.Report(errorreporting.Entry{
		// 	Error: e.err,
		// 	Req:   r,
		// 	Stack: e.stack,
		// })
		// e.b.errorClient.Flush()
	}
}

// specific error method for our application
// constructs an error to pass back to our multiplexor
// basically, like format app Error, but could be known as build app error
func (b *Bookshelf) appErrorf(r *http.Request, err error, format string, v ...interface{}) *appError {
	return &appError{
		err:     err,
		message: fmt.Sprintf(format, v...),
		code:    500,
		req:     r,
		b:       b,
		stack:   debug.Stack(),
	}
}

// bookFromRequest retrieves a book from the database given a book ID in the
// URL's path.
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

// bookFromForm populates the fields of a Book from form values
// (see templates/edit.html).
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

// uploadFileFromForm uploads a file if it's present in the "image" form field.
func (b *Bookshelf) uploadFileFromForm(ctx context.Context, r *http.Request) (url string, err error) {
	f, fh, err := r.FormFile("image")
	if err == http.ErrMissingFile {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	// storage bucket stuff, probably some mdb stuff instead of google cloud storage
	if b.StorageBucket == nil {
		return "", errors.New("storage bucket is missing: check bookshelf.go")
	}
	if _, err := b.StorageBucket.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			return "", fmt.Errorf("bucket %q does not exist: check bookshelf.go", b.StorageBucketName)
		}
		return "", fmt.Errorf("could not get bucket: %v", err)
	}

	// random filename, retaining existing extension.
	name := uuid.Must(uuid.NewV4()).String() + path.Ext(fh.Filename)

	w := b.StorageBucket.Object(name).NewWriter(ctx)

	// Warning: storage.AllUsers gives public read access to anyone.
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.ContentType = fh.Header.Get("Content-Type")

	// Entries are immutable, be aggressive about caching (1 day).
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
