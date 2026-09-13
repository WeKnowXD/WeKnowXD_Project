package main

import (
	"crypto/md5"
	"encoding/hex"

	"net/http"

	"log"

	"reflect"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	mux := http.NewServeMux()

	htmlPages := http.FileServer(http.Dir("./templates"))

	mux.Handle("/", htmlPages)

	mux.HandleFunc("GET /login", login)
	mux.HandleFunc("POST /api/login", apiLogin)

	mux.HandleFunc("/api/logout", apiLogout)

	http.ListenAndServe(":8080", mux)
}

type User struct {
	id       int
	username string
	email    string
	password string
}

func login(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./templates/login.html")
}

func apiLogin(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	error := r.ParseForm()
	if error != nil {
		log.Fatal(error)
	}

	user := queryDB(reflect.TypeOf(User{}), db, "SELECT * FROM users WHERE username = "+r.FormValue("username"))
	if len(user) == 0 {
		log.Println("func: apiLogin, queryDB returned empty userlist when seaching for username: " + r.FormValue("username"))
	} else if verifyPassword(user[0].password, r.FormValue("password")) == false {
		log.Println("func: apiLogin, user verifacation password missmatch")
	} else {
		//Add session logic
	}
}

func apiLogout(w http.ResponseWriter, r *http.Request) {
	//TODO: add session handleling, stop user session.

	http.ServeFile(w, r, "./templates/search.html")
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

func queryDB(interchangeableStruct reflect.Type, db *sql.DB, query string) []any {
	var structArray []any
	rows, error := db.Query(query)

	if error != nil {
		log.Fatal(error)
	}
	defer rows.Close()

	for rows.Next() {
		structElement := reflect.ValueOf(&interchangeableStruct).Elem()
		numCols := structElement.NumField()
		columns := make([]any, numCols)
		for i := 0; i < numCols; i++ {
			field := structElement.Field(i)
			columns[i] = field.Addr().Interface()
		}

		error := rows.Scan(columns...)
		if error != nil {
			log.Fatal(error)
		}

		structArray = append(structArray, structElement)
	}

	return structArray
}
