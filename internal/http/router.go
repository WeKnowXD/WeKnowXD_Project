package http

import (
	"net/http"

	"github.com/WeKnowXD/WeKnowXD_Project/web"
)

func Router(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", layoutHandler)
	mux.HandleFunc("GET /register", registerHandler)
	mux.HandleFunc("GET /search", searchHandler)
	mux.HandleFunc("GET /api/search", apiSearch)
	mux.HandleFunc("GET /login", loginHandler)
	mux.HandleFunc("GET /api/weather", apiWeather)

	mux.HandleFunc("GET /weather", weatherHandler)

	mux.HandleFunc("POST /api/register", apiRegister)
	mux.HandleFunc("POST /api/login", apiLogin)
	mux.HandleFunc("GET /api/logout", apiLogout)
	mux.HandleFunc("POST /test", testSessions)

	mux.Handle("/static/", http.FileServer(http.FS(web.Static)))
}
