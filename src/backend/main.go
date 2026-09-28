package main

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"

	"log"
	"net/http"

	"time"

	"fmt"

	"strings"

	"sync"

	"html/template"

	"reflect"

	"database/sql"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB // shared connection, every handler in this file can just use this directly
var templates = template.Must(template.ParseFiles(
	"templates/search.html",
	"templates/register.html",
	"templates/layout.html",
	"templates/login.html",
	"templates/weather.html"))

var (
	Sessions      map[string]SessionData
	SessionsMutex sync.RWMutex
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using normal env variables")
	}

	Sessions = make(map[string]SessionData)
	initDB()
	mux := http.NewServeMux()
	router(mux)
	http.ListenAndServe(":8080", mux)
}

func router(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", layoutHandler)
	mux.HandleFunc("GET /register", registerHandler)
	mux.HandleFunc("GET /search", searchHandler)
	mux.HandleFunc("GET /api/search", getSearch)
	mux.HandleFunc("GET /login", loginHandler)
	mux.HandleFunc("GET /api/weather", apiWeather)

	mux.HandleFunc("GET /weather", weatherHandler)

	mux.HandleFunc("POST /api/register", postRegister)
	mux.HandleFunc("POST /api/login", apiLogin)
	mux.HandleFunc("GET /api/logout", apiLogout)
	mux.HandleFunc("POST /test", testSessions)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

func testSessions(w http.ResponseWriter, r *http.Request) {
	fmt.Println(Sessions)
}

type AuthResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type SessionData struct {
	UserID    int
	Username  string
	ExpiresAt time.Time
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

type LoginData struct {
	Error    string
	Username string
}

// TODO: currently passing nil since we don't have session/auth handling yet.
// Once that's built, replace this with a struct (e.g. LayoutData) holding
// User (nil if not logged in) and Flashes ([]string), so layout.html's
// {{ if .User }} and {{ if .Flashes }} blocks actually have data to work with.
func layoutHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "layout.html", nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{Query: r.URL.Query().Get("q")}
	templates.ExecuteTemplate(w, "layout.html", data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "register.html", RegisterData{})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", LoginData{})
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "weather.html", LoginData{})
}

func generateSessionToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func apiLogin(w http.ResponseWriter, r *http.Request) {
	error := r.ParseForm()
	if error != nil {
		log.Fatal(error)
		return
	}

	userList := queryDB(reflect.TypeOf(User{}), "SELECT * FROM users WHERE username = ?", r.FormValue("username"))
	if len(userList) == 0 {
		log.Println("func: apiLogin, queryDB returned empty userlist when seaching for username: " + r.FormValue("username"))
		w.WriteHeader(http.StatusUnauthorized)
		templates.ExecuteTemplate(w, "login.html", LoginData{

			Error:    "Invalid username or password",
			Username: r.FormValue("username")})
		return
	}

	foundUser := userList[0].(User)

	if verifyPassword(foundUser.Password, r.FormValue("password")) == false {
		log.Println("func: apiLogin, user verifacation password missmatch")
		w.WriteHeader(http.StatusUnauthorized)
		templates.ExecuteTemplate(w, "login.html", LoginData{

			Error:    "Invalid username or password",
			Username: r.FormValue("username"),
		})
		return
	} else {
		token := generateSessionToken()

		SessionsMutex.Lock()
		Sessions[token] = SessionData{
			UserID:    foundUser.Id,
			Username:  foundUser.Username,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		SessionsMutex.Unlock()

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

func apiLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if errors.Is(err, http.ErrNoCookie) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "no cookie was found",
		})
		return
	} else if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "server error",
		})
		return
	}

	sessionToken := cookie.Value

	SessionsMutex.Lock()
	delete(Sessions, sessionToken)
	SessionsMutex.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		StatusCode: http.StatusOK,
		Message:    "Logout successful",
	})
}

func queryDB(interchangeableStruct reflect.Type, query string, args ...any) []any {
	var structArray []any
	rows, error := db.Query(query, args...)

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
	passwordBytes := []byte(password)
	passwordHash := md5.Sum(passwordBytes)
	passwordHashString := hex.EncodeToString(passwordHash[:])

	return passwordHashString
}

func verifyPassword(storedHash string, password string) bool {
	passwordHash := hashPassword(password)
	return storedHash == passwordHash
}

// the last forecast we got from weatherapi, shared between all requests, so we don't get fucked and run out
var (
	weatherCache      map[string]any
	weatherCachedAt   time.Time
	weatherCacheMutex sync.Mutex // requests run at the same time, so only one can touch the cache at once
)

// how long we keep using the same forecast before asking weatherapi again
const weatherCacheDuration = 30 * time.Minute

// calls weatherapi and returns the forecast, doesn't know anything about the request/response to our own users
func fetchWeather() (map[string]any, error) {
	apiKey := os.Getenv("WEATHER_API_KEY") // comes from .env locally, github secrets later
	url := "https://api.weatherapi.com/v1/forecast.json?key=" + apiKey + "&q=Copenhagen&days=3"

	client := http.Client{Timeout: 10 * time.Second} // normal http.Get has no timeout and could hang forever
	resp, err := client.Get(url)
	if err != nil { // no answer at all, like no internet or timeout
		return nil, err
	}
	defer resp.Body.Close() // body is still an open connection, has to be closed

	// weatherapi did answer but with an error, like bad key or out of calls
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weatherapi returned status %d", resp.StatusCode)
	}

	// turn the json from weatherapi into a go map
	var weatherData map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return nil, err
	}

	return weatherData, nil
}

func apiWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// lock so two requests don't mess with the cache at the same time
	// also means if 50 requests come in when the cache is old, only the first one calls weatherapi
	weatherCacheMutex.Lock()
	defer weatherCacheMutex.Unlock()

	// only ask weatherapi if we have nothing saved or what we have is too old
	if weatherCache == nil || time.Since(weatherCachedAt) > weatherCacheDuration {
		log.Println("func: apiWeather, cache empty or expired, fetching new forecast")

		weatherData, err := fetchWeather()
		if err != nil {
			log.Println("func: apiWeather, fetching weather failed:", err)

			if weatherCache == nil { // nothing old to fall back on
				w.WriteHeader(http.StatusBadGateway) // 502, our server is fine but weatherapi failed
				json.NewEncoder(w).Encode(map[string]any{
					"data": map[string]any{
						"error": "Could not fetch the weather forecast right now",
					},
				})
				return
			}
			// otherwise we just keep using the old forecast below
			// cachedAt isn't updated so the next request tries again
		} else {
			weatherCache = weatherData
			weatherCachedAt = time.Now()
		}
	}

	// spec wants it wrapped in "data"
	json.NewEncoder(w).Encode(map[string]any{"data": weatherCache})
}
