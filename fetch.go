package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

const (
	checkStatusSuccess = "SUCCESS"
	checkStatusFailure = "FAILURE"
	checkStatusPending = "PENDING"
	checkStatusUnknown = "UNKNOWN"
)

const reviewDecisionApproved = "APPROVED"

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
	ReviewDecision string
	Repository     struct {
		NameWithOwner string
	}
	Commits Commits `graphql:"commits(last: 1)"`
}

func (pr *PullRequest) toPullRequestItem() PullRequestItem {
	checkStatus := checkStatusUnknown
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
		ReviewDecision: pr.ReviewDecision,
	}
}

type Search struct {
	Nodes []struct {
		PullRequest `graphql:"... on PullRequest"`
	}
}

type searchQuery struct {
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
	ReviewDecision string
}

func formatQueryString(org string, opts *Options) string {
	queryString := fmt.Sprintf("is:open is:pr archived:false org:%s", org)
	for _, exclude := range opts.Excludes {
		if strings.Contains(exclude, "/") {
			queryString += fmt.Sprintf(" -repo:%s", exclude)
		} else {
			queryString += fmt.Sprintf(" -repo:%s/%s", org, exclude)
		}
	}
	if opts.Author != "" {
		queryString += fmt.Sprintf(" author:%s", opts.Author)
	}
	for _, query := range opts.AdditionalQueries {
		queryString += fmt.Sprintf(" %s", query)
	}

	return queryString
}

func fetchPullRequests(queryString string, limit int) ([]RepositoryItem, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return []RepositoryItem{}, err
	}

	var query = searchQuery{}
	variables := map[string]interface{}{
		"first": graphql.Int(limit),
		"query": graphql.String(queryString),
	}
	err = client.Query("PullRequests", &query, variables)
	if err != nil {
		return []RepositoryItem{}, err
	}

	pullRequests := make([]PullRequest, 0, len(query.Nodes))
	for _, node := range query.Nodes {
		pullRequests = append(pullRequests, node.PullRequest)
	}

	return groupAndSortPullRequests(pullRequests), nil
}

// groupAndSortPullRequests groups pull requests by repository, sorts pull
// requests within each repository by Number descending, and sorts
// repositories by name.
func groupAndSortPullRequests(pullRequests []PullRequest) []RepositoryItem {
	repoMap := map[string][]PullRequestItem{}
	for _, pr := range pullRequests {
		name := pr.Repository.NameWithOwner
		repoMap[name] = append(repoMap[name], pr.toPullRequestItem())
	}

	repoNames := make([]string, 0, len(repoMap))
	for name := range repoMap {
		repoNames = append(repoNames, name)
	}
	sort.Strings(repoNames)

	repositories := make([]RepositoryItem, 0, len(repoMap))
	for _, name := range repoNames {
		items := repoMap[name]
		sort.Slice(items, func(i, j int) bool {
			return items[i].Number > items[j].Number
		})
		repositories = append(repositories, RepositoryItem{Name: name, PullRequestItems: items})
	}

	return repositories
}
