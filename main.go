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

type server struct {
	mux    *chi.Mux
	config *chaletConfig
}

func (s *server) routes() {
	s.mux = chi.NewRouter()
	s.mux.Use(middleware.Logger)
	s.mux.Use(auth)
	s.mux.Get("/", serveTemplate("index.html", s.config))
	s.mux.Get("/chalet.css", serveFile("static/chalet.css"))
	s.mux.Get("/chalet.js", serveFile("static/chalet.js"))
	serveDir(s.mux, "/img/", "img")
	s.mux.Get("/api/*", routeAPI)
}

func main() {
	log.Println("chalet-cloud starting")
	var s server
	s.readConfig()
	s.routes()

	log.Println("listening in port 10753...")
	err := http.ListenAndServeTLS(":10753", "cert.pem", "privkey.pem", s.mux)
	log.Fatal(err)
}

func routeAPI(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("this is the API\n"))
	w.Write([]byte(fmt.Sprintf("%+v", r)))
}
