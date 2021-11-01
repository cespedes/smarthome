package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

func (s *server) authCookieAndBasicAuth(next http.Handler) http.Handler {
	log.Println("authCookieAndBasicAuth() init")
	sessionStore := make(map[string]bool)
	var storageMutex sync.RWMutex
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("home")
		if err != nil && err != http.ErrNoCookie {
			log.Println("auth error")
			http.Error(w, err.Error(), 500)
			return
		}
		var present bool
		var client bool
		if cookie != nil {
			storageMutex.RLock()
			client, present = sessionStore[cookie.Value]
			storageMutex.RUnlock()
		} else {
			present = false
		}
		if present == false {
			cookie = &http.Cookie{
				Name:   "home",
				Path:   "/",
				MaxAge: 24 * 60 * 60,
				Value:  uuid.New().String(),
			}
			storageMutex.Lock()
			sessionStore[cookie.Value] = false
			storageMutex.Unlock()
		}
		http.SetCookie(w, cookie)
		if client == true {
			next.ServeHTTP(w, r)
			return
		}
		log.Println("auth: not logged in, doing BasicAuth")
		middleware.BasicAuth("home", s.config.Auth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("basic auth: logged in")
			storageMutex.Lock()
			sessionStore[cookie.Value] = true
			storageMutex.Unlock()
			next.ServeHTTP(w, r)
		})).ServeHTTP(w, r)
		return
	})
}

func (s *server) authBasicAuth(next http.Handler) http.Handler {
	log.Println("authBasicAuth() init")
	return middleware.BasicAuth("home", s.config.Auth)(next)
}

func (s *server) auth() func(http.Handler) http.Handler {
	// return s.authCookieAndBasicAuth
	return s.authBasicAuth
}
