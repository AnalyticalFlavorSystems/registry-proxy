package main

import (
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

var (
	ENV           string
	REGISTRY_PORT string
	REGISTRY_HOST string
)

const (
	PORT = ":5000"
)

type AppContext struct {
	db *bolt.DB
}

func main() {
	log.Println("[STARTING SERVER]")

	db, err := bolt.Open("registry.db", 0666, &bolt.Options{ReadOnly: true})
	if err != nil {
		log.Fatal(err)
	}
	appC := &AppContext{db}
	r := mux.NewRouter()
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	middleware := alice.New(
		recoverHandler,
		loggingHandler,
		setHeaders,
		corsHandler,
		appC.basicauth,
	)
	r.Handle("/v2{rest:.*}", middleware.ThenFunc(registry))
	r.Handle("/ui{rest:.*}", middleware.ThenFunc(ui))
	r.Handle("/", middleware.ThenFunc(registry))

	http.Handle("/", r)
	log.Println("[RUNNING] Serving on " + PORT)
	log.Fatal(http.ListenAndServe(PORT, nil))

}
func init() {
	ENV = os.Getenv("GOENV")
	REGISTRY_HOST = os.Getenv("REGISTRY_HOST")
	if REGISTRY_HOST == "" {
		REGISTRY_HOST = "localhost"
	}
	REGISTRY_PORT = os.Getenv("REGISTRY_PORT")
	if REGISTRY_PORT == "" {
		REGISTRY_PORT = "5001"
	}
}
func registry(w http.ResponseWriter, r *http.Request) {
	defer timeTrack(time.Now(), "registry")
	director := func(req *http.Request) {
		reverse := REGISTRY_HOST + ":" + REGISTRY_PORT
		req = r
		req.URL.Scheme = "http"
		req.URL.Host = reverse
	}
	roundTripper := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   900 * time.Second,
			KeepAlive: 65 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	proxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: roundTripper,
	}
	proxy.ServeHTTP(w, r)
}
func ui(w http.ResponseWriter, r *http.Request) {
}
