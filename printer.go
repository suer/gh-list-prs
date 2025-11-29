package main

import (
	"fmt"
	"sort"
)

func (pri *PullRequestItem) printLine(numberWidth int, authorWidth, updatedAtWidth int, formatter Formatter) {
	numberString := formatter.FormatPRNumber(pri)
	numberPadding := numberWidth - len(fmt.Sprintf("#%d", pri.Number))
	login := formatter.FormatAuthor(pri)
	updatedAt := formatter.FormatUpdatedAt(pri)
	title := formatter.FormatTitle(pri)
	statusSymbol := formatter.FormatCheckStatus(pri)

	fmt.Printf("%s%-*s%-*s %-*s %s %s\n", numberString, numberPadding+1, "", authorWidth, login, updatedAtWidth, updatedAt, title, statusSymbol)
}

func (ri *RepositoryItem) printList(opts *Options) {
	formatter := NewFormatter(opts.NoColor)

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

	fmt.Printf("# %s\n", formatter.FormatRepositoryName(ri.Name))
	prs := ri.PullRequestItems
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].Number > prs[j].Number
	})

	for _, pr := range prs {
		pr.printLine(numberWidth, authorWidth, updatedAtWidth, formatter)
	}
}
