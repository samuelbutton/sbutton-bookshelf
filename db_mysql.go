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

// MYSQL IMPLEMENTATION MUST BE UPDATED FOR ACCOUNTS
// var _ BookDatabase = &mysqlDB{}

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

	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("db.Ping: %v", err)
	}
	return &mysqlDB{
		database: db,
	}, nil
}

func (db *mysqlDB) Close(context.Context) error {
	return db.database.Close()
}

func (db *mysqlDB) GetBook(ctx context.Context, id string) (*Book, error) {
	rs, err := db.database.QueryContext(ctx, fmt.Sprintf("SELECT * FROM bookshelf WHERE id = %q", id))
	if err != nil {
		return nil, fmt.Errorf("mysqldb query GetBook: Get: %v", err)
	}

	ready := rs.Next()
	if ready == false {
		return nil, fmt.Errorf("mysqldb next GetBook: %v", ready)
	}
	b := &Book{}

	err = rs.Scan(&b.ID, &b.Title, &b.Author, &b.Pages, &b.PublishedDate, &b.ImageURL, &b.Description)
	if err != nil {
		return nil, fmt.Errorf("mysqldb scan GetBook: Get: %v", err)
	}
	defer rs.Close()
	return b, nil
}

func (db *mysqlDB) AddBook(ctx context.Context, b *Book) (id string, err error) {
	stmtIns, err := db.database.PrepareContext(ctx, "INSERT INTO bookshelf (title, author, pages, publishedDate, imageURL, description) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return "", fmt.Errorf("mysqldb prepare AddBook: Get: %v", err)
	}
	// temp fix on line below
	res, err := stmtIns.ExecContext(ctx, UseString(b.Title), UseString(b.Author), UseString(b.Pages), UseString(b.PublishedDate), UseString(b.ImageURL), UseString(b.Description))
	if err != nil {
		return "", fmt.Errorf("mysqldb exec AddBook: Get: %v", err)
	}

	lastID, err := res.LastInsertId()

	if err != nil {
		return "", fmt.Errorf("mysqldb result AddBook: Get: %v", err)
	}

	b.ID = UsePointer(strconv.FormatInt(lastID, 10))
	return strconv.FormatInt(lastID, 10), nil
}

func (db *mysqlDB) DeleteBook(ctx context.Context, id string) error {
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

func (db *mysqlDB) UpdateBook(ctx context.Context, b *Book) error {
	stmtIns, err := db.database.PrepareContext(ctx, fmt.Sprintf("UPDATE bookshelf SET title=?, author=?, pages=?, publishedDate=?, imageURL=?, description=? WHERE ID = %q", UseString(b.ID)))
	if err != nil {
		return fmt.Errorf("mysqldb prepare UpdateBook: Get: %v", err)
	}

	_, err = stmtIns.ExecContext(ctx, UseString(b.Title), UseString(b.Author), UseString(b.Pages), UseString(b.PublishedDate), UseString(b.ImageURL), UseString(b.Description))
	if err != nil {
		return fmt.Errorf("mysqldb exec UpdateBook: Get: %v", err)
	}
	return nil
}

func (db *mysqlDB) ListBooks(ctx context.Context) ([]*Book, error) {
	books := make([]*Book, 0)
	rs, err := db.database.QueryContext(ctx, "SELECT * FROM bookshelf ORDER BY title ASC")
	if err != nil {
		return nil, fmt.Errorf("mysqldb query ListBooks: Get: %v", err)
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
			return nil, fmt.Errorf("mysqldb: could not list books: %v", err)
		}
		log.Printf("Book %q ID: %q", UseString(b.Title), UseString(b.ID))
		books = append(books, b)
	}

	return books, nil
}
