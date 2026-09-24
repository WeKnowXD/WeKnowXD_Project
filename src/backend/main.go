// Package main implements the backend server for the WhoKnows search and registration app.
package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3" // this one's blank on purpose, we're not calling anything from it
	// directly, it just needs to load so sqlite3 gets registered as a driver
	//note: youll need to install a C compiler for this to work, since the SQLite3 driver is a cgo package:) and also do go env -w CGO_ENABLED=1 if its not already enabled
)

var templates = template.Must(template.ParseFiles("templates/search.html", "templates/register.html", "templates/layout.html"))

func main() {
	initDB() // gotta connect to the db before the server starts taking requests
	mux := http.NewServeMux()
	router(mux)
	http.ListenAndServe(":8080", mux)
}

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

func router(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", layoutHandler)
	mux.HandleFunc("GET /register", registerHandler)
	mux.HandleFunc("GET /search", searchHandler)
	mux.HandleFunc("GET /api/search", getSearch)
	mux.HandleFunc("POST /api/register", postRegister)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

}

// TODO: currently passing nil since we don't have session/auth handling yet.
// Once that's built, replace this with a struct (e.g. LayoutData) holding
// User (nil if not logged in) and Flashes ([]string), so layout.html's
// {{ if .User }} and {{ if .Flashes }} blocks actually have data to work with.
func layoutHandler(w http.ResponseWriter, _ *http.Request) {
	templates.ExecuteTemplate(w, "layout.html", nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{Query: r.URL.Query().Get("q")}
	templates.ExecuteTemplate(w, "layout.html", data)
}

func registerHandler(w http.ResponseWriter, _ *http.Request) {
	templates.ExecuteTemplate(w, "register.html", RegisterData{})
}

var db *sql.DB // shared connection, every handler in this file can just use this directly

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "../whoknows.db") // note this is = not :=, since db already exists above
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil { // Open doesn't actually connect, Ping is what forces the real check
		log.Fatal(err)
	}
}

func getSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q") // whatever the user typed into the search bar
	language := r.URL.Query().Get("language")
	if q == "" {
		response := map[string]any{"data": []map[string]any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	if language == "" {
		language = "en"
	}

	// the ? is a stand-in, the real value gets slotted in safely as the second argument
	// this is basically to stop SQL injections
	rows, err := db.Query("SELECT title, url, content FROM pages WHERE language = ? AND content LIKE ?", language, "%"+q+"%")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close() // makes sure this closes once the function's done, no matter how it exits

	var results []map[string]any
	for rows.Next() { // grabs one row at a time until there's nothing left
		var title, url, content string
		if err := rows.Scan(&title, &url, &content); err != nil { // pulls that row's values into these three
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, map[string]any{
			"title":   title,
			"url":     url,
			"content": content,
		})
	}
	response := map[string]any{"data": results}

	// rows.Next() returning false could mean "all done" or "something broke" — this catches the second case
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func postRegister(w http.ResponseWriter, r *http.Request) {
	// this route gets form data, not JSON, so FormValue instead of decoding a JSON body
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	password2 := r.FormValue("password2") // not required, but old code checked it so keeping that behavior

	var errorMsg string
	if username == "" {
		errorMsg = "You have to enter a username"
	} else if email == "" || !strings.Contains(email, "@") {
		errorMsg = "You have to enter a valid email address"
	} else if password == "" {
		errorMsg = "You have to enter a password"
	} else if password2 != "" && password != password2 {
		errorMsg = "The two passwords do not match"
	} else {
		// only bother hitting the db if everything else checked out already
		var existingID int
		err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&existingID)
		if err == nil { // a row came back, so that username's taken
			errorMsg = "Username already taken"
		}
	}

	if errorMsg != "" {
		response := map[string]any{
			"statusCode": http.StatusBadRequest,
			"message":    errorMsg,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	hashedPassword := hashPassword(password) // never store the raw password

	_, err := db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, hashedPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"statusCode": http.StatusOK,
		"message":    "Registered successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func hashPassword(password string) string {
	hash := md5.Sum([]byte(password)) // md5 for now, swapping to bcrypt later
	return hex.EncodeToString(hash[:])
}


