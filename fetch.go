package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/logrusorgru/aurora/v4"
)

type Commits struct {
	Nodes []struct {
		Commit struct {
			StatusCheckRollup struct {
				State string
			}
		}
	}
}

type PullRequest struct {
	Number    int
	Title     string
	Url       string
	UpdatedAt time.Time
	IsDraft   bool
	Author    struct {
		Login string
	}
	Repository struct {
		Name string
	}
	Commits Commits `graphql:"commits(last: 1)"`
}

func (pr *PullRequest) toPullRequestItem() PullRequestItem {
	checkStatus := "PENDING"
	if len(pr.Commits.Nodes) > 0 {
		checkStatus = pr.Commits.Nodes[0].Commit.StatusCheckRollup.State
	}

	return PullRequestItem{
		Number:         pr.Number,
		Title:          pr.Title,
		Author:         pr.Author.Login,
		UpdatedAt:      pr.UpdatedAt,
		IsDraft:        pr.IsDraft,
		Url:            pr.Url,
		RepositoryName: pr.Repository.Name,
		CheckStatus:    checkStatus,
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

type RepositoryItem struct {
	Name             string
	PullRequestItems []PullRequestItem
}

type PullRequestItem struct {
	Number         int
	Title          string
	Author         string
	UpdatedAt      time.Time
	IsDraft        bool
	Url            string
	RepositoryName string
	CheckStatus    string
}

func (pri *PullRequestItem) numberWithLink() aurora.Value {
	return aurora.Magenta(fmt.Sprintf("#%d", pri.Number)).Bold().Hyperlink(pri.Url)
}

func (pri *PullRequestItem) checkStatusSymbol() string {
	if pri.CheckStatus == "SUCCESS" {
		return aurora.Green("✔").String()
	} else if pri.CheckStatus == "FAILURE" {
		return aurora.Red("✘").String()
	} else if pri.CheckStatus == "PENDING" {
		return "⏳"
	} else {
		return ""
	}
}

func (pri *PullRequestItem) printLine(numberWidth int, authorWidth, updatedAtWidth int) {
	number := pri.numberWithLink()
	if pri.IsDraft {
		number = aurora.Gray(8, number)
	}
	numberPadding := numberWidth - len(fmt.Sprintf("#%d", pri.Number))

	login := aurora.Green(pri.Author)
	if pri.IsDraft {
		login = aurora.Gray(8, pri.Author)
	}

	updatedAt := pri.UpdatedAt.In(time.Local).Format("2006-01-02")
	if pri.IsDraft {
		updatedAt = aurora.Gray(8, updatedAt).String()
	}

	title := pri.Title
	if pri.IsDraft {
		title = aurora.Gray(8, title+" (draft)").String()
	}

	fmt.Printf("%s%-*s%-*s%-*s%s %s\n", number, numberPadding+1, "", authorWidth+1, login, updatedAtWidth+1, updatedAt, title, pri.checkStatusSymbol())
}

func (ri *RepositoryItem) printList() {
	numberWidth := 0
	authorWidth := 0
	updatedAtWidth := len("2006-01-02")
	for _, pr := range ri.PullRequestItems {
		nWidth := len(fmt.Sprintf("#%d", pr.Number))
		if nWidth > numberWidth {
			numberWidth = nWidth
		}

		aWidth := len(pr.Author)
		if aWidth > authorWidth {
			authorWidth = aWidth
		}
	}

	fmt.Print(aurora.Gray(0, fmt.Sprintf("# %s\n", ri.Name)).BgGray(18))
	prs := ri.PullRequestItems
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].Number > prs[j].Number
	})

	for _, pr := range prs {
		pr.printLine(numberWidth, authorWidth, updatedAtWidth)
	}
}

func formatQueryString(org string, opts *Options) string {
	queryString := fmt.Sprintf("is:open is:pr archived:false org:%s", org)
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

func fetchPullRequests(queryString string, limit int) ([]RepositoryItem, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return []RepositoryItem{}, err
	}

	var query = query{}
	variables := map[string]interface{}{
		"first": graphql.Int(limit),
		"query": graphql.String(queryString),
	}
	err = client.Query("PullRequests", &query, variables)
	if err != nil {
		return []RepositoryItem{}, err
	}

	repoMap := map[string]RepositoryItem{}
	for _, node := range query.Search.Nodes {
		pr := node.PullRequest
		if _, ok := repoMap[pr.Repository.Name]; !ok {
			repoMap[pr.Repository.Name] = RepositoryItem{Name: pr.Repository.Name, PullRequestItems: []PullRequestItem{}}
		}
		pullRequestItems := append(repoMap[pr.Repository.Name].PullRequestItems, pr.toPullRequestItem())
		repositoryItem := RepositoryItem{Name: pr.Repository.Name, PullRequestItems: pullRequestItems}
		repoMap[pr.Repository.Name] = repositoryItem
	}

	// sort pull requests in each repositories
	for _, name := range repoMap {
		prs := name.PullRequestItems
		sort.Slice(prs, func(i, j int) bool {
			return prs[i].Number < prs[j].Number
		})
	}

	// get sorted repositories
	repoNames := make([]string, 0, len(repoMap))
	for name := range repoMap {
		repoNames = append(repoNames, name)
	}
	sort.Strings(repoNames)

	repositories := make([]RepositoryItem, 0, len(repoMap))
	for _, name := range repoNames {
		repositories = append(repositories, repoMap[name])
	}

	return repositories, nil
}
