package cmd

import (
	"fmt"
	"github.com/pcanilho/gh-tidy/helpers"
	"github.com/pcanilho/gh-tidy/models"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var staleBranchesCmd = &cobra.Command{
	Use:     "branches",
	Aliases: []string{"b", "br"},
	Example: `$ gh tidy stale branches <owner/repo> -t 72h`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("at least one <owner>/<repository> needs to be provided")
		}
		view := make(map[string][]*models.GitHubBranch)
		for _, repo := range args {
			if len(owner) == 0 && strings.Contains(repo, "/") {
				composite := strings.Split(repo, "/")
				owner, repo = composite[0], composite[1]
			}

			// Owner
			if len(owner) == 0 {
				return fmt.Errorf("the [owner] flag must be provided")
			}
			brs, err := ghApi.ListBranches(cmd.Context(), owner, repo)
			if err != nil {
				return err
			}
			view[repo] = brs
		}

		for repo, branches := range view {
			var filteredBranches []*models.GitHubBranch
			for _, branch := range branches {
				if excludeRegex != nil && excludeRegex.MatchString(branch.Name) {
					continue
				}

				if branch.LastCommitDate.Before(time.Now().Add(-staleThreshold)) {
					filteredBranches = append(filteredBranches, branch)
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
			return fmt.Errorf("no results found")
		}
		view := out.(map[string][]*models.GitHubBranch)
		if remove {
			for repo, branches := range view {
				if !force {
					if !helpers.Prompt(fmt.Sprintf("Delete [%d] branches in repo [%v]?", len(branches), repo)) {
						continue
					}
				}

				for _, branch := range branches {
					_, err := ghApi.DeleteRefs(cmd.Context(), branch.Id)
					if err != nil {
						return fmt.Errorf("unable to delete branch: %v. error: %v", branch.Name, err)
					}
				}
			}
		}
		return nil
	},
}

func init() {
	staleBranchesCmd.PersistentFlags().StringVar(&excludePattern, "exclude", "", "If provided, it will be used to exclude branches that match the pattern (regexp)")
}
