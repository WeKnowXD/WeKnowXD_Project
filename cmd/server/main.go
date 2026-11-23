package main

import (
	"net/http"

	"github.com/WeKnowXD/WeKnowXD_Project/internal/auth"
	"github.com/WeKnowXD/WeKnowXD_Project/internal/config"
	"github.com/WeKnowXD/WeKnowXD_Project/internal/database"
	apphttp "github.com/WeKnowXD/WeKnowXD_Project/internal/http"
	"github.com/WeKnowXD/WeKnowXD_Project/internal/logging"
)

func main() {
	config.Load()
	logging.Init(config.LogPath())
	auth.Init()
	database.Init(config.DBPath())
	database.InitFTS()
	mux := http.NewServeMux()
	apphttp.Router(mux)
	http.ListenAndServe(config.Addr(), mux)
}
