package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
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

func main() {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		log.Fatal(err)
	}

	var query = query{}
	org := os.Args[1] // TODO
	variables := map[string]interface{}{
		"first": graphql.Int(30),
		"query": graphql.String(fmt.Sprintf("is:open is:pr org:%s", org)),
	}
	err = client.Query("PullRequests", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	repo_to_prs := make(map[string][]interface{})
	for _, node := range query.Search.Nodes {
		pr := node.PullRequest
		if _, ok := repo_to_prs[pr.Repository.Name]; !ok {
			repo_to_prs[pr.Repository.Name] = []interface{}{}
		}
		repo_to_prs[pr.Repository.Name] = append(repo_to_prs[pr.Repository.Name], pr)
	}

	repos := []string{}
	for repo := range repo_to_prs {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	for _, repo := range repos {
		fmt.Printf("# %s\n", repo)
		prs := repo_to_prs[repo]
		sort.Slice(prs, func(i, j int) bool {
			return prs[i].(PullRequest).Number > prs[j].(PullRequest).Number
		})
		for _, pr := range prs {
			pr := pr.(PullRequest)
			fmt.Printf("#%d\t%s\t%s\t%s\t%s\n", pr.Number, pr.Author.Login, pr.Title, pr.Url, pr.UpdatedAt.In(time.Local).Format("2006-01-02 15:04:05"))
		}
		fmt.Println()
	}
}
