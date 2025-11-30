package main

import (
	"strings"
	"testing"
	"time"
)

func TestColorFormatterFormatPRNumber(t *testing.T) {
	cf := &ColorFormatter{}
	tests := []struct {
		name         string
		pri          *PullRequestItem
		wantContains string
		wantURL      string
	}{
		{
			name: "regular PR",
			pri: &PullRequestItem{
				Number:  123,
				Url:     "https://github.com/test/repo/pull/123",
				IsDraft: false,
			},
			wantContains: "#123",
			wantURL:      "https://github.com/test/repo/pull/123",
		},
		{
			name: "draft PR",
			pri: &PullRequestItem{
				Number:  456,
				Url:     "https://github.com/test/repo/pull/456",
				IsDraft: true,
			},
			wantContains: "#456",
			wantURL:      "https://github.com/test/repo/pull/456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cf.FormatPRNumber(tt.pri)
			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("FormatPRNumber() = %q, want to contain %q", result, tt.wantContains)
			}
			if !strings.Contains(result, tt.wantURL) {
				t.Errorf("FormatPRNumber() = %q, want to contain URL %q", result, tt.wantURL)
			}
		})
	}
}

func TestColorFormatterFormatAuthor(t *testing.T) {
	cf := &ColorFormatter{}
	tests := []struct {
		name       string
		pri        *PullRequestItem
		wantAuthor string
	}{
		{
			name: "regular PR author",
			pri: &PullRequestItem{
				Author:  "alice",
				IsDraft: false,
			},
			wantAuthor: "alice",
		},
		{
			name: "draft PR author",
			pri: &PullRequestItem{
				Author:  "bob",
				IsDraft: true,
			},
			wantAuthor: "bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cf.FormatAuthor(tt.pri)
			if !strings.Contains(result, tt.wantAuthor) {
				t.Errorf("FormatAuthor() = %q, want to contain %q", result, tt.wantAuthor)
			}
		})
	}
}

func TestColorFormatterFormatUpdatedAt(t *testing.T) {
	cf := &ColorFormatter{}
	testTime := time.Date(2024, 11, 30, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		pri      *PullRequestItem
		wantDate string
	}{
		{
			name: "regular PR updated at",
			pri: &PullRequestItem{
				UpdatedAt: testTime,
				IsDraft:   false,
			},
			wantDate: "2024-11-30",
		},
		{
			name: "draft PR updated at",
			pri: &PullRequestItem{
				UpdatedAt: testTime,
				IsDraft:   true,
			},
			wantDate: "2024-11-30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cf.FormatUpdatedAt(tt.pri)
			if !strings.Contains(result, tt.wantDate) {
				t.Errorf("FormatUpdatedAt() = %q, want to contain %q", result, tt.wantDate)
			}
		})
	}
}

func TestColorFormatterFormatTitle(t *testing.T) {
	cf := &ColorFormatter{}
	tests := []struct {
		name      string
		pri       *PullRequestItem
		wantTitle string
	}{
		{
			name: "regular PR title",
			pri: &PullRequestItem{
				Title:   "Fix bug",
				IsDraft: false,
			},
			wantTitle: "Fix bug",
		},
		{
			name: "draft PR title",
			pri: &PullRequestItem{
				Title:   "WIP feature",
				IsDraft: true,
			},
			wantTitle: "WIP feature (draft)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cf.FormatTitle(tt.pri)
			if !strings.Contains(result, tt.wantTitle) {
				t.Errorf("FormatTitle() = %q, want to contain %q", result, tt.wantTitle)
			}
		})
	}
}

func TestColorFormatterFormatCheckStatus(t *testing.T) {
	cf := &ColorFormatter{}
	tests := []struct {
		name        string
		checkStatus string
		wantSymbol  string
	}{
		{
			name:        "success status",
			checkStatus: "SUCCESS",
			wantSymbol:  "✔",
		},
		{
			name:        "failure status",
			checkStatus: "FAILURE",
			wantSymbol:  "✘",
		},
		{
			name:        "pending status",
			checkStatus: "PENDING",
			wantSymbol:  "⏳",
		},
		{
			name:        "unknown status",
			checkStatus: "UNKNOWN",
			wantSymbol:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pri := &PullRequestItem{CheckStatus: tt.checkStatus}
			result := cf.FormatCheckStatus(pri)
			if tt.wantSymbol == "" {
				if result != "" {
					t.Errorf("FormatCheckStatus() = %q, want empty string", result)
				}
			} else if !strings.Contains(result, tt.wantSymbol) {
				t.Errorf("FormatCheckStatus() = %q, want to contain %q", result, tt.wantSymbol)
			}
		})
	}
}

