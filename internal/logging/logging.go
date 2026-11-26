package logging

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

// sends all logs as JSON to both the terminal and the log file
func Init(path string) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatal(err)
		}
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}

	writer := io.MultiWriter(os.Stdout, file)
	slog.SetDefault(slog.New(slog.NewJSONHandler(writer, nil)))
}
