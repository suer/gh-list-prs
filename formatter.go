package main

import (
	"fmt"
	"time"

	"github.com/logrusorgru/aurora/v4"
)

type Formatter interface {
	FormatPRNumber(pri *PullRequestItem) string
	FormatAuthor(pri *PullRequestItem) string
	FormatUpdatedAt(pri *PullRequestItem) string
	FormatTitle(pri *PullRequestItem) string
	FormatCheckStatus(pri *PullRequestItem) string
	FormatRepositoryName(name string) string
}

type ColorFormatter struct{}

func (cf *ColorFormatter) FormatPRNumber(pri *PullRequestItem) string {
	num := fmt.Sprintf("#%d", pri.Number)
	if pri.IsDraft {
		return aurora.Gray(8, aurora.Magenta(num).Bold().Hyperlink(pri.Url)).String()
	}
	return aurora.Magenta(num).Bold().Hyperlink(pri.Url).String()
}

func (cf *ColorFormatter) FormatAuthor(pri *PullRequestItem) string {
	if pri.IsDraft {
		return aurora.Gray(8, pri.Author).String()
	}
	return aurora.Green(pri.Author).String()
}

func (cf *ColorFormatter) FormatUpdatedAt(pri *PullRequestItem) string {
	updatedAt := pri.UpdatedAt.In(time.Local).Format("2006-01-02")
	if pri.IsDraft {
		return aurora.Gray(8, updatedAt).String()
	}
	return updatedAt
}

func (cf *ColorFormatter) FormatTitle(pri *PullRequestItem) string {
	title := pri.Title
	if pri.IsDraft {
		title = title + " (draft)"
		return aurora.Gray(8, title).String()
	}
	return title
}

func (cf *ColorFormatter) FormatCheckStatus(pri *PullRequestItem) string {
	switch pri.CheckStatus {
	case "SUCCESS":
		return aurora.Green("✔").String()
	case "FAILURE":
		return aurora.Red("✘").String()
	case "PENDING":
		return "⏳"
	default:
		return ""
	}
}

func (cf *ColorFormatter) FormatRepositoryName(name string) string {
	repoLink := fmt.Sprintf("https://github.com/%s", name)
	return aurora.Hyperlink(name, repoLink).String()
}

type NoColorFormatter struct{}

func (ncf *NoColorFormatter) FormatPRNumber(pri *PullRequestItem) string {
	return fmt.Sprintf("#%d", pri.Number)
}

func (ncf *NoColorFormatter) FormatAuthor(pri *PullRequestItem) string {
	return pri.Author
}

func (ncf *NoColorFormatter) FormatUpdatedAt(pri *PullRequestItem) string {
	return pri.UpdatedAt.In(time.Local).Format("2006-01-02")
}

func (ncf *NoColorFormatter) FormatTitle(pri *PullRequestItem) string {
	title := pri.Title
	if pri.IsDraft {
		title = title + " (draft)"
	}
	return title
}

func (ncf *NoColorFormatter) FormatCheckStatus(pri *PullRequestItem) string {
	switch pri.CheckStatus {
	case "SUCCESS":
		return "✔"
	case "FAILURE":
		return "✘"
	case "PENDING":
		return "⏳"
	default:
		return ""
	}
}

func (ncf *NoColorFormatter) FormatRepositoryName(name string) string {
	return name
}

func NewFormatter(noColor bool) Formatter {
	if noColor {
		return &NoColorFormatter{}
	}
	return &ColorFormatter{}
}
