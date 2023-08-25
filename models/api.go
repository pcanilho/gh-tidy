package models

import "time"

type GitHubRef struct {
	Id             string     `json:"id,omitempty"`
	Name           string     `json:"name,omitempty"`
	LastCommitDate *time.Time `json:"last_commit_date"`
	TagDate        *time.Time `json:"tag_date"`
}

type GitHubPR struct {
	Id             string    `json:"id,omitempty"`
	Source         string    `json:"source,omitempty"`
	Target         string    `json:"target,omitempty"`
	LastCommitDate time.Time `json:"last_commit_date"`
	Number         int       `json:"number,omitempty"`
	Url            string    `json:"url,omitempty"`
}
