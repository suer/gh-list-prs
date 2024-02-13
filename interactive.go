package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/browser"
)

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

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd { return nil }

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

func (m model) View() string { return m.list.View() }

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
