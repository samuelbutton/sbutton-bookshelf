package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type mysqlDB struct {
	database *sql.DB
}

var _ BookDatabase = &mysqlDB{}

func newMySQLDB() (*mysqlDB, error) {
	ctx := context.Background()
	dbDSN := os.Getenv("MYSQL_DSN")
	if dbDSN == "" {
		log.Fatal("MYSQL_DSN must be set")
	}
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}

	// verify that a connection can be made before making a query
	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("db.Ping: %v", err)
	}
	return &mysqlDB{
		database: db,
	}, nil
}

// Close closes the database.
func (db *mysqlDB) Close(context.Context) error {
	return db.database.Close()
}

// Book retrieves a book by its ID.
func (db *mysqlDB) GetBook(ctx context.Context, id string) (*Book, error) {
	// pull the book from db (return *Rows)
	rs, err := db.database.QueryContext(ctx, fmt.Sprintf("SELECT * FROM bookshelf WHERE id = %q", id))
	if err != nil {
		return nil, fmt.Errorf("mysqldb query GetBook: Get: %v", err)
	}

	defer rs.Close()

	b := &Book{}
	if err := rs.Scan(b); err != nil {
		return nil, fmt.Errorf("mysqldb scan GetBook: Get: %v", err)
	}
	return b, nil
}

// AddBook saves a given book, assigning it a new ID.
func (db *mysqlDB) AddBook(ctx context.Context, b *Book) (id string, err error) {
	// add the book to db
	stmtIns, err := db.database.PrepareContext(ctx, "INSERT INTO bookshelf (title, author, pages, publishedDate, imageURL, description) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return "", fmt.Errorf("mysqldb prepare AddBook: Get: %v", err)
	}
	// temp fix on line below
	res, err := stmtIns.ExecContext(ctx, b.Title, b.Author, b.Pages, b.PublishedDate, b.ImageURL, b.Description)
	if err != nil {
		return "", fmt.Errorf("mysqldb exec AddBook: Get: %v", err)
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("mysqldb result AddBook: Get: %v", err)
	}

	b.ID = strconv.FormatInt(lastID, 10)
	return strconv.FormatInt(lastID, 10), nil
}

// DeleteBook removes a given book by its ID.
func (db *mysqlDB) DeleteBook(ctx context.Context, id string) error {
	// delete the book from db

	stmt, err := db.database.PrepareContext(ctx, fmt.Sprintf("DELETE FROM bookshelf WHERE id = %q", id))
	if err != nil {
		return fmt.Errorf("mysqldb prepare DeleteBook: Get: %v", err)
	}

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("mysqldb exec DeleteBook: Get: %v", err)
	}
	return nil
}

// UpdateBook updates the entry for a given book.
func (db *mysqlDB) UpdateBook(ctx context.Context, b *Book) error {
	// set book from db.db
	stmtIns, err := db.database.PrepareContext(ctx, fmt.Sprintf("UPDATE bookshelf SET id=?, title=?, author=?, pages=?, publishedDate=?, imageURL=?, description=? WHERE id = %q", b.ID))
	if err != nil {
		return fmt.Errorf("mysqldb prepare UpdateBook: Get: %v", err)
	}

	_, err = stmtIns.ExecContext(ctx, b)
	if err != nil {
		return fmt.Errorf("mysqldb exec UpdateBook: Get: %v", err)
	}
	return nil
}

// ListBooks returns a list of books, ordered by title.
func (db *mysqlDB) ListBooks(ctx context.Context) ([]*Book, error) {
	books := make([]*Book, 0)
	// list all books from db by title

	rs, err := db.database.QueryContext(ctx, "SELECT * FROM bookshelf ORDER BY title ASC")
	if err != nil {
		return nil, fmt.Errorf("mysqldb query ListBooks: Get: %v", err)
	}

	defer rs.Close()

	for {
		ready := rs.Next()
		fmt.Printf("Next %v", ready)
		if ready == false {
			break
		}
		b := &Book{}

		err := rs.Scan(&b.ID, &b.Title, &b.Author, &b.Pages, &b.PublishedDate, &b.ImageURL, &b.Description)
		if err != nil {
			return nil, fmt.Errorf("mysqldb: could not list books: %v", err)
		}
		log.Printf("Book %q ID: %q", b.Title, b.ID)
		books = append(books, b)
	}

	return books, nil
}
