package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

func main() {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		log.Fatal(err)
	}

	var query struct {
		Search struct {
			Nodes []struct {
				PullRequest struct {
					Number int
					Title  string
					Url    string
					Author struct {
						Login string
					}
					Repository struct {
						Name string
					}
				} `graphql:"... on PullRequest"`
			}
		} `graphql:"search(first: $first, type: ISSUE, query: $query)"`
	}
	org := os.Args[1] // TODO
	variables := map[string]interface{}{
		"first": graphql.Int(30),
		"query": graphql.String(fmt.Sprintf("is:open is:pr org:%s", org)),
	}
	err = client.Query("PullRequests", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	for _, node := range query.Search.Nodes {
		pr := node.PullRequest
		fmt.Printf("%s\t#%d\t%s\t%s\t%s\n", pr.Repository.Name, pr.Number, pr.Author.Login, pr.Title, pr.Url)
	}
}
