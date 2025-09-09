package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/browser"
)

type listItem struct {
	pullRequestItem PullRequestItem
}

func (li listItem) Title() string {
	return fmt.Sprintf("%s #%d @%s %s", li.pullRequestItem.RepositoryName, li.pullRequestItem.Number, li.pullRequestItem.Author, li.pullRequestItem.checkStatusSymbol(false))
}
func (li listItem) Description() string { return li.pullRequestItem.Title }
func (li listItem) FilterValue() string {
	return li.pullRequestItem.RepositoryName + li.pullRequestItem.Author + li.pullRequestItem.Title
}

type model struct {
	list list.Model
	keys *listKeyMap
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if key.Matches(msg, m.keys.openWithBrowser) {
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

func (m model) View() string { return m.list.View() }

type listKeyMap struct {
	openWithBrowser key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		openWithBrowser: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open with browser"),
		),
	}
}

func printResultInteractive(orgs []string, repositories []RepositoryItem) error {
	items := []list.Item{}
	for _, repo := range repositories {
		for _, pr := range repo.PullRequestItems {
			items = append(items, listItem{pullRequestItem: pr})
		}
	}

	listKeys := newListKeyMap()
	groceryList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	groceryList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.openWithBrowser,
		}
	}
	groceryList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.openWithBrowser,
		}
	}
	m := model{list: groceryList, keys: listKeys}
	if len(orgs) == 1 {
		m.list.Title = fmt.Sprintf("PRs in %s", orgs[0])
	} else {
		m.list.Title = fmt.Sprintf("PRs in %d orgs", len(orgs))
	}
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
