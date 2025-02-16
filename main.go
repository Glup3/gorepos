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
	"sync"

	github "github.com/glup3/gorepos/internal"
)

var (
	//go:embed repos.json
	content embed.FS

	//go:embed assets
	assets embed.FS
)

var pool = sync.Pool{
	New: func() interface{} {
		return make([]github.GoRepo, 0, 16)
	},
}

func main() {
	tmpl := template.Must(template.New("index").
		Funcs(template.FuncMap{"formatStars": formatStars}).
		ParseFiles("templates/index.html"))

	file, err := content.Open("repos.json")
	if err != nil {
		slog.Error("unable to read data file", slog.Any("error", err))
		os.Exit(1)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var goData github.GoData
	if err := decoder.Decode(&goData); err != nil {
		slog.Error("unable parse json", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("loaded json in memory", slog.Int("repoCount", len(goData.Data)))

	mux := http.NewServeMux()
	mux.Handle("GET /assets/", http.FileServer(http.FS(assets)))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, map[string]interface{}{
			"DiscoveryRepos": goData.Data[:16],
		})
		if err != nil {
			slog.Error("unable to render", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("GET /repos/discover", func(w http.ResponseWriter, r *http.Request) {
		repos, err := getRandomItems(goData.Data, 16)
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

		pool.Put(repos[:0])
	})

	slog.Info("server started on :8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		slog.Error("server error", slog.Any("error", err))
	}
}

func getRandomItems(arr []github.GoRepo, k int) ([]github.GoRepo, error) {
	n := len(arr)
	if k > n {
		return nil, fmt.Errorf("k cannot be larger than the array size")
	}

	result := pool.Get().([]github.GoRepo)
	result = result[:0] // Reset length (keep capacity)

	perm := rand.Perm(n)[:k]
	for _, idx := range perm {
		result = append(result, arr[idx])
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
