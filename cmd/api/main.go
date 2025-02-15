package main

import (
	"html/template"
	"log/slog"
	"net/http"
)

func main() {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	})

	slog.Info("server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error("server error", slog.Any("error", err))
	}
}
