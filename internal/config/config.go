package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func Load() {
	path := ".env"
	if root, ok := projectRoot(); ok {
		path = filepath.Join(root, ".env")
	}
	if err := godotenv.Load(path); err != nil {
		log.Println("No .env file found, using normal env variables")
	}
}

func Addr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":8080"
}

func DBPath() string {
	if path := os.Getenv("DB_PATH"); path != "" {
		return path
	}
	if root, ok := projectRoot(); ok {
		return filepath.Join(root, "data", "whoknows.db")
	}
	return filepath.Join("data", "whoknows.db")
}

func projectRoot() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
