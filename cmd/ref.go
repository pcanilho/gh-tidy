package cmd

import (
	"fmt"
	"github.com/pcanilho/gh-tidy/api"
	"github.com/pcanilho/gh-tidy/api/helpers"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var deleteRefCmd = &cobra.Command{
	Use:     "delete",
	Example: `$ gh tidy delete --ref <ref>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("only on <owner>/<repository> can to be provided")
		}

		if refs == nil || len(refs) == 0 {
			return fmt.Errorf("at least one ref must be provided")
		}

		repo := args[0]
		if len(owner) == 0 && strings.Contains(args[0], "/") {
			composite := strings.Split(args[0], "/")
			owner, repo = composite[0], composite[1]
		}

		// Owner
		if len(owner) == 0 {
			return fmt.Errorf("the [owner] flag must be provided")
		}
		brs, err := ghApi.ListRefs(cmd.Context(), owner, repo, api.BranchRefType)
		if err != nil {
			return err
		}

		tgs, err := ghApi.ListRefs(cmd.Context(), owner, repo, api.TagRefType)
		if err != nil {
			return err
		}

		rfs := append(brs, tgs...)
		var toDeleteIds []string
		for _, rf := range rfs {
			for _, ref := range refs {
				if rf.Name == ref {
					toDeleteIds = append(toDeleteIds, rf.Id)
				}
			}
		}

		if len(toDeleteIds) == 0 {
			return nil
		}

		if !force {
			if !helpers.Prompt(fmt.Sprintf("Delete [%d] refs?", len(toDeleteIds))) {
				os.Exit(0)
			}
		}
		if err = ghApi.DeleteRefs(cmd.Context(), toDeleteIds...); err != nil {
			return fmt.Errorf("unable to delete [refs=%v]. error: %v", refs, err)
		}

		out = fmt.Sprintf("Deleted [refs=%v] with [ids=%v]\n", refs, toDeleteIds)
		return nil
	},
}

func init() {
	deleteRefCmd.PersistentFlags().StringArrayVar(&refs, "ref", nil, "The provided ref(s) to be deleted. Only branch or tag name or are supported")
}
