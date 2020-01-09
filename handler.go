package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// StaticDir accessed for static files
const StaticDir = "/assets/"

var (
	listTmpl   = parseTemplate("list.html")
	editTmpl   = parseTemplate("edit.html")
	detailTmpl = parseTemplate("detail.html")
	loginTmpl  = parseTemplate("login.html")
	createTmpl = parseTemplate("create.html")
)

func (b *Bookshelf) registerHandlers() {
	r := mux.NewRouter()
	r.Use(JwtAuthentication)

	r.PathPrefix(StaticDir).Handler(http.StripPrefix(StaticDir, http.FileServer(http.Dir("."+StaticDir))))
	r.Handle("/", http.RedirectHandler("/login", http.StatusFound))

	r.Methods("GET").Path("/new").
		Handler(appHandler(b.createAccountFormHandler))
	r.Methods("GET").Path("/login").
		Handler(appHandler(b.loginFormHandler))
	r.Methods("GET").Path("/books").
		Handler(appHandler(b.listHandler))
	r.Methods("GET").Path("/books/add").
		Handler(appHandler(b.addFormHandler))
	r.Methods("GET").Path("/books/{id:[0-9a-zA-Z_\\-]+}").
		Handler(appHandler(b.detailHandler))
	r.Methods("GET").Path("/books/{id:[0-9a-zA-Z_\\-]+}/edit").
		Handler(appHandler(b.editFormHandler))

	r.Methods("POST").Path("/new").
		Handler(appHandler(b.createAccountHandler))
	r.Methods("POST").Path("/login").
		Handler(appHandler(b.authenticateAccountHandler))
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

func (b *Bookshelf) listHandler(w http.ResponseWriter, r *http.Request) *appError {
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

func (b *Bookshelf) createAccountFormHandler(w http.ResponseWriter, r *http.Request) *appError {
	return createTmpl.Execute(b, w, r, nil)
}

func (b *Bookshelf) loginFormHandler(w http.ResponseWriter, r *http.Request) *appError {
	return loginTmpl.Execute(b, w, r, nil)
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

func (b *Bookshelf) createHandler(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	book, err := b.bookFromForm(r)
	if err != nil {
		return b.appErrorf(r, err, "could not parse book from form: %v", err)
	}
	id, err := b.DB.AddBook(ctx, book)
	if err != nil {
		return b.appErrorf(r, err, "could not save book: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/books/%s", id), http.StatusFound)
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

func (b *Bookshelf) createAccountHandler(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	// a.Email = UsePointer(mux.Vars(r)["email"]) // REST
	// a.Password = UsePointer(mux.Vars(r)["password"]) // REST
	// err := json.NewDecoder(r.Body).Decode(account) // REST

	a, err := b.accountFromForm(r)
	if err != nil {
		return b.appErrorf(r, err, "could not parse account from form: %v", err)
	}
	id, err := b.DB.CreateAccount(ctx, a)
	if err != nil {
		return b.appErrorf(r, err, "could not creat account: %v", err)
	}
	if id > 0 {
		http.Redirect(w, r, "/books", http.StatusFound)
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	return nil
}

func (b *Bookshelf) authenticateAccountHandler(w http.ResponseWriter, r *http.Request) *appError {
	ctx := r.Context()
	a, err := b.accountFromForm(r)

	// a.Email = UsePointer(mux.Vars(r)["email"])
	// a.Password = UsePointer(mux.Vars(r)["password"])
	// err := json.NewDecoder(r.Body).Decode(account) // REST
	id, err := b.DB.LoginAccount(ctx, UseString(a.Email), UseString(a.Password))
	if err != nil {
		return b.appErrorf(r, err, "LoginAccount Error: %v", err)
	}
	if id > 0 {
		http.Redirect(w, r, "/books", http.StatusFound)
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	return nil
}
