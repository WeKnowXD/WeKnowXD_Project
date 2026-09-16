package main

import (
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseFiles("templates/search.html", "templates/register.html"))

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

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		data := PageData{Query: r.URL.Query().Get("q")}
		templates.ExecuteTemplate(w, "search.html", data)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		templates.ExecuteTemplate(w, "register.html", RegisterData{})
	})
	
	http.ListenAndServe(":8080", nil)
}