package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type PRItem struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	URL       string   `json:"url"`
	Author    string   `json:"author"`
	Draft     bool     `json:"draft"`
	Labels    []string `json:"labels"`
	ChecksOK  *bool    `json:"checks_ok,omitempty"`
	UpdatedAt string   `json:"updated_at"`
}

type Adapter struct {
	Owner string
	Repo  string
	Token string
	HTTP  *http.Client
}

func FromEnv() *Adapter {
	slug := os.Getenv("ARAH_GITHUB_REPO") // owner/name
	if slug == "" {
		return nil
	}
	parts := strings.SplitN(slug, "/", 2)
	if len(parts) != 2 {
		return nil
	}
	return &Adapter{
		Owner: parts[0],
		Repo:  parts[1],
		Token: os.Getenv("GITHUB_TOKEN"),
		HTTP:  &http.Client{Timeout: 12 * time.Second},
	}
}

func (a *Adapter) OpenPRs() ([]PRItem, error) {
	if a == nil {
		return nil, nil
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=open&per_page=20", a.Owner, a.Repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if a.Token != "" {
		req.Header.Set("Authorization", "Bearer "+a.Token)
	}
	resp, err := a.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github api: %s", resp.Status)
	}
	var raw []struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		HTMLURL   string `json:"html_url"`
		Draft     bool   `json:"draft"`
		UpdatedAt string `json:"updated_at"`
		User      struct {
			Login string `json:"login"`
		} `json:"user"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]PRItem, 0, len(raw))
	for _, p := range raw {
		labels := make([]string, 0, len(p.Labels))
		for _, l := range p.Labels {
			labels = append(labels, l.Name)
		}
		out = append(out, PRItem{
			Number:    p.Number,
			Title:     p.Title,
			URL:       p.HTMLURL,
			Author:    p.User.Login,
			Draft:     p.Draft,
			Labels:    labels,
			UpdatedAt: p.UpdatedAt,
		})
	}
	return out, nil
}
