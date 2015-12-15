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
	ENV                string
	REGISTRY_PORT_ADDR string
	REGISTRY_HOST_ADDR string
	yield              *render.Render
	store              sessions.Store
)

const (
	PORT = ":3000"
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

	db, err := bolt.Open("db/registry.db", 0666, nil)
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
	get.Handle("/ui/repos/{name}", appMiddleware.ThenFunc(appC.repoShow))
	get.Handle("/ui/users", appMiddleware.ThenFunc(appC.users))
	get.Handle("/ui/users/new", appMiddleware.ThenFunc(appC.newUser))
	post.Handle("/ui/users/{name}/destroy", appMiddleware.ThenFunc(appC.destroyUser))
	post.Handle("/ui/users/create", appMiddleware.ThenFunc(appC.createUser))

	// post
	post.Handle("/login", middleware.ThenFunc(appC.loginPostHandler))

	http.Handle("/", r)
	log.Println("[RUNNING] Serving on " + PORT)
	log.Fatal(http.ListenAndServe(PORT, nil))

}
func init() {
	ENV = os.Getenv("GOENV")
	REGISTRY_HOST_ADDR = os.Getenv("REGISTRY_HOST_ADDR")
	if REGISTRY_HOST_ADDR == "" {
		REGISTRY_HOST_ADDR = "localhost"
	}
	REGISTRY_PORT_ADDR = os.Getenv("REGISTRY_PORT_ADDR")
	if REGISTRY_PORT_ADDR == "" {
		REGISTRY_PORT_ADDR = "5000"
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
