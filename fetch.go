package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
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
		NameWithOwner string
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
		RepositoryName: pr.Repository.NameWithOwner,
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

func formatQueryString(org string, opts *Options) string {
	queryString := fmt.Sprintf("is:open is:pr archived:false org:%s", org)
	for _, exclude := range *opts.Excludes {
		if strings.Contains(exclude, "/") {
			queryString += fmt.Sprintf(" -repo:%s", exclude)
		} else {
			queryString += fmt.Sprintf(" -repo:%s/%s", org, exclude)
		}
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
		if _, ok := repoMap[pr.Repository.NameWithOwner]; !ok {
			repoMap[pr.Repository.NameWithOwner] = RepositoryItem{Name: pr.Repository.NameWithOwner, PullRequestItems: []PullRequestItem{}}
		}
		pullRequestItems := append(repoMap[pr.Repository.NameWithOwner].PullRequestItems, pr.toPullRequestItem())
		repositoryItem := RepositoryItem{Name: pr.Repository.NameWithOwner, PullRequestItems: pullRequestItems}
		repoMap[pr.Repository.NameWithOwner] = repositoryItem
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
