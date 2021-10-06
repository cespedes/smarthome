package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func serveFile(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	}
}

func serveDir(r *chi.Mux, prefix string, dir string) {
	r.Mount(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(dir))))
}

func main() {
	log.Println("chalet-cloud starting")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(auth)

	r.Get("/", serveFile("index.html"))
	r.Get("/chalet.css", serveFile("chalet.css"))
	r.Get("/chalet.js", serveFile("chalet.js"))
	serveDir(r, "/img/", "img")
	// serveDir(r, "/js/", "js")
	// serveDir(r, "/css/", "css")
	r.Get("/api/*", routeAPI)

	//http.Handle("/hello", helloWorldHandler{})
	//http.Handle("/secureHello", authenticate(helloWorldHandler{}))
	//http.HandleFunc("/login", handleLogin)

	log.Println("listening in port 10753...")
	err := http.ListenAndServeTLS(":10753", "cert.pem", "privkey.pem", r)
	log.Fatal(err)
}

func routeAPI(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("this is the API\n"))
	w.Write([]byte(fmt.Sprintf("%+v", r)))
}
