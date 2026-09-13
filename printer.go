package main

import (
	"fmt"
	"strings"
)

type columnWidths struct {
	number    int
	author    int
	updatedAt int
}

func (pri *PullRequestItem) printLine(widths columnWidths, formatter Formatter) {
	numberString := formatter.FormatPRNumber(pri)
	numberPadding := widths.number - len(fmt.Sprintf("#%d", pri.Number))
	login := formatter.FormatAuthor(pri)
	updatedAt := formatter.FormatUpdatedAt(pri)
	title := formatter.FormatTitle(pri)
	symbols := make([]string, 0, 3)
	for _, s := range []string{formatter.FormatMergeable(pri), formatter.FormatCheckStatus(pri), formatter.FormatReviewDecision(pri)} {
		if s != "" {
			symbols = append(symbols, s)
		}
	}

	line := fmt.Sprintf("%s%-*s%-*s %-*s %s", numberString, numberPadding+1, "", widths.author, login, widths.updatedAt, updatedAt, title)
	if len(symbols) > 0 {
		line += " " + strings.Join(symbols, " ")
	}
	fmt.Println(line)
}

func (ri *RepositoryItem) printList(opts *Options) {
	formatter := NewFormatter(opts.NoColor)

	widths := columnWidths{updatedAt: len("2006-01-02")}
	for _, pr := range ri.PullRequestItems {
		nWidth := len(fmt.Sprintf("#%d", pr.Number))
		if nWidth > widths.number {
			widths.number = nWidth
		}

		aWidth := len(pr.Author)
		if aWidth > widths.author {
			widths.author = aWidth
		}
	}

	fmt.Printf("# %s\n", formatter.FormatRepositoryName(ri.Name))

	for _, pr := range ri.PullRequestItems {
		pr.printLine(widths, formatter)
	}
}
