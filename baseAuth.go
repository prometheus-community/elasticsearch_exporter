package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"
	"os"
)

func authWrapper(h http.Handler) http.Handler {
	bauthUser := os.Getenv("METRICS_USERNAME")
	bauthPassword := os.Getenv("METRICS_PASSWORD")

	if len(bauthUser) < 1 {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r) // call original
			return
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if ok {
			log.Println("Doing basic auth verification")

			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(bauthUser))
			expectedPasswordHash := sha256.Sum256([]byte(bauthPassword))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				h.ServeHTTP(w, r) // call original
				return
			}
		}

		// Wrong user or password
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	})
}
