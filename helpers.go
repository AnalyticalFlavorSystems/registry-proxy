package main

import (
	"log"
	"net/http"
	"time"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"registry.localhost\"")
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
