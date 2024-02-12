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
	Limit             int
	Excludes          *[]string
	Author            string
	AdditionalQueries *[]string
	Verbose           bool
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
	cmd.Flags().StringVarP(&opts.Author, "author", "a", "", "Filter by author")
	opts.AdditionalQueries = cmd.Flags().StringArrayP("additional-query", "q", []string{}, "additional query")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "verbose output")
	return cmd
}

func run(org string, opts *Options) error {
	queryString := formatQueryString(org, opts)

	repoToPrs, err := fetchPullRequests(queryString, opts.Limit)
	if err != nil {
		return err
	}

	printResult(repoToPrs)

	return nil
}

func formatQueryString(org string, opts *Options) string {
	queryString := fmt.Sprintf("is:mpen is:pr archived:false org:%s", org)
	for _, exclude := range *opts.Excludes {
		queryString += fmt.Sprintf(" -repo:%s/%s", org, exclude)
	}
	if opts.Author != "" {
		queryString += fmt.Sprintf(" author:%s", opts.Author)
	}
	for _, query := range *opts.AdditionalQueries {
		queryString += fmt.Sprintf(" %s", query)
	}

	if opts.Verbose {
		fmt.Printf("query: %s\n", queryString)
	}

	return queryString
}

func fetchPullRequests(queryString string, limit int) (map[string][]PullRequest, error) {
	repoToPrs := make(map[string][]PullRequest)

	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, err
	}

	var query = query{}
	variables := map[string]interface{}{
		"first": graphql.Int(limit),
		"query": graphql.String(queryString),
	}
	err = client.Query("PullRequests", &query, variables)
	if err != nil {
		return repoToPrs, err
	}

	for _, node := range query.Search.Nodes {
		pr := node.PullRequest
		if _, ok := repoToPrs[pr.Repository.Name]; !ok {
			repoToPrs[pr.Repository.Name] = []PullRequest{}
		}
		repoToPrs[pr.Repository.Name] = append(repoToPrs[pr.Repository.Name], pr)
	}
	return repoToPrs, nil
}

func printResult(repoToPrs map[string][]PullRequest) {
	repos := []string{}
	for repo := range repoToPrs {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	numberWidth := 0
	authorWidth := 0
	createdAtWidth := len("2006-01-02")
	for _, repo := range repos {
		for _, pr := range repoToPrs[repo] {
			nWidth := len(fmt.Sprintf("#%d", pr.Number))
			if nWidth > numberWidth {
				numberWidth = nWidth
			}

			aWidth := len(pr.Author.Login)
			if aWidth > authorWidth {
				authorWidth = aWidth
			}
		}
	}

	for _, repo := range repos {
		fmt.Print(aurora.Gray(0, fmt.Sprintf("# %s\n", repo)).BgGray(18))
		prs := repoToPrs[repo]
		sort.Slice(prs, func(i, j int) bool {
			return prs[i].Number > prs[j].Number
		})
		for _, pr := range prs {
			number := aurora.Magenta(fmt.Sprintf("#%d", pr.Number)).Bold().Hyperlink(pr.Url)
			numberPadding := numberWidth - len(fmt.Sprintf("#%d", pr.Number))
			login := aurora.Green(pr.Author.Login)
			fmt.Printf("%s%-*s%-*s%-*s%s\n", number, numberPadding+1, "", authorWidth+1, login, createdAtWidth+1, pr.UpdatedAt.In(time.Local).Format("2006-01-02"), pr.Title)
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
