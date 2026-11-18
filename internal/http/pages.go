package http

import (
	"net/http"

	"github.com/WeKnowXD/WeKnowXD_Project/internal/models"
)

func layoutHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "layout.html", nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	data := models.PageData{Query: r.URL.Query().Get("q")}
	templates.ExecuteTemplate(w, "layout.html", data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "register.html", models.RegisterData{})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", models.LoginData{})
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "weather.html", models.LoginData{})
}
