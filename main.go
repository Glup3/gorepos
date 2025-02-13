package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

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

	var totalRepos []GoRepo
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

		resp, err := getRepos(perPage, currentPage, minStars, maxStars)
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

			totalRepos = append(totalRepos, GoRepo{
				ID:              repo.ID,
				NodeID:          repo.NodeID,
				FullName:        repo.FullName,
				AvatarURL:       repo.Owner.AvatarURL,
				StargazersCount: repo.StargazersCount,
				Archived:        repo.Archived,
				LicenseSpdxID:   repo.License.SpdxID,
				Topics:          repo.Topics,
				CreatedAt:       JSONTime(repo.CreatedAt),
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

	jsonData, err := json.MarshalIndent(GoData{Data: totalRepos}, "", "  ")
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

func githubQuery(perPage, page, minStars, maxStars int) string {
	u, _ := url.Parse("https://api.github.com/search/repositories")

	query := url.Values{}
	query.Set("q", fmt.Sprintf("language:%s stars:%d...%d", "go", minStars, maxStars))
	query.Set("sort", "stars")
	query.Set("order", "desc")
	query.Set("per_page", strconv.Itoa(perPage))
	query.Set("page", strconv.Itoa(page))

	u.RawQuery = query.Encode()

	return u.String()
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))), nil
}

type GoData struct {
	Data []GoRepo `json:"data"`
}

type GoRepo struct {
	ID              int      `json:"id"`
	NodeID          string   `json:"node_id"`
	FullName        string   `json:"full_name"`
	AvatarURL       string   `json:"avatar_url"`
	StargazersCount int      `json:"stargazers_count"`
	Archived        bool     `json:"archived"`
	LicenseSpdxID   string   `json:"license_spdx_id"`
	Topics          []string `json:"topics"`
	CreatedAt       JSONTime `json:"created_at"`
}

type GitHubRepo struct {
	ID       int    `json:"id"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Owner    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		UserViewType      string `json:"user_view_type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"owner"`
	HTMLURL          string    `json:"html_url"`
	Description      string    `json:"description"`
	Fork             bool      `json:"fork"`
	URL              string    `json:"url"`
	ForksURL         string    `json:"forks_url"`
	KeysURL          string    `json:"keys_url"`
	CollaboratorsURL string    `json:"collaborators_url"`
	TeamsURL         string    `json:"teams_url"`
	HooksURL         string    `json:"hooks_url"`
	IssueEventsURL   string    `json:"issue_events_url"`
	EventsURL        string    `json:"events_url"`
	AssigneesURL     string    `json:"assignees_url"`
	BranchesURL      string    `json:"branches_url"`
	TagsURL          string    `json:"tags_url"`
	BlobsURL         string    `json:"blobs_url"`
	GitTagsURL       string    `json:"git_tags_url"`
	GitRefsURL       string    `json:"git_refs_url"`
	TreesURL         string    `json:"trees_url"`
	StatusesURL      string    `json:"statuses_url"`
	LanguagesURL     string    `json:"languages_url"`
	StargazersURL    string    `json:"stargazers_url"`
	ContributorsURL  string    `json:"contributors_url"`
	SubscribersURL   string    `json:"subscribers_url"`
	SubscriptionURL  string    `json:"subscription_url"`
	CommitsURL       string    `json:"commits_url"`
	GitCommitsURL    string    `json:"git_commits_url"`
	CommentsURL      string    `json:"comments_url"`
	IssueCommentURL  string    `json:"issue_comment_url"`
	ContentsURL      string    `json:"contents_url"`
	CompareURL       string    `json:"compare_url"`
	MergesURL        string    `json:"merges_url"`
	ArchiveURL       string    `json:"archive_url"`
	DownloadsURL     string    `json:"downloads_url"`
	IssuesURL        string    `json:"issues_url"`
	PullsURL         string    `json:"pulls_url"`
	MilestonesURL    string    `json:"milestones_url"`
	NotificationsURL string    `json:"notifications_url"`
	LabelsURL        string    `json:"labels_url"`
	ReleasesURL      string    `json:"releases_url"`
	DeploymentsURL   string    `json:"deployments_url"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PushedAt         time.Time `json:"pushed_at"`
	GitURL           string    `json:"git_url"`
	SSHURL           string    `json:"ssh_url"`
	CloneURL         string    `json:"clone_url"`
	SvnURL           string    `json:"svn_url"`
	Homepage         string    `json:"homepage"`
	Size             int       `json:"size"`
	StargazersCount  int       `json:"stargazers_count"`
	WatchersCount    int       `json:"watchers_count"`
	Language         string    `json:"language"`
	HasIssues        bool      `json:"has_issues"`
	HasProjects      bool      `json:"has_projects"`
	HasDownloads     bool      `json:"has_downloads"`
	HasWiki          bool      `json:"has_wiki"`
	HasPages         bool      `json:"has_pages"`
	HasDiscussions   bool      `json:"has_discussions"`
	ForksCount       int       `json:"forks_count"`
	MirrorURL        any       `json:"mirror_url"`
	Archived         bool      `json:"archived"`
	Disabled         bool      `json:"disabled"`
	OpenIssuesCount  int       `json:"open_issues_count"`
	License          struct {
		Key    string `json:"key"`
		Name   string `json:"name"`
		SpdxID string `json:"spdx_id"`
		URL    string `json:"url"`
		NodeID string `json:"node_id"`
	} `json:"license"`
	AllowForking             bool     `json:"allow_forking"`
	IsTemplate               bool     `json:"is_template"`
	WebCommitSignoffRequired bool     `json:"web_commit_signoff_required"`
	Topics                   []string `json:"topics"`
	Visibility               string   `json:"visibility"`
	Forks                    int      `json:"forks"`
	OpenIssues               int      `json:"open_issues"`
	Watchers                 int      `json:"watchers"`
	DefaultBranch            string   `json:"default_branch"`
	Score                    float64  `json:"score"`
}

type GitHubResponse struct {
	Items      []GitHubRepo `json:"items"`
	TotalCount int          `json:"total_count"`
}

func getRepos(perPage, page, minStars, maxStars int) (*GitHubResponse, error) {
	githubToken, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		return nil, errors.New("loading github token failed")
	}

	req, _ := http.NewRequest("GET", githubQuery(perPage, page, minStars, maxStars), nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+githubToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected response status code 200, but given %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result GitHubResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
