package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
)

// Token is a JWT claims struct
type Token struct {
	UserID uint
	jwt.StandardClaims
}

// Account contains user detail
type Account struct {
	ID       uint
	Email    *string
	Password *string
	Token    *string
}

// JwtAuthentication is used for all requests except new and login
var JwtAuthentication = func(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := r.URL.Path

		if requestPath == "/new" || requestPath == "/login" || requestPath == "/assets/style.css" {
			next.ServeHTTP(w, r)
			return
		}

		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				// w.WriteHeader(http.StatusUnauthorized)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			// w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("bad request failed: %v", err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		tknStr := c.Value

		token := &Token{}

		tkn, err := jwt.ParseWithClaims(tknStr, token, func(tken *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("BOOKSHELF_TOKEN_PASSWORD")), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				// w.WriteHeader(http.StatusUnauthorized)
				fmt.Println("sig failed")
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			// w.WriteHeader(http.StatusBadRequest)
			fmt.Println(tknStr)
			fmt.Printf("bad 2 request failed: %v", err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if !tkn.Valid {
			// w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("token failed")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// w.Write([]byte(fmt.Sprintf("Welcome %v!", token.UserID)))

		fmt.Printf("User %v", token.UserID)
		ctx := context.WithValue(r.Context(), "user", token.UserID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
