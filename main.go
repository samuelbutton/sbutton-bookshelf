package main

import (
	"context"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	ctx := context.Background()

	db, err := newpostgresDB()
	if err != nil {
		log.Fatalf("newpostgresDB: %v", err)
	}

	defer db.Close(ctx)

	b, err := NewBookshelf(db)
	if err != nil {
		log.Fatalf("NewBookshelf: %v", err)
	}

	b.registerHandlers()

	log.Printf("Listening on localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
