package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/logrusorgru/aurora/v4"
	"github.com/pkg/browser"
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
	Interactive       bool
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
	Url            string
	RepositoryName string
}

func (pr *PullRequest) toPullRequestItem() PullRequestItem {
	return PullRequestItem{
		Number:         pr.Number,
		Title:          pr.Title,
		Author:         pr.Author.Login,
		UpdatedAt:      pr.UpdatedAt,
		Url:            pr.Url,
		RepositoryName: pr.Repository.Name,
	}
}

type listItem struct {
	pullRequestItem PullRequestItem
}

func (li listItem) Title() string {
	return fmt.Sprintf("%s #%d", li.pullRequestItem.RepositoryName, li.pullRequestItem.Number)
}
func (li listItem) Description() string { return li.pullRequestItem.Title }
func (li listItem) FilterValue() string {
	return li.pullRequestItem.RepositoryName + li.pullRequestItem.Author + li.pullRequestItem.Title
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
	cmd.Flags().BoolVarP(&opts.Interactive, "interactive", "i", false, "interactive mode")
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
		printResult(repositories)
	}

	return nil
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

func (ri *RepositoryItem) print() {
	numberWidth := 0
	authorWidth := 0
	createdAtWidth := len("2006-01-02")
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
		number := aurora.Magenta(fmt.Sprintf("#%d", pr.Number)).Bold().Hyperlink(pr.Url)
		numberPadding := numberWidth - len(fmt.Sprintf("#%d", pr.Number))
		login := aurora.Green(pr.Author)
		fmt.Printf("%s%-*s%-*s%-*s%s\n", number, numberPadding+1, "", authorWidth+1, login, createdAtWidth+1, pr.UpdatedAt.In(time.Local).Format("2006-01-02"), pr.Title)
	}
}

func printResult(repositories []RepositoryItem) {
	for _, repo := range repositories {
		repo.print()
		fmt.Println()
	}
}

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		} else if msg.String() == "o" {
			browser.OpenURL(m.list.SelectedItem().(listItem).pullRequestItem.Url)
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.list.View()
}

func printResultInteractive(org string, repositories []RepositoryItem) error {
	items := []list.Item{}
	for _, repo := range repositories {
		for _, pr := range repo.PullRequestItems {
			items = append(items, listItem{pullRequestItem: pr})
		}
	}

	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = fmt.Sprintf("PRs in %s", org)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

func main() {
	cmd := rootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
