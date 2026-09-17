package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	"log"
	"net/http"

	"time"

	"strings"

	"sync"

	"html/template"

	"fmt"

	"reflect"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB // shared connection, every handler in this file can just use this directly
var templates = template.Must(template.ParseFiles("templates/search.html", "templates/register.html", "templates/layout.html"))

func main() {
	dataBase := newDataBase()
	defer dataBase.Close()

	server := newServer(dataBase)

	http.ListenAndServe(":8080", server.Mux)
}

func (server *Server) router() {
	server.Mux.HandleFunc("POST /api/login", server.apiLogin)
	server.Mux.HandleFunc("GET /api/test", server.apiTestDB)

	fileServer := http.FileServer(http.Dir("./templates"))

	server.Mux.Handle("/templates/", http.StripPrefix("/templates/", fileServer))

	server.Mux.Handle("/", fileServer)
}

func (server *Server) apiTestDB(w http.ResponseWriter, r *http.Request) {
	userList := server.queryDB(reflect.TypeOf(User{}), "SELECT * FROM users")

	if len(userList) == 0 {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("database connected"))
		return
	}

	firstUser := userList[0].(User)
	println("Test succes, found user:", firstUser.Username)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userList)
}

type Server struct {
	Mux           *http.ServeMux
	DB            *sql.DB
	Sessions      map[string]int
	SessionsMutex sync.RWMutex
}

type User struct {
	Id       int
	Username string
	Email    string
	Password string
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

func newServer(dataBase *sql.DB) *Server {
	server := &Server{
		DB:       dataBase,
		Mux:      http.NewServeMux(),
		Sessions: make(map[string]int),
	}

	server.router()
	return server
}

func newDataBase() *sql.DB {
	db, err := sql.Open("sqlite3", "../whoknows.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Successfully connected to SQLite database")

	return db
}

func generateSessionToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (server *Server) apiLogin(w http.ResponseWriter, r *http.Request) {
	error := r.ParseForm()
	if error != nil {
		log.Fatal(error)
	}

	userList := server.queryDB(reflect.TypeOf(User{}), "SELECT * FROM users WHERE username = ?", r.FormValue("username"))
	if len(userList) == 0 {
		log.Println("func: apiLogin, queryDB returned empty userlist when seaching for username: " + r.FormValue("username"))
	}

	foundUser := userList[0].(User)

	if verifyPassword(foundUser.Password, r.FormValue("password")) == false {
		log.Println("func: apiLogin, user verifacation password missmatch")
	} else {
		token := generateSessionToken()

		server.SessionsMutex.Lock()
		server.Sessions[token] = foundUser.Id
		server.SessionsMutex.Unlock()

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			Path:     "/",
			HttpOnly: true,
		})

		println("User", foundUser.Username, "is currently logged in with session")
		w.Write([]byte("Login succesfull"))
	}
}

func (server *Server) queryDB(interchangeableStruct reflect.Type, query string, args ...any) []any {
	var structArray []any
	rows, error := server.DB.Query(query, args...)

	if error != nil {
		log.Fatal(error)
	}
	defer rows.Close()

	for rows.Next() {
		newStructPointer := reflect.New(interchangeableStruct)
		structElement := newStructPointer.Elem()

		numCols := structElement.NumField()
		columns := make([]any, numCols)

		for i := range numCols {
			field := structElement.Field(i)
			columns[i] = field.Addr().Interface()
		}

		error := rows.Scan(columns...)
		if error != nil {
			log.Fatal(error)
		}

		structArray = append(structArray, structElement.Interface())
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return structArray
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
	passwordBytes := []byte(password)
	passwordHash := md5.Sum(passwordBytes)
	passwordHashString := hex.EncodeToString(passwordHash[:])

	return passwordHashString
}

func verifyPassword(storedHash string, password string) bool {
	passwordHash := hashPassword(password)
	return storedHash == passwordHash
}
