package main

import (
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/justinas/alice"
	"github.com/unrolled/render"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
)

var (
	ENV           string
	REGISTRY_PORT string
	REGISTRY_HOST string
	yield         *render.Render
	store         sessions.Store
)

const (
	PORT = ":5000"
)

type ProxyContext struct {
	db *bolt.DB
}
type AppContext struct {
	db     *bolt.DB
	layout render.HTMLOptions
}

func main() {
	log.Println("[STARTING SERVER]")

	db, err := bolt.Open("registry.db", 0666, &bolt.Options{ReadOnly: true})
	if err != nil {
		log.Fatal(err)
	}
	proxyC := &ProxyContext{db}
	appC := &AppContext{
		db,
		render.HTMLOptions{Layout: "layouts/mainLayout"},
	}
	r := mux.NewRouter()
	get := r.Methods("GET").Subrouter()
	post := r.Methods("POST").Subrouter()

	store = sessions.NewCookieStore([]byte("ndJSfN0wEt1JRrFIxoELEZYXRad2g82l9kYtJjOca23VXK3SvvPbDYuF0bqP"))
	middleware := alice.New(
		recoverHandler,
		loggingHandler,
		setHeaders,
		corsHandler,
	)
	if ENV == "production" {
		yield = render.New(render.Options{
			Directory:     "views",
			Layout:        "layouts/plainLayout",
			IsDevelopment: false,
		})
	} else {
		yield = render.New(render.Options{
			Directory:     "views",
			Layout:        "layouts/plainLayout",
			IsDevelopment: true,
		})
	}

	proxyMiddleware := middleware.Append(proxyC.basicauth)
	appMiddleware := middleware.Append(appC.authHandler)

	// registry
	r.Handle("/v2{rest:.*}", proxyMiddleware.ThenFunc(registry))
	r.Handle("/", proxyMiddleware.ThenFunc(registry))

	// GET assets
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))

	get.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// registry ui
	get.Handle("/login", middleware.ThenFunc(appC.loginHandler))
	get.Handle("/ui", appMiddleware.ThenFunc(appC.ui))
	get.Handle("/ui/{name}", appMiddleware.ThenFunc(appC.repoShow))

	// post
	post.Handle("/login", middleware.ThenFunc(appC.loginPostHandler))

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

func SetPassword(password string) []byte {
	hpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hpass
}

func (c *AppContext) Login(username, password string) (bool, error) {
	var HashedPassword []byte

	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("auth"))
		HashedPassword = b.Get([]byte(username))
		return nil
	})
	if err != nil {
		return false, err
	}

	// Test if exists
	if string(HashedPassword) == "" {
		return false, nil
	}

	err = bcrypt.CompareHashAndPassword(HashedPassword, []byte(password))
	if err != nil {
		return false, err
	}
	return true, err
}
