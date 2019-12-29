package main

import (
	"context"
	"log"
	"net/http"
	"os"
	// _ "github.com/go-sql-driver/mysql"
	// "github.com/gofrs/uuid"
)

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

	// db := newMemoryDB()
	db, err := newMySQLDB()
	if err != nil {
		log.Fatalf("newMySQLDB: %v", err)
	}

	// close after surrounding function ends with defer
	defer db.Close(ctx)

	// create new object with use of database (either mysql or persistence in memory database)
	b, err := NewBookshelf(db)
	if err != nil {
		log.Fatalf("NewBookshelf: %v", err)
	}

	// build a handler specific function
	b.registerHandlers()

	log.Printf("Listening on localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
