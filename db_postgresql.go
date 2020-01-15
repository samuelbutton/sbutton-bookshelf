package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type postgresDB struct {
	database *sql.DB
}

var _ BookDatabase = &postgresDB{}

func newpostgresDB() (*postgresDB, error) {
	ctx := context.Background()
	psqlInfo := os.Getenv("DATABASE_URL")
	if psqlInfo == "" {
		log.Fatal("DATABASE_URL must be set")
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

// func (db *postgresDB) GetUser(ctx context.Context, id string) (*Account, error) {
// 	intID, err := strconv.ParseInt(id, 10, 64)
// 	if err != nil {
// 		return nil, fmt.Errorf("postgresqlDB parse GetUser: %v", err)
// 	}

// 	rs, err := db.database.QueryContext(ctx, `SELECT * FROM accounts WHERE id = $1`, intID)
// 	defer rs.Close()
// 	if err != nil {
// 		return nil, fmt.Errorf("postgresqlDB query GetUser: Get: %v", err)
// 	}

// 	ready := rs.Next()
// 	if ready == false {
// 		return nil, fmt.Errorf("postgresqlDB next GetUser: %v", ready)
// 	}

// 	a := &Account{}

// 	err = rs.Scan(&a.ID, &a.Email, &a.Password)
// 	if UseString(a.Email) == "" {
// 		return nil, fmt.Errorf("postgresqlDB scan GetUser: Get: %v", err)
// 	}

// 	a.Password = UsePointer("")
// 	return a, nil
// }

func (db *postgresDB) ValidateAccount(ctx context.Context, a *Account) error {
	// probably refactor
	re := regexp.MustCompile(".+@.+\\..+")
	matched := re.Match([]byte(UseString(a.Email)))
	if matched == false {
		return fmt.Errorf("postgresqlDB email ValidateAccount")
	}

	// probably refactor
	if len(UseString(a.Password)) < 8 {
		return fmt.Errorf("postgresqlDB password ValidateAccount")
	}

	tempAccount := &Account{}
	var (
		tempToken *string
		tempExp   *time.Time
	)

	rs, err := db.database.QueryContext(ctx, `SELECT * FROM accounts WHERE email = $1`, UseString(a.Email))
	defer rs.Close()
	rs.Next()
	err = rs.Scan(&tempAccount.ID, &tempAccount.Email, &tempAccount.Password, &tempToken, &tempExp)
	tempToken = UsePointer("")
	tempExp = nil
	if UseString(tempAccount.Email) != "" {
		return fmt.Errorf("Email already in use!: %v", err)
	}

	return nil
}

func (db *postgresDB) CreateAccount(ctx context.Context, a *Account) (*Account, time.Time, error) {
	if err := db.ValidateAccount(ctx, a); err != nil {
		return nil, time.Now().Local(), err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(UseString(a.Password)), bcrypt.DefaultCost)
	a.Password = UsePointer(string(hashedPassword))

	err = db.database.QueryRow(`INSERT INTO accounts(email, password)
		VALUES($1, $2) RETURNING id`, UseString(a.Email), UseString(a.Password)).Scan(&a.ID)
	if err != nil {
		return nil, time.Now().Local(), fmt.Errorf("Could not create account!: %v", err)
	}

	if a.ID <= 0 {
		return nil, time.Now().Local(), fmt.Errorf("Could not create account!: %v", err)
	}

	expirationTime := time.Now().Local().Add(5 * time.Minute)
	tk := &Token{
		UserID: a.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := token.SignedString([]byte(os.Getenv("BOOKSHELF_TOKEN_PASSWORD")))
	if err != nil {
		return nil, time.Now().Local(), fmt.Errorf("Please log in with new account!: %v", err)
	}
	a.Token = UsePointer(tokenString)

	a.Password = UsePointer("")

	return a, expirationTime, nil
}

func (db *postgresDB) GetResetToken(ctx context.Context, email string) (string, error) {
	a := &Account{}
	var (
		tempToken *string
		tempExp   *time.Time
	)

	rs, err := db.database.QueryContext(ctx, `SELECT * FROM accounts WHERE email = $1`, email)
	defer rs.Close()
	rs.Next()
	err = rs.Scan(&a.ID, &a.Email, &a.Password, &tempToken, &tempExp)
	tempToken = UsePointer("")
	tempExp = nil
	a.Password = UsePointer("")
	if UseString(a.Email) == "" {
		return "", fmt.Errorf("postgresqlDB email GetResetToken: %v", err)
	}

	expirationTime := time.Now().Local().Add(60 * time.Minute)

	tk := &Token{
		UserID: a.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := token.SignedString([]byte(os.Getenv("BOOKSHELF_RESET_TOKEN_PASSWORD")))
	if err != nil {
		return "", fmt.Errorf("postgresqlDB sign GetResetToken: %v", err)
	}

	_, err = db.database.ExecContext(ctx, `UPDATE accounts SET reset_token = $1, token_expiration
		= $2 where email = $3`, tokenString, expirationTime, UseString(a.Email))
	if err != nil {
		return "", fmt.Errorf("postgresqlDB exec GetResetToken: %v", err)
	}

	return tokenString, nil
}

func (db *postgresDB) CheckTokenValidity(ctx context.Context, tokenString string, a *Account) error {
	rs, err := db.database.QueryContext(ctx, `SELECT * FROM accounts WHERE email = $1`, UseString(a.Email))
	defer rs.Close()
	var (
		savedToken      *string
		tokenExpiration *time.Time
	)
	rs.Next()
	tempAccount := &Account{}
	err = rs.Scan(&tempAccount.ID, &tempAccount.Email, &tempAccount.Password, &savedToken, &tokenExpiration)
	tempAccount.Password = UsePointer("")
	if UseString(tempAccount.Email) == "" {
		savedToken = UsePointer("")
		return fmt.Errorf("postgresqlDB email CheckTokenValidity: %v", err)
	}
	if tokenExpiration == nil {
		savedToken = UsePointer("")
		return fmt.Errorf("postgresqlDB expiration nil CheckTokenValidity: %v", err)
	}
	if time.Now().Local().After(*tokenExpiration) {
		savedToken = UsePointer("")
		return fmt.Errorf("postgresqlDB expired CheckTokenValidity: %v", err)
	}
	if UseString(savedToken) == "" {
		savedToken = UsePointer("")
		return fmt.Errorf("postgresqlDB token empty CheckTokenValidity: %v", err)
	}
	if UseString(savedToken) != tokenString {
		savedToken = UsePointer("")
		return fmt.Errorf("postgresqlDB token incorrect CheckTokenValidity: %v", err)
	}
	return nil
}

func (db *postgresDB) RemoveToken(ctx context.Context, a *Account) error {
	_, err := db.database.ExecContext(ctx, `UPDATE accounts SET reset_token = $1, token_expiration
		= $2 where email = $3`, "", time.Now().Local(), UseString(a.Email))
	if err != nil {
		return fmt.Errorf("postgresqlDB exec RemoveToken: %v", err)
	}
	return nil
}

func (db *postgresDB) UpdatePassword(ctx context.Context, a *Account) error {
	// probably refactor
	if len(UseString(a.Password)) < 8 {
		return fmt.Errorf("postgresqlDB password UpdatePassword")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(UseString(a.Password)), bcrypt.DefaultCost)
	a.Password = UsePointer(string(hashedPassword))

	_, err = db.database.ExecContext(ctx, `UPDATE accounts SET password = $1 where email = $2`,
		UseString(a.Password), UseString(a.Email))
	if err != nil {
		return fmt.Errorf("postgresqlDB exec UpdatePassword: %v", err)
	}
	a.Password = UsePointer("")
	return nil
}

func (db *postgresDB) LoginAccount(ctx context.Context, email string, password string) (*Account, time.Time, error) {

	rs, err := db.database.QueryContext(ctx, `SELECT * FROM accounts WHERE email = $1`, email)
	defer rs.Close()
	if err != nil {
		return nil, time.Now().Local(), fmt.Errorf("Server error, please try again!: %v", err)
	}

	ready := rs.Next()
	if ready == false {
		return nil, time.Now().Local(), fmt.Errorf("Invalid credentials, please try again!: %v", ready)
	}

	a := &Account{}
	var (
		tempToken *string
		tempExp   *time.Time
	)

	err = rs.Scan(&a.ID, &a.Email, &a.Password, &tempToken, &tempExp)
	tempToken = UsePointer("")
	tempExp = nil
	if UseString(a.Email) == "" {
		return nil, time.Now().Local(), fmt.Errorf("Invalid credentials, please try again!: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(UseString(a.Password)), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, time.Now().Local(), fmt.Errorf("Invalid credentials, please try again!: %v", err)
		}
		return nil, time.Now().Local(), fmt.Errorf("Server error, please try again!: %v", err)
	}

	a.Password = UsePointer("")

	expirationTime := time.Now().Local().Add(5 * time.Minute)
	tk := &Token{
		UserID: a.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, err := token.SignedString([]byte(os.Getenv("BOOKSHELF_TOKEN_PASSWORD")))
	if err != nil {
		return nil, time.Now().Local(), fmt.Errorf("Server error, please try again!: %v", err)
	}
	a.Token = UsePointer(tokenString)

	return a, expirationTime, nil
}

func (db *postgresDB) GetBook(ctx context.Context, id string) (*Book, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("postgresqlDB parse GetBook: %v", err)
	}

	rs, err := db.database.QueryContext(ctx, `SELECT * FROM bookshelf WHERE id = $1`, intID)
	if err != nil {
		return nil, fmt.Errorf("postgresqlDB query GetBook: %v", err)
	}

	ready := rs.Next()
	if ready == false {
		return nil, fmt.Errorf("postgresqlDB next GetBook: %v", ready)
	}
	b := &Book{}

	err = rs.Scan(&b.ID, &b.Title, &b.Author, &b.Pages, &b.PublishedDate, &b.ImageURL, &b.Description)
	if err != nil {
		return nil, fmt.Errorf("postgresqlDB scan GetBook: %v", err)
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
		return "", fmt.Errorf("postgresqlDB exec AddBook: %v", err)
	}

	b.ID = UsePointer(strconv.Itoa(lastID))
	return strconv.Itoa(lastID), nil
}

func (db *postgresDB) DeleteBook(ctx context.Context, id string) error {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("postgresqlDB parse DeleteBook: %v", err)
	}

	stmt, err := db.database.PrepareContext(ctx, `DELETE FROM bookshelf WHERE id = $1`)
	if err != nil {
		return fmt.Errorf("postgresqlDB prepare DeleteBook: %v", err)
	}

	_, err = stmt.ExecContext(ctx, intID)
	if err != nil {
		return fmt.Errorf("postgresqlDB exec DeleteBook: %v", err)
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
		return fmt.Errorf("postgresqlDB prepare UpdateBook: %v", err)
	}

	_, err = stmtIns.ExecContext(ctx, UseString(b.Title), UseString(b.Author), UseString(b.Pages), UseString(b.PublishedDate), UseString(b.ImageURL), UseString(b.Description))
	if err != nil {
		return fmt.Errorf("postgresqlDB exec UpdateBook: %v", err)
	}
	return nil
}

func (db *postgresDB) ListBooks(ctx context.Context) ([]*Book, error) {
	books := make([]*Book, 0)
	rs, err := db.database.QueryContext(ctx, "SELECT * FROM bookshelf ORDER BY title ASC")
	if err != nil {
		return nil, fmt.Errorf("postgresqlDB query ListBooks: %v", err)
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
