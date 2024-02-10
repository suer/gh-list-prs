package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/logrusorgru/aurora/v4"
	"github.com/spf13/cobra"
)

type PullRequest struct {
	Number    int
	Title     string
	Url       string
	UpdatedAt time.Time
	Author    struct {
		Login string
	}
	Repository struct {
		Name string
	}
}

type Search struct {
	Nodes []struct {
		PullRequest `graphql:"... on PullRequest"`
	}
}

type query struct {
	Search `graphql:"search(first: $first, type: ISSUE, query: $query)"`
}

type Options struct {
	Limit    int
	Excludes *[]string
}

func rootCmd() *cobra.Command {
	opts := &Options{}
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

	opts.Excludes = cmd.Flags().StringArrayP("exclude", "e", []string{}, "exclude repositories")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "l", 50, "Max number of search results in all repository")
	return cmd
}

func run(org string, opts *Options) error {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return err
	}

	queryString := fmt.Sprintf("is:open is:pr org:%s", org)
	for _, exclude := range *opts.Excludes {
		queryString += fmt.Sprintf(" -repo:%s/%s", org, exclude)
	}

	var query = query{}
	variables := map[string]interface{}{
		"first": graphql.Int(opts.Limit),
		"query": graphql.String(queryString),
	}
	err = client.Query("PullRequests", &query, variables)
	if err != nil {
		return err
	}

	repo_to_prs := make(map[string][]interface{})
	for _, node := range query.Search.Nodes {
		pr := node.PullRequest
		if _, ok := repo_to_prs[pr.Repository.Name]; !ok {
			repo_to_prs[pr.Repository.Name] = []interface{}{}
		}
		repo_to_prs[pr.Repository.Name] = append(repo_to_prs[pr.Repository.Name], pr)
	}

	print_result(repo_to_prs)

	return nil
}

func print_result(repo_to_prs map[string][]interface{}) {
	repos := []string{}
	for repo := range repo_to_prs {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	for _, repo := range repos {
		fmt.Print(aurora.Gray(0, fmt.Sprintf("# %s\n", repo)).BgGray(18))
		prs := repo_to_prs[repo]
		sort.Slice(prs, func(i, j int) bool {
			return prs[i].(PullRequest).Number > prs[j].(PullRequest).Number
		})
		for _, pr := range prs {
			pr := pr.(PullRequest)
			number := aurora.Magenta(fmt.Sprintf("#%d", pr.Number)).Bold().Hyperlink(pr.Url)
			login := aurora.Green(pr.Author.Login)
			fmt.Printf("%s\t%s\t%s\t%s\n", number, login, pr.Title, pr.UpdatedAt.In(time.Local).Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}
}

func main() {
	cmd := rootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
