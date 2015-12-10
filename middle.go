package main

import (
	"bytes"
	"encoding/base64"
	"github.com/boltdb/bolt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

func corsHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == `^(docker\/1\.(3|4|5(?!\.[0-9]-dev))|Go ).*$` {
			http.NotFound(w, r)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers",
				"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}
		// Stop here if its Preflighted OPTIONS request
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func setHeaders(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
		w.Header().Set("Strict-Transport-Security", "max-age=15638400")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (c *ProxyContext) basicauth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header["Authorization"]
		if len(authHeader) <= 0 {
			unauthorized(w)
			return
		}
		authorization := strings.TrimSpace(authHeader[0])
		credentials := strings.Split(authorization, " ")

		// Check if basic
		if len(credentials) != 2 || credentials[0] != "Basic" {
			unauthorized(w)
			return
		}
		authstr, err := base64.StdEncoding.DecodeString(credentials[1])
		if err != nil {
			unauthorized(w)
			return
		}
		userpass := strings.Split(string(authstr), ":")
		if len(userpass) != 2 {
			unauthorized(w)
			return
		}
		var pwd []byte
		err = c.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("auth"))
			pwd = b.Get([]byte(userpass[0]))
			return nil
		})

		if bytes.Equal(pwd, []byte(userpass[1])) {
			next.ServeHTTP(w, r)
		} else {
			unauthorized(w)
		}
	}
	return http.HandlerFunc(fn)
}

func loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v: %s", err, debug.Stack())
				http.Error(w, http.StatusText(500), 500)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
func (c *AppContext) authHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// REGULAR SESSION
		session, err := store.Get(r, "registry")
		if err != nil {
			log.Println(err)
			return
		}
		if session.Values["username"] == nil {
			c.loginHandler(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
