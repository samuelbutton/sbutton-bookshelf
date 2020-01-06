package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

type postgresDB struct {
	database *sql.DB
}

var _ BookDatabase = &postgresDB{}

const (
	host     = "localhost"
	port     = 5432
	user     = "sambutton"
	password = "laoz16"
	dbname   = "sambutton"
)

func newpostgresDB() (*postgresDB, error) {
	ctx := context.Background()
	psqlInfo := os.Getenv("PSQL_INFO")
	if psqlInfo == "" {
		log.Fatal("PSQL_INFO must be set")
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}

	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("db.Ping: %v", err)
	}
	return &postgresDB{
		database: db,
	}, nil
}

func (db *postgresDB) Close(context.Context) error {
	return db.database.Close()
}

func (db *postgresDB) GetBook(ctx context.Context, id string) (*Book, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("postgresqlDB parse GetBook: %v", err)
	}

	rs, err := db.database.QueryContext(ctx, fmt.Sprintf("SELECT * FROM bookshelf WHERE id = %v", intID))
	if err != nil {
		return nil, fmt.Errorf("postgresqlDB query GetBook: Get: %v", err)
	}

	ready := rs.Next()
	if ready == false {
		return nil, fmt.Errorf("postgresqlDB next GetBook: %v", ready)
	}
	b := &Book{}

	err = rs.Scan(&b.ID, &b.Title, &b.Author, &b.Pages, &b.PublishedDate, &b.ImageURL, &b.Description)
	if err != nil {
		return nil, fmt.Errorf("postgresqlDB scan GetBook: Get: %v", err)
	}
	defer rs.Close()
	return b, nil
}

func (db *postgresDB) AddBook(ctx context.Context, b *Book) (id string, err error) {
	var lastID int

	err = db.database.QueryRow(`INSERT INTO bookshelf(title, author, pages, publisheddate, imageurl, description)
		VALUES($1, $2, $3, $4, $5, $6) RETURNING id`, UseString(b.Title), UseString(b.Author), UseString(b.Pages),
		UseString(b.PublishedDate), UseString(b.ImageURL), UseString(b.Description)).Scan(&lastID)

	if err != nil {
		return "", fmt.Errorf("postgresqlDB exec AddBook: Get: %v", err)
	}

	b.ID = UsePointer(strconv.Itoa(lastID))
	return strconv.Itoa(lastID), nil
}

func (db *postgresDB) DeleteBook(ctx context.Context, id string) error {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("postgresqlDB parse DeleteBook: %v", err)
	}

	stmt, err := db.database.PrepareContext(ctx, fmt.Sprintf("DELETE FROM bookshelf WHERE id = %v", intID))
	if err != nil {
		return fmt.Errorf("postgresqlDB prepare DeleteBook: Get: %v", err)
	}

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("postgresqlDB exec DeleteBook: Get: %v", err)
	}
	return nil
}

func (db *postgresDB) UpdateBook(ctx context.Context, b *Book) error {
	intID, err := strconv.ParseInt(UseString(b.ID), 10, 64)
	if err != nil {
		return fmt.Errorf("postgresqlDB parse UpdateBook: %v", err)
	}

	stmtIns, err := db.database.PrepareContext(ctx, fmt.Sprintf("UPDATE bookshelf SET title=$1, author=$2, pages=$3, publishedDate=$4, imageURL=$5, description=$6 WHERE ID = %v", intID))
	if err != nil {
		return fmt.Errorf("postgresqlDB prepare UpdateBook: Get: %v", err)
	}

	_, err = stmtIns.ExecContext(ctx, UseString(b.Title), UseString(b.Author), UseString(b.Pages), UseString(b.PublishedDate), UseString(b.ImageURL), UseString(b.Description))
	if err != nil {
		return fmt.Errorf("postgresqlDB exec UpdateBook: Get: %v", err)
	}
	return nil
}

func (db *postgresDB) ListBooks(ctx context.Context) ([]*Book, error) {
	books := make([]*Book, 0)
	rs, err := db.database.QueryContext(ctx, "SELECT * FROM bookshelf ORDER BY title ASC")
	if err != nil {
		return nil, fmt.Errorf("postgresqlDB query ListBooks: Get: %v", err)
	}

	defer rs.Close()

	for {
		ready := rs.Next()
		if ready == false {
			break
		}
		b := &Book{}

		err := rs.Scan(&b.ID, &b.Title, &b.Author, &b.Pages, &b.PublishedDate, &b.ImageURL, &b.Description)
		if err != nil {
			return nil, fmt.Errorf("postgresqlDB: could not list books: %v", err)
		}
		log.Printf("Book %q ID: %q", UseString(b.Title), UseString(b.ID))
		books = append(books, b)
	}

	return books, nil
}
