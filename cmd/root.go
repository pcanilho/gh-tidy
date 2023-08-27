package cmd

import (
	"fmt"
	"github.com/pcanilho/gh-tidy/api"
	"github.com/pcanilho/gh-tidy/output"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
	"time"
)

// internal
var (
	ghApi      *api.GitHub
	serializer output.Serializer
	out        any
)

// commands
var (
	staleThreshold time.Duration
	excludePattern string
	excludeRegex   *regexp.Regexp
	refs           []string
	remove         bool
)

var (
	owner  string
	format string
	force  bool
	timed  bool
)

var (
	startTime = time.Now()
)

var rootCmd = &cobra.Command{
	Use: "gh-tidy",
	Example: `$ direnv allow || read -s GITHUB_TOKEN; export GITHUB_TOKEN
$ gh tidy stale branches <owner/repo> -t 72h
$ gh tidy stale prs      <owner/repo> -t 72h -s OPEN -s MERGED
$ gh tidy stale tags     <owner/repo> -t 72h
$ gh tidy delete         <owner/repo> -t 72h --ref <branch_name> --ref <tag_name>`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		// Format
		switch strings.TrimSpace(strings.ToLower(format)) {
		case output.JSON:
			serializer = new(output.JsonSerializer)
		case output.YAML:
			serializer = new(output.YamlSerializer)
		default:
			return fmt.Errorf("the provided format [%v] is not supported", format)
		}

		// Patterns
		if len(excludePattern) > 0 {
			excludeRegex, err = regexp.Compile(excludePattern)
			if err != nil {
				return err
			}
		}

		// Internal :: Session
		ghApi, err = api.NewSession()
		if err != nil {
			return err
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if out == nil || len(strings.TrimSpace(fmt.Sprintf("%v", out))) == 0 {
			return nil
		}
		content, err := serializer.Serialize(out)
		if err != nil {
			return fmt.Errorf("[INTERNAL] unable to serialise output. Error: %v", err)
		}
		fmt.Println(string(content))

		if timed {
			fmt.Printf("\nruntime: %v\n", time.Since(startTime))
		}
		return nil
	},
}

var staleCmd = &cobra.Command{
	Use:     "stale",
	Aliases: []string{"inactive"},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&owner, "owner", "o", "", "The GitHub owner value. (Automatically set if the repository is given in the 'owner/repository' format")
	rootCmd.PersistentFlags().StringVar(&format, "format", "json", "The desired output format. Supported values are: yaml, json")
	rootCmd.PersistentFlags().BoolVar(&remove, "rm", false, "If specified, this flag enable the removal mode of the correlated sub-command")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "If specified, all interactive operations will be disabled")
	rootCmd.PersistentFlags().BoolVar(&timed, "timed", false, "If specified, the total execution time will be printed")

	staleCmd.PersistentFlags().DurationVarP(&staleThreshold, "threshold", "t", time.Hour*24*7*4, "The stale threshold value. [1 month]")

	staleCmd.AddCommand(staleBranchesCmd)
	staleCmd.AddCommand(stalePrsCmd)
	staleCmd.AddCommand(staleTagsCmd)

	rootCmd.AddCommand(staleCmd)
	rootCmd.AddCommand(deleteRefCmd)
}
