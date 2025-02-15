package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	github "github.com/glup3/gorepos/internal"
)

//go:embed repos.json
var content embed.FS

func main() {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	data, err := content.ReadFile("repos.json")
	if err != nil {
		slog.Error("unable to read data file", slog.Any("error", err))
		os.Exit(1)
	}

	var goData github.GoData
	if err := json.Unmarshal(data, &goData); err != nil {
		slog.Error("unable parse json", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("loaded json in memory", slog.Int("repoCount", len(goData.Data)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, goData.Data[:5])
	})

	slog.Info("server started on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("server error", slog.Any("error", err))
	}
}
