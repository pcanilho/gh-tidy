package api

import "time"

type GitHubRef struct {
	Id             string     `json:"id,omitempty" yaml:"id,omitempty"`
	Name           string     `json:"name,omitempty" yaml:"name,omitempty"`
	LastCommitDate *time.Time `json:"last_commit_date,omitempty" yaml:"last_commit_date,omitempty"`
	TagDate        *time.Time `json:"tag_date,omitempty" yaml:"tag_date"`
}

type GitHubPR struct {
	Source         string    `json:"source,omitempty" yaml:"source,omitempty"`
	Target         string    `json:"target,omitempty" yaml:"target,omitempty"`
	LastCommitDate time.Time `json:"last_commit_date" yaml:"last_commit_date"`
	Id             string    `json:"id,omitempty" yaml:"id,omitempty"`
	Number         int       `json:"number,omitempty" yaml:"number,omitempty"`
	Url            string    `json:"url,omitempty" yaml:"url,omitempty"`
}
