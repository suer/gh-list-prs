package main

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/spf13/cobra"
)

type Options struct {
	Limit             int
	Excludes          []string
	Author            string
	AdditionalQueries []string
	Verbose           bool
	Interactive       bool
	NoColor           bool
}

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "unknown"
}

func rootCmd() *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:           "gh list-prs <org> [<org>...]",
		Short:         "List PRs for one or more orgs",
		Version:       buildVersion(),
		Args:          cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			orgs := args

			if opts.Limit <= 0 {
				return errors.New("invalid limit")
			}

			return run(orgs, opts)
		},
	}
	cmd.SetVersionTemplate("{{.Version}}\n")

	cmd.Flags().StringArrayVarP(&opts.Excludes, "exclude", "e", []string{}, "exclude repositories")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 50, "Max number of search results in all repository")
	cmd.Flags().StringVarP(&opts.Author, "author", "a", "", "Filter by author")
	cmd.Flags().StringArrayVarP(&opts.AdditionalQueries, "additional-query", "q", []string{}, "additional query")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "verbose output")
	cmd.Flags().BoolVarP(&opts.Interactive, "interactive", "i", false, "interactive mode")
	cmd.Flags().BoolVar(&opts.NoColor, "no-color", false, "disable color output and show plain URLs")
	return cmd
}

func run(orgs []string, opts *Options) error {
	type result struct {
		repositories []RepositoryItem
		err          error
	}

	results := make([]result, len(orgs))
	var wg sync.WaitGroup

	for i, org := range orgs {
		wg.Add(1)
		go func(i int, org string) {
			defer wg.Done()
			queryString := formatQueryString(org, opts)
			if opts.Verbose {
				fmt.Printf("query: %s\n", queryString)
			}
			repositories, err := fetchPullRequests(queryString, opts.Limit)
			results[i] = result{repositories: repositories, err: err}
		}(i, org)
	}

	wg.Wait()

	var allRepositories []RepositoryItem
	for _, r := range results {
		if r.err != nil {
			return r.err
		}
		allRepositories = append(allRepositories, r.repositories...)
	}

	if opts.Interactive {
		if err := printResultInteractive(orgs, allRepositories, opts.NoColor); err != nil {
			return err
		}
	} else {
		printResult(allRepositories, opts)
	}

	return nil
}

func printResult(repositories []RepositoryItem, opts *Options) {
	for _, repo := range repositories {
		repo.printList(opts)
		fmt.Println()
	}
}
