package github

import (
	"fmt"
	"time"
)

type GoData struct {
	Data []GoRepo `json:"data"`
}

type GoRepo struct {
	ID              int      `json:"id"`
	NodeID          string   `json:"node_id"`
	FullName        string   `json:"full_name"`
	Description     string   `json:"description"`
	AvatarURL       string   `json:"avatar_url"`
	StargazersCount int      `json:"stargazers_count"`
	Archived        bool     `json:"archived"`
	LicenseSpdxID   string   `json:"license_spdx_id"`
	CreatedAt       JSONTime `json:"created_at"`
	Topics          []string `json:"topics"`
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) error {
	parsedT, err := time.Parse(time.RFC3339, string(data[1:21]))
	if err != nil {
		return err
	}

	*t = JSONTime(parsedT)
	return nil
}
