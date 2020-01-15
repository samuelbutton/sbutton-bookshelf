package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

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
func (b *Bookshelf) jwtAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := r.URL.Path

		noAuthPaths := []string{
			"/login",
			"/assets/style.css",
			"/logout",
			"/new",
			"/forgot",
			"/reset",
		}

		regex := regexp.MustCompile(`^/reset/+[A-Za-z0-9]+.+[A-Za-z0-9]+.+[A-Za-z0-9]*$`)
		for _, value := range noAuthPaths {

			match := regex.Find([]byte(requestPath))
			if value == requestPath || string(match) == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
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
				fmt.Println("sig failed")
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			fmt.Println(tknStr)
			fmt.Printf("bad 2 request failed: %v", err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if !tkn.Valid {
			fmt.Println("token failed")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		expirationTime := time.Now().Local().Add(5 * time.Minute)
		token.ExpiresAt = expirationTime.Unix()
		newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
		tokenString, err := newToken.SignedString([]byte(os.Getenv("BOOKSHELF_TOKEN_PASSWORD")))
		if err != nil {
			fmt.Printf("token error: %v", err)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
		})

		ctx := context.WithValue(r.Context(), "user", token.UserID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
