package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"os"
)

func baseAuthFunc(next http.HandlerFunc) http.HandlerFunc {
	/*
		Super simple implementation of BaseAuth.

	*/
	bauthUser := os.Getenv("METRICS_USERNAME")
	bauthPassword := os.Getenv("METRICS_PASSWORD")

	if len(bauthUser) < 1 {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(bauthUser))
			expectedPasswordHash := sha256.Sum256([]byte(bauthPassword))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
