package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func main() {
	//fmt.Println("Hello, world!")
	router()
}

func router() {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/search", getSearch)
	router.HandleFunc("POST /api/register", postRegister)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println(err)
	}
}

// /Need to figure out of to handle the HTML because i re-write the page route.
func getSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	_ = q // Use the query parameter 'q' as needed
	language := r.URL.Query().Get("language")
	if language == "" {
		language = "en"
	}

	/// DB logic goes above this. Need to figure out how to set up DB, then ill connect it
	/// Talk about DB setup
	response := map[string]any{
		"data": []map[string]any{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func postRegister(w http.ResponseWriter, r *http.Request) {
	// The spec says this route expects form-urlencoded data (like an HTML
	// form submit), NOT JSON. So we use r.FormValue instead of json.Decode.
	// r.FormValue reads either URL query params or form-body fields —
	// here it'll be reading from the POST body since that's where the
	// client is expected to send it.
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	password2 := r.FormValue("password2") // optional per the spec, but legacy code checks it — keeping for now

	// Same validation logic as the legacy Python version.
	var errorMsg string
	if username == "" {
		errorMsg = "You have to enter a username"
	} else if email == "" || !strings.Contains(email, "@") {
		errorMsg = "You have to enter a valid email address"
	} else if password == "" {
		errorMsg = "You have to enter a password"
	} else if password2 != "" && password != password2 {
		// Only check the match if password2 was actually sent,
		// since the spec doesn't require it.
		errorMsg = "The two passwords do not match"
	}

	// TODO: legacy code also checks if the username is already taken
	// (needs a DB lookup — not done yet, waiting on DB setup)

	// TODO: legacy code hashes the password before storing it
	// (not done yet — needs DB + hashing decision)

	// TODO: legacy code inserts the new user into the DB here
	// (not done yet — waiting on DB setup)

	if errorMsg != "" {
		// Response shape matches the AuthResponse schema: statusCode + message
		response := map[string]any{
			"statusCode": http.StatusBadRequest,
			"message":    errorMsg,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Success response — matches AuthResponse shape.
	// TODO: this should only actually say "success" once the DB insert
	// above is real — right now nothing is actually being saved.
	response := map[string]any{
		"statusCode": http.StatusOK,
		"message":    "Registered successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
