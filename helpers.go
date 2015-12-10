package main

import (
	"net/http"
)

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"registry.localhost\"")
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
