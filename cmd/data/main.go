package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"sort"
	"time"

	github "github.com/glup3/gorepos/internal"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("loading .env", slog.Any("error", err))
		os.Exit(1)
	}

	perPage := 100
	currentPage := 1
	lastPage := 1000 / perPage
	minStars := 200
	maxStars := 1_000_000

	currentIteration := 1
	const maxIterations = 200

	var totalRepos []github.GoRepo
	seenRepos := make(map[int]bool)

	for {
		slog.Info(
			"fetching repos",
			slog.Int("page", currentPage),
			slog.Int("perPage", perPage),
			slog.Int("minStars", minStars),
			slog.Int("maxStars", maxStars),
			slog.Int("iteration", currentIteration),
		)

		resp, err := github.GetRepos(perPage, currentPage, minStars, maxStars)
		if err != nil {
			slog.Error(
				"failed fetching repos",
				slog.Any("error", err),
				slog.Int("page", currentPage),
				slog.Int("perPage", perPage),
				slog.Int("minStars", minStars),
				slog.Int("maxStars", maxStars),
				slog.Int("iteration", currentIteration),
			)
			os.Exit(1)
		}

		for i, repo := range resp.Items {
			if seenRepos[repo.ID] {
				continue
			}

			totalRepos = append(totalRepos, github.GoRepo{
				ID:              repo.ID,
				NodeID:          repo.NodeID,
				FullName:        repo.FullName,
				Description:     repo.Description,
				AvatarURL:       repo.Owner.AvatarURL,
				StargazersCount: repo.StargazersCount,
				Archived:        repo.Archived,
				LicenseSpdxID:   repo.License.SpdxID,
				CreatedAt:       github.JSONTime(repo.CreatedAt),
				Topics:          repo.Topics,
			})
			seenRepos[repo.ID] = true

			if currentPage == lastPage && i == len(resp.Items)-1 {
				maxStars = repo.StargazersCount
			}
		}

		currentIteration++

		if len(resp.Items) == 0 {
			break
		}

		if currentIteration >= maxIterations {
			slog.Warn("max iterations exceeded", slog.Int("maxIterations", maxIterations))
			break
		}

		if currentPage == lastPage {
			if maxStars <= minStars {
				slog.Warn(
					"there are still more entries after the last page",
					slog.Int("minStars", minStars),
					slog.Int("maxStars", maxStars),
				)
			}

			currentPage = 1
			continue
		}

		currentPage++
	}

	sort.Slice(totalRepos, func(i, j int) bool {
		if totalRepos[i].StargazersCount != totalRepos[j].StargazersCount {
			return totalRepos[i].StargazersCount > totalRepos[j].StargazersCount
		}
		return time.Time(totalRepos[i].CreatedAt).Before(time.Time(totalRepos[j].CreatedAt))
	})

	file, err := os.Create("repos.json")
	if err != nil {
		slog.Error("creating json", slog.Any("error", err))
		os.Exit(1)
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(github.GoData{Data: totalRepos}, "", "  ")
	if err != nil {
		slog.Error("marshaling json", slog.Any("error", err))
		os.Exit(1)
	}

	if _, err := file.Write(jsonData); err != nil {
		slog.Error("writing json", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("successfully created json with go repos", slog.Int("count", len(totalRepos)))
}
