package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	github "github.com/glup3/gorepos/internal"
)

var (
	//go:embed repos.json
	content embed.FS

	//go:embed assets/css/output.css
	css embed.FS
)

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

	mux := http.NewServeMux()
	mux.Handle("GET /assets/css/output.css", http.FileServer(http.FS(css)))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, goData.Data[:16])
	})

	mux.HandleFunc("GET /repos/random", func(w http.ResponseWriter, r *http.Request) {
		repos, err := getRandomItems(goData.Data, 16, time.Now().UnixNano())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "items", repos)
	})

	slog.Info("server started on :8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		slog.Error("server error", slog.Any("error", err))
	}
}

func getRandomItems[T any](arr []T, k int, seed int64) ([]T, error) {
	n := len(arr)
	if k > n {
		return nil, fmt.Errorf("k cannot be larger than the array size")
	}

	rng := rand.New(rand.NewSource(seed))
	selectedIndices := make(map[int]struct{})
	result := make([]T, 0, k)

	for len(selectedIndices) < k {
		index := rng.Intn(n)
		if _, exists := selectedIndices[index]; !exists {
			selectedIndices[index] = struct{}{}
			result = append(result, arr[index])
		}
	}

	return result, nil
}