func TestColorFormatterFormatRepositoryName(t *testing.T) {
	cf := &ColorFormatter{}
	tests := []struct {
		name         string
		repoName     string
		wantRepoName string
		wantLink     string
	}{
		{
			name:         "format repository name",
			repoName:     "test/repo",
			wantRepoName: "test/repo",
			wantLink:     "https://github.com/test/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cf.FormatRepositoryName(tt.repoName)
			if !strings.Contains(result, tt.wantRepoName) {
				t.Errorf("FormatRepositoryName() = %q, want to contain %q", result, tt.wantRepoName)
			}
			if !strings.Contains(result, tt.wantLink) {
				t.Errorf("FormatRepositoryName() = %q, want to contain %q", result, tt.wantLink)
			}
		})
	}
}

func TestNoColorFormatterFormatPRNumber(t *testing.T) {
	ncf := &NoColorFormatter{}
	pri := &PullRequestItem{Number: 123}
	expected := "#123"
	result := ncf.FormatPRNumber(pri)

	if result != expected {
		t.Errorf("FormatPRNumber() = %q, want %q", result, expected)
	}
}

func TestNoColorFormatterFormatAuthor(t *testing.T) {
	ncf := &NoColorFormatter{}
	pri := &PullRequestItem{Author: "alice"}
	expected := "alice"
	result := ncf.FormatAuthor(pri)

	if result != expected {
		t.Errorf("FormatAuthor() = %q, want %q", result, expected)
	}
}

func TestNoColorFormatterFormatUpdatedAt(t *testing.T) {
	ncf := &NoColorFormatter{}
	testTime := time.Date(2024, 11, 30, 12, 0, 0, 0, time.UTC)
	pri := &PullRequestItem{UpdatedAt: testTime}
	result := ncf.FormatUpdatedAt(pri)

	expectedDate := testTime.In(time.Local).Format("2006-01-02")
	if result != expectedDate {
		t.Errorf("FormatUpdatedAt() = %q, want %q", result, expectedDate)
	}
}

func TestNoColorFormatterFormatTitle(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		isDraft  bool
		expected string
	}{
		{
			name:     "regular title",
			title:    "Fix bug",
			isDraft:  false,
			expected: "Fix bug",
		},
		{
			name:     "draft title",
			title:    "WIP feature",
			isDraft:  true,
			expected: "WIP feature (draft)",
		},
	}

	ncf := &NoColorFormatter{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pri := &PullRequestItem{Title: tt.title, IsDraft: tt.isDraft}
			result := ncf.FormatTitle(pri)

			if result != tt.expected {
				t.Errorf("FormatTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNoColorFormatterFormatCheckStatus(t *testing.T) {
	ncf := &NoColorFormatter{}
	tests := []struct {
		name        string
		checkStatus string
		want        string
	}{
		{
			name:        "success status",
			checkStatus: "SUCCESS",
			want:        "✔",
		},
		{
			name:        "failure status",
			checkStatus: "FAILURE",
			want:        "✘",
		},
		{
			name:        "pending status",
			checkStatus: "PENDING",
			want:        "⏳",
		},
		{
			name:        "unknown status",
			checkStatus: "UNKNOWN",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pri := &PullRequestItem{CheckStatus: tt.checkStatus}
			result := ncf.FormatCheckStatus(pri)

			if result != tt.want {
				t.Errorf("FormatCheckStatus() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestNoColorFormatterFormatRepositoryName(t *testing.T) {
	ncf := &NoColorFormatter{}
	repoName := "test/repo"
	result := ncf.FormatRepositoryName(repoName)

	if result != repoName {
		t.Errorf("FormatRepositoryName() = %q, want %q", result, repoName)
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name    string
		noColor bool
		want    Formatter
	}{
		{
			name:    "with color",
			noColor: false,
			want:    &ColorFormatter{},
		},
		{
			name:    "without color",
			noColor: true,
			want:    &NoColorFormatter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFormatter(tt.noColor)
			// Check type instead of equality since we can't compare pointers directly
			switch v := result.(type) {
			case *ColorFormatter:
				if tt.noColor {
					t.Errorf("NewFormatter(true) returned ColorFormatter, want NoColorFormatter")
				}
			case *NoColorFormatter:
				if !tt.noColor {
					t.Errorf("NewFormatter(false) returned NoColorFormatter, want ColorFormatter")
				}
			default:
				t.Errorf("NewFormatter() returned unknown type %T", v)
			}
		})
	}
}
