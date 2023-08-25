package cmd

import (
	"fmt"
	"github.com/pcanilho/gh-tidy/helpers"
	"github.com/pcanilho/gh-tidy/models"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var (
	prState []string
)

var stalePrsCmd = &cobra.Command{
	Use:     "prs",
	Aliases: []string{"pr"},
	Example: `$ gh tidy stale prs <owner/repo> -t 72h`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("at least one <owner>/<repository> needs to be provided")
		}
		view := make(map[string][]*models.GitHubPR)
		for _, repo := range args {
			if len(owner) == 0 && strings.Contains(repo, "/") {
				composite := strings.Split(repo, "/")
				owner, repo = composite[0], composite[1]
			}

			// Owner
			if len(owner) == 0 {
				return fmt.Errorf("the [owner] flag must be provided")
			}
			prs, err := ghApi.ListPRs(cmd.Context(), prState, owner, repo)
			if err != nil {
				return err
			}
			view[repo] = prs
		}
		for repo, prs := range view {
			var filteredBranches []*models.GitHubPR
			for _, pr := range prs {
				if pr.LastCommitDate.Before(time.Now().Add(-staleThreshold)) {
					filteredBranches = append(filteredBranches, pr)
				}
			}
			view[repo] = filteredBranches
		}
		out = view
		//if len(args) == 1 {
		//	repo := args[0]
		//	if strings.Contains(repo, "/") {
		//		repo = strings.Split(repo, "/")[1]
		//	}
		//	out = view[repo]
		//}
		return nil
	},
	PostRunE: func(cmd *cobra.Command, args []string) error {
		if out == nil {
			return fmt.Errorf("not results found")
		}
		view := out.(map[string][]*models.GitHubPR)
		if remove {
			for repo, prs := range view {
				if !force {
					if !helpers.Prompt(fmt.Sprintf("Close [%d] PRs in repo [%v]?", repo, len(prs))) {
						fmt.Println("cancelled...")
						return nil
					}
				}

				for _, pr := range prs {
					_, err := ghApi.ClosePRs(cmd.Context(), pr.Id)
					if err != nil {
						return fmt.Errorf("unable to delete PR[%d]: %v -> %v. error: %v", pr.Number, pr.Source, pr.Target, err)
					}
				}
			}
		}
		return nil
	},
}

func init() {
	stalePrsCmd.PersistentFlags().StringArrayVarP(&prState, "state", "s", []string{"OPEN"}, "The PR state. Supported values are: OPEN, MERGED or CLOSED")
}
