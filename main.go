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
	slog.Info("wow", "hi", 57100%1000)

	tmpl := template.Must(template.New("index").
		Funcs(template.FuncMap{"formatStars": formatStars}).
		ParseFiles("templates/index.html"))

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
		err := tmpl.Execute(w, goData.Data[:16])
		if err != nil {
			slog.Error("unable to render", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("GET /repos/discover", func(w http.ResponseWriter, r *http.Request) {
		repos, err := getRandomItems(goData.Data, 16, time.Now().UnixNano())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "items", repos)
		if err != nil {
			slog.Error("unable to render", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

func formatStars(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	rest := n % 1000
	if rest >= 100 {
		return fmt.Sprintf("%d.%dK", n/1000, rest/100)
	}

	return fmt.Sprintf("%dK", n/1000)
}
