package main

import (
	"errors"
	"fmt"
	"sync"

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
		Use:           "gh list-prs <org> [<org>...]",
		Short:         "List PRs for one or more orgs",
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

	cmd.Flags().StringArrayVarP(opts.Excludes, "exclude", "e", []string{}, "exclude repositories")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 50, "Max number of search results in all repository")
	cmd.Flags().StringVarP(&opts.Author, "author", "a", "", "Filter by author")
	cmd.Flags().StringArrayVarP(opts.AdditionalQueries, "additional-query", "q", []string{}, "additional query")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "verbose output")
	cmd.Flags().BoolVarP(&opts.Interactive, "interactive", "i", false, "interactive mode")
	cmd.Flags().BoolVar(&opts.NoColor, "no-color", false, "disable color output and show plain URLs")
	return cmd
}

func run(orgs []string, opts *Options) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allRepositories []RepositoryItem
	var firstError error

	for _, org := range orgs {
		wg.Add(1)
		go func(org string) {
			defer wg.Done()
			
			queryString := formatQueryString(org, opts)
			repositories, err := fetchPullRequests(queryString, opts.Limit)
			
			mu.Lock()
			defer mu.Unlock()
			
			if err != nil && firstError == nil {
				firstError = err
				return
			}
			
			allRepositories = append(allRepositories, repositories...)
		}(org)
	}

	wg.Wait()

	if firstError != nil {
		return firstError
	}

	if opts.Interactive {
		printResultInteractive(orgs, allRepositories)
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
