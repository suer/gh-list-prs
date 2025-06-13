package main

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

type Options struct {
	Limit             int
	Excludes          *[]string
	Author            string
	AdditionalQueries *[]string
	Verbose           bool
	Interactive       bool
	NoColor           bool
}

func rootCmd() *cobra.Command {
	opts := &Options{Excludes: &[]string{}, AdditionalQueries: &[]string{}}
	cmd := &cobra.Command{
		Use:           "gh list-prs <org>",
		Short:         "List PRs for an org",
		Args:          cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			org := args[0]

			if opts.Limit <= 0 {
				return errors.New("invalid limit")
			}

			return run(org, opts)
		},
	}

	cmd.Flags().StringArrayVarP(opts.Excludes, "exclude", "e", []string{}, "exclude repositories")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 50, "Max number of search results in all repository")
	cmd.Flags().StringVarP(&opts.Author, "author", "a", "", "Filter by author")
	cmd.Flags().StringArrayVarP(opts.AdditionalQueries, "additional-query", "q", []string{}, "additional query")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "verbose output")
	cmd.Flags().BoolVarP(&opts.Interactive, "interactive", "i", false, "interactive mode")
	cmd.Flags().BoolVar(&opts.NoColor, "no-color", false, "disable color output and show plain URLs")
	return cmd
}

func run(org string, opts *Options) error {
	queryString := formatQueryString(org, opts)

	repositories, err := fetchPullRequests(queryString, opts.Limit)
	if err != nil {
		return err
	}

	if opts.Interactive {
		printResultInteractive(org, repositories)
	} else {
		printResult(repositories, opts)
	}

	return nil
}

func printResult(repositories []RepositoryItem, opts *Options) {
	for _, repo := range repositories {
		repo.printList(opts)
		fmt.Println()
	}
}
