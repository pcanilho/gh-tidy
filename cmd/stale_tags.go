package cmd

import (
	"fmt"
	"github.com/pcanilho/gh-tidy/api"
	"github.com/pcanilho/gh-tidy/helpers"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var staleTagsCmd = &cobra.Command{
	Use:     "tags",
	Aliases: []string{"t"},
	Example: `$ gh tidy stale tags <owner/repo> -t 72h`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("at least one <owner>/<repository> needs to be provided")
		}
		view := make(map[string][]*api.GitHubRef)
		for _, repo := range args {
			if len(owner) == 0 && strings.Contains(repo, "/") {
				composite := strings.Split(repo, "/")
				owner, repo = composite[0], composite[1]
			}

			// Owner
			if len(owner) == 0 {
				return fmt.Errorf("the [owner] flag must be provided")
			}
			tags, err := ghApi.ListRefs(cmd.Context(), owner, repo, api.TagRefType)
			if err != nil {
				return err
			}
			view[repo] = tags
		}

		for repo, tags := range view {
			var filteredTags []*api.GitHubRef
			for _, tag := range tags {
				if excludeRegex != nil && excludeRegex.MatchString(tag.Name) {
					continue
				}

				timeToCompare := tag.TagDate
				if timeToCompare == nil {
					timeToCompare = tag.LastCommitDate
				}
				if timeToCompare.Before(time.Now().Add(-staleThreshold)) {
					filteredTags = append(filteredTags, tag)
				}
			}
			view[repo] = filteredTags
		}
		out = view
		return nil
	},
	PostRunE: func(cmd *cobra.Command, args []string) error {
		if out == nil {
			return fmt.Errorf("no results found")
		}
		view := out.(map[string][]*api.GitHubRef)
		if remove {
			for repo, tags := range view {
				if !force {
					if !helpers.Prompt(fmt.Sprintf("Delete [%d] tags in repo [%v]?", len(tags), repo)) {
						continue
					}
				}

				for _, tag := range tags {
					if err := ghApi.DeleteRefs(cmd.Context(), tag.Id); err != nil {
						return fmt.Errorf("unable to delete tags: %v. error: %v", tag.Name, err)
					}
				}
			}
		}
		return nil
	},
}

func init() {
	staleTagsCmd.PersistentFlags().StringVar(&excludePattern, "exclude", "", "If provided, it will be used to exclude tags that match the pattern (regexp)")
}
