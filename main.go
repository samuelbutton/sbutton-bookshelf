package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/gofrs/uuid"
	// "github.com/gorilla/handlers"
	// "github.com/gorilla/mux"
)

// parse templates

func main() {
	// set port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// pull context
	// background returns a non-nill, empty Context, it is never canceled, has no values,
	// and has no deadline (used as the top level context in a main function processing
	// incoming requests)
	ctx := context.Background()

	// pull in new database
	db, err := sql.Open("mysql", "root:button16@/samTestSchema")
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}

	// close after surrounding function ends with defer
	defer db.Close()

	// verify that a connection can be made before making a query
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("db.Ping: %v", err)
	}

	// Prepare statement for inserting data
	stmtIns, err := db.PrepareContext(ctx, "INSERT INTO squareNum VALUES( ?, ? )") // ? = placeholder
	if err != nil {
		log.Fatalf("db.Prepare INSERT: %v", err)
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	// Prepare statement for reading data
	stmtOut, err := db.PrepareContext(ctx, "SELECT squareNumber FROM squarenum WHERE number = ?")
	if err != nil {
		log.Fatalf("db.Prepare SELECT: %v", err)
	}
	defer stmtOut.Close()

	// Insert square numbers for 0-24 in the database
	// for i := 25; i < 50; i++ {
	// 	_, err = stmtIns.Exec(i, (i * i)) // Insert tuples (i, i^2)
	// 	if err != nil {
	// 		log.Fatalf("db.Exec: %v", err)
	// 	}
	// }

	var squareNum int // we "scan" the result in here

	// Query the square-number of 13
	err = stmtOut.QueryRow(26).Scan(&squareNum) // WHERE number = 13
	if err != nil {
		log.Fatalf("db.Query13: %v", err)
	}
	fmt.Printf("The square number of 13 is: %d \n", squareNum)

	// Query another number.. 1 maybe?
	err = stmtOut.QueryRow(49).Scan(&squareNum) // WHERE number = 1
	if err != nil {
		log.Fatalf("db.Query1: %v", err)
	}
	fmt.Printf("The square number of 1 is: %d \n", squareNum)

	// create new object with use of database
	// the google original took in a projectID as well, possibly add authentication
	// b, err := NewBookshelf(db)
	// if err != nil {
	// 	log.Fatalf("NewBookshelf: %v", err)
	// }

	// b.registerHandlers()

	// log.Printf("Listening on localhost:%s", port)
	// if err := http.ListenAndServe(":"+port, nil); err != nil {
	// 	log.Fatal(err)
	// }
}

// func (b *Bookshelf) registerHandlers() {
// 	r := mux.NewRouter()

// 	r.Handle("/", http.RedirectHandler("/books", http.StatusFound))

// 	r.Methods("GET").Path("/books").
// 		Handler(appHandler(b.listHandler))
// 	r.Methods("GET").Path("/books/add").
// 		Handler(appHandler(b.addFormHandler))
// 	r.Methods("GET").Path("/books/{id:[0-9a-zA-Z_\\-]+}").
// 		Handler(appHandler(b.detailHandler))
// 	r.Methods("GET").Path("/books/{id:[0-9a-zA-Z_\\-]+}/edit").
// 		Handler(appHandler(b.editFormHandler))

// 	r.Methods("POST").Path("/books").
// 		Handler(appHandler(b.createHandler))
// 	r.Methods("POST", "PUT").Path("/books/{id:[0-9a-zA-Z_\\-]+}").
// 		Handler(appHandler(b.updateHandler))
// 	r.Methods("POST").Path("/books/{id:[0-9a-zA-Z_\\-]+}:delete").
// 		Handler(appHandler(b.deleteHandler)).Name("delete")

// 	// Respond to App Engine and Compute Engine health checks.
// 	// Indicate the server is healthy.
// 	r.Methods("GET").Path("/_ah/health").HandlerFunc(
// 		func(w http.ResponseWriter, r *http.Request) {
// 			w.Write([]byte("ok"))
// 		})

// 	r.Methods("GET").Path("/logs").Handler(appHandler(b.sendLog))
// 	r.Methods("GET").Path("/errors").Handler(appHandler(b.sendError))

// 	// Delegate all of the HTTP routing and serving to the gorilla/mux router.
// 	// Log all requests using the standard Apache format.
// 	http.Handle("/", handlers.CombinedLoggingHandler(b.logWriter, r))
// }

// type appHandler func(http.ResponseWriter, *http.Request) *appError

// type appError struct {
// 	err     error
// 	message string
// 	code    int
// 	req     *http.Request
// 	b       *Bookshelf
// 	stack   []byte
// }
