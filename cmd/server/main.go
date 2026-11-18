package main

import (
	"net/http"

	"github.com/WeKnowXD/WeKnowXD_Project/internal/auth"
	"github.com/WeKnowXD/WeKnowXD_Project/internal/config"
	"github.com/WeKnowXD/WeKnowXD_Project/internal/database"
	apphttp "github.com/WeKnowXD/WeKnowXD_Project/internal/http"
)

func main() {
	config.Load()
	auth.Init()
	database.Init(config.DBPath())
	mux := http.NewServeMux()
	apphttp.Router(mux)
	http.ListenAndServe(config.Addr(), mux)
}
