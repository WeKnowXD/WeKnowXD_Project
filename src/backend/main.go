package main

import (
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseFiles("templates/search_X.html", "templates/register_X.html", "templates/login_X.html"))

type PageData struct {
	Query   string
	Results []struct {
		Title       string
		URL         string
		Description string
	}
}

type RegisterData struct {
	Error    string
	Username string
	Email    string
}

type LoginData struct {
	Error    string
	Username string
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		data := PageData{Query: r.URL.Query().Get("q")}
		templates.ExecuteTemplate(w, "search_X.html", data)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		templates.ExecuteTemplate(w, "register_X.html", RegisterData{})
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		templates.ExecuteTemplate(w, "login_X.html", LoginData{})
	})

	http.ListenAndServe(":8080", nil)
}
