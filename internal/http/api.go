package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/WeKnowXD/WeKnowXD_Project/internal/auth"
	"github.com/WeKnowXD/WeKnowXD_Project/internal/database"
	"github.com/WeKnowXD/WeKnowXD_Project/internal/models"
	"github.com/WeKnowXD/WeKnowXD_Project/internal/search"
	"github.com/WeKnowXD/WeKnowXD_Project/internal/weather"
)

// the different ways a session check can fail, getSession returns one of these
var (
	errNoSession      = errors.New("no session cookie")
	errSessionUnknown = errors.New("session not found")
	errSessionExpired = errors.New("session expired")
)

// the only place that reads a session, checks the cookie exists, the token is known and it hasn't expired
func getSession(r *http.Request) (models.SessionData, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return models.SessionData{}, errNoSession
	}

	auth.SessionsMutex.Lock()
	session, ok := auth.Sessions[cookie.Value]
	if ok {
		session.LastSeen = time.Now()
		auth.Sessions[cookie.Value] = session
	}
	auth.SessionsMutex.Unlock()

	if !ok {
		return models.SessionData{}, errSessionUnknown
	}

	if time.Now().After(session.ExpiresAt) {
		auth.SessionsMutex.Lock()
		delete(auth.Sessions, cookie.Value) // drop it here so old sessions don't pile up
		auth.SessionsMutex.Unlock()
		return models.SessionData{}, errSessionExpired
	}

	return session, nil
}

// tells the caller if their session cookie is still valid and who it belongs to
func apiSession(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, models.AuthResponse{StatusCode: http.StatusUnauthorized, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, models.AuthResponse{StatusCode: http.StatusOK, Message: session.Username})
}

// only exposes the number of distinct logged-in users active in the last 10 minutes
func apiActiveUsers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"count": auth.ActiveUsers(10 * time.Minute),
	})
}

func apiLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, models.HTTPValidationError{
			Detail: []models.ValidationError{{
				Loc:  []any{"body", 0},
				Msg:  "Invalid form data or a bad request",
				Type: "value_error.missing",
			}},
		})
		return
	}

	var missing []models.ValidationError
	for _, f := range []string{"username", "password"} {
		if _, ok := r.PostForm[f]; !ok {
			missing = append(missing, models.ValidationError{Loc: []any{"body", f}, Msg: "Field required", Type: "missing"})
		}
	}
	if len(missing) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, models.HTTPValidationError{Detail: missing})
		return
	}

	userList, err := database.QueryDB(reflect.TypeOf(models.User{}), "SELECT * FROM users WHERE username = $1", r.FormValue("username"))
	if err != nil {
		log.Println(err)
		writeJSON(w, http.StatusInternalServerError, models.AuthResponse{StatusCode: http.StatusInternalServerError, Message: "server error"})
		return
	}
	if len(userList) == 0 {
		writeJSON(w, http.StatusUnauthorized, models.AuthResponse{StatusCode: http.StatusUnauthorized, Message: "Invalid username"})
		return
	}

	foundUser := userList[0].(models.User)

	if auth.VerifyPassword(foundUser.Password, r.FormValue("password")) == false {
		writeJSON(w, http.StatusUnauthorized, models.AuthResponse{StatusCode: http.StatusUnauthorized, Message: "Invalid password"})
		return
	}

	token, err := auth.GenerateSessionToken()
	if err != nil {
		log.Println(err)
		writeJSON(w, http.StatusInternalServerError, models.AuthResponse{StatusCode: http.StatusInternalServerError, Message: "server error"})
		return
	}

	auth.SessionsMutex.Lock()
	auth.Sessions[token] = models.SessionData{
		UserID:    foundUser.Id,
		Username:  foundUser.Username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		LastSeen:  time.Now(),
	}
	auth.SessionsMutex.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	})

	fmt.Println("User", foundUser.Username, "is currently logged in with session")
	writeJSON(w, http.StatusOK, models.AuthResponse{StatusCode: http.StatusOK, Message: "Login successful"})
}

func logoutSession(r *http.Request) (int, string) {
	cookie, err := r.Cookie("session_token")
	if errors.Is(err, http.ErrNoCookie) {
		return http.StatusBadRequest, "no cookie was found"
	} else if err != nil {
		log.Println(err)
		return http.StatusInternalServerError, "server error"
	}

	auth.SessionsMutex.Lock()
	if _, ok := auth.Sessions[cookie.Value]; !ok {
		log.Println("logout: unknown session") // still clear the cookie below so the user ends up logged out
	}
	delete(auth.Sessions, cookie.Value)
	auth.SessionsMutex.Unlock()

	return http.StatusOK, "Logout successful"
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

func apiLogout(w http.ResponseWriter, r *http.Request) {
	status, msg := logoutSession(r)
	if status != http.StatusOK {
		writeJSON(w, status, models.AuthResponse{StatusCode: status, Message: msg})
		return
	}

	clearSessionCookie(w)
	writeJSON(w, status, models.AuthResponse{StatusCode: status, Message: msg})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	logoutSession(r) // redirect even if there was no cookie
	clearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	language := r.URL.Query().Get("language")

	results, err := search.Get(database.DB, q, language)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, models.SearchResponse{Data: results})
}

func apiRegister(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

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
		var existingID int
		err := database.DB.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&existingID)
		if err == nil {
			errorMsg = "Username already taken"
		}
	}

	if errorMsg != "" {
		writeJSON(w, http.StatusUnprocessableEntity, models.HTTPValidationError{
			Detail: []models.ValidationError{{
				Loc:  []any{"body", 0},
				Msg:  errorMsg,
				Type: "value_error.UnprocessableEntity",
			}},
		})
		return
	}

	hashedPassword := auth.HashPassword(password)

	_, err := database.DB.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)", username, email, hashedPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"statusCode": http.StatusOK,
		"message":    "Registered successfully",
	})
}

func apiWeather(w http.ResponseWriter, r *http.Request) {
	data, err := weather.Get()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"data": map[string]any{
				"error": "Could not fetch the weather forecast right now",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}
