package models

import "time"

type GitHubBranch struct {
	Id             string
	Name           string
	LastCommitDate time.Time
}

type GitHubPR struct {
	Id             string
	Source, Target string
	LastCommitDate time.Time
	Number         int
	Url            string
}
