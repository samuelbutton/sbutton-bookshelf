package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/gofrs/uuid"
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

	// [below is an example of how the mysql database can be used correctly]

	// Prepare statement for inserting data
	// stmtIns, err := db.PrepareContext(ctx, "INSERT INTO squareNum VALUES( ?, ? )")
	// if err != nil {
	// 	log.Fatalf("db.Prepare INSERT: %v", err)
	// }
	// defer stmtIns.Close()

	// Prepare statement for reading data
	// stmtOut, err := db.PrepareContext(ctx, "SELECT squareNumber FROM squarenum WHERE number = ?")
	// if err != nil {
	// 	log.Fatalf("db.Prepare SELECT: %v", err)
	// }
	// defer stmtOut.Close()

	// Insert square numbers for 0-24 in the database
	// for i := 25; i < 50; i++ {
	// 	_, err = stmtIns.Exec(i, (i * i)) // Insert tuples (i, i^2)
	// 	if err != nil {
	// 		log.Fatalf("db.Exec: %v", err)
	// 	}
	// }

	// var squareNum int // we "scan" the result in here

	// Query the square-number of 13
	// err = stmtOut.QueryRow(26).Scan(&squareNum) // WHERE number = 13
	// if err != nil {
	// 	log.Fatalf("db.Query13: %v", err)
	// }
	// fmt.Printf("The square number of 13 is: %d \n", squareNum)

	// Query another number.. 1 maybe?
	// err = stmtOut.QueryRow(49).Scan(&squareNum) // WHERE number = 1
	// if err != nil {
	// 	log.Fatalf("db.Query1: %v", err)
	// }
	// fmt.Printf("The square number of 1 is: %d \n", squareNum)

	// create new object with use of database
	// the google original took in a projectID as well, possibly add authentication
	// makes new object with database
	b, err := NewBookshelf(db)
	if err != nil {
		log.Fatalf("NewBookshelf: %v", err)
	}

	// build a handler specific function
	b.registerHandlers()

	// log.Printf("Listening on localhost:%s", port)
	// if err := http.ListenAndServe(":"+port, nil); err != nil {
	// 	log.Fatal(err)
	// }
}
