package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

// Token is a JWT claims struct
type Token struct {
	UserID uint
	jwt.StandardClaims
}

// Account contains user detail
type Account struct {
	ID       uint    `json:"id"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Token    *string `json:"token"`
}

// THE BELOW NEEDS TO BE REFACTORED IN STYLE OF APP

// JwtAuthentication is used for all requests except new and login
var JwtAuthentication = func(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := r.URL.Path

		if requestPath == "/new" || requestPath == "/login" || requestPath == "/assets/style.css" {
			next.ServeHTTP(w, r)
			return
		}

		// response := make(map[string]interface{})
		header := r.Header.Get("Authorization")

		if header == "" {
			// response = Message(false, "Missing auth token")
			// w.WriteHeader(http.StatusForbidden)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		authStringArr := strings.Split(header, " ")
		if len(authStringArr) != 2 {
			// response = Message(false, "Invalid/Malformed auth token")
			// w.WriteHeader(http.StatusForbidden)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		tok := &Token{}

		token, err := jwt.ParseWithClaims(authStringArr[1], tok,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("BOOKSHELF_TOKEN_PASSWORD")), nil
			})

		if err != nil {
			// response = Message(false, "Malformed authentication token")
			// w.WriteHeader(http.StatusForbidden)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if !token.Valid {
			// response = Message(false, "Token is not valid.")
			// w.WriteHeader(http.StatusForbidden)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		fmt.Printf("User %v", tok.UserID)
		ctx := context.WithValue(r.Context(), "user", tok.UserID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
