package main

import (
	"testing"
	"time"
)

func TestFormatQueryString(t *testing.T) {
	tests := []struct {
		name           string
		org            string
		opts           *Options
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "basic query",
			org:  "myorg",
			opts: &Options{
				Excludes:           &[]string{},
				AdditionalQueries:  &[]string{},
				Author:             "",
				Verbose:            false,
			},
			wantContains: []string{
				"is:open",
				"is:pr",
				"archived:false",
				"org:myorg",
			},
			wantNotContain: []string{},
		},
		{
			name: "with excludes",
			org:  "myorg",
			opts: &Options{
				Excludes:          &[]string{"repo1", "org2/repo2"},
				AdditionalQueries: &[]string{},
				Author:            "",
				Verbose:           false,
			},
			wantContains: []string{
				"org:myorg",
				"-repo:myorg/repo1",
				"-repo:org2/repo2",
			},
			wantNotContain: []string{},
		},
		{
			name: "with author",
			org:  "myorg",
			opts: &Options{
				Excludes:          &[]string{},
				AdditionalQueries: &[]string{},
				Author:            "alice",
				Verbose:           false,
			},
			wantContains: []string{
				"org:myorg",
				"author:alice",
			},
			wantNotContain: []string{},
		},
		{
			name: "with additional queries",
			org:  "myorg",
			opts: &Options{
				Excludes:          &[]string{},
				AdditionalQueries: &[]string{"label:bug", "state:draft"},
				Author:            "",
				Verbose:           false,
			},
			wantContains: []string{
				"org:myorg",
				"label:bug",
				"state:draft",
			},
			wantNotContain: []string{},
		},
		{
			name: "with all options",
			org:  "myorg",
			opts: &Options{
				Excludes:          &[]string{"repo1", "org2/repo2"},
				AdditionalQueries: &[]string{"label:bug"},
				Author:            "bob",
				Verbose:           false,
			},
			wantContains: []string{
				"is:open",
				"is:pr",
				"archived:false",
				"org:myorg",
				"-repo:myorg/repo1",
				"-repo:org2/repo2",
				"author:bob",
				"label:bug",
			},
			wantNotContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatQueryString(tt.org, tt.opts)

			for _, want := range tt.wantContains {
				if !contains(result, want) {
					t.Errorf("formatQueryString() = %q, want to contain %q", result, want)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if contains(result, notWant) {
					t.Errorf("formatQueryString() = %q, should not contain %q", result, notWant)
				}
			}
		})
	}
}

func TestToPullRequestItem(t *testing.T) {
	tests := []struct {
		name string
		pr   *PullRequest
		want PullRequestItem
	}{
		{
			name: "PR with status check",
			pr: &PullRequest{
				Number: 123,
				Title:  "Fix bug",
				Url:    "https://github.com/test/repo/pull/123",
				UpdatedAt: time.Date(2024, 11, 30, 12, 0, 0, 0, time.UTC),
				IsDraft: false,
				Author: struct{ Login string }{Login: "alice"},
				Repository: struct{ NameWithOwner string }{NameWithOwner: "test/repo"},
				Commits: Commits{
					Nodes: []struct {
						Commit struct {
							StatusCheckRollup struct {
								State string
							}
						}
					}{
						{
							Commit: struct {
								StatusCheckRollup struct {
									State string
								}
							}{
								StatusCheckRollup: struct {
									State string
								}{
									State: "SUCCESS",
								},
							},
						},
					},
				},
			},
			want: PullRequestItem{
				Number:         123,
				Title:          "Fix bug",
				Author:         "alice",
				UpdatedAt:      time.Date(2024, 11, 30, 12, 0, 0, 0, time.UTC),
				IsDraft:        false,
				Url:            "https://github.com/test/repo/pull/123",
				RepositoryName: "test/repo",
				CheckStatus:    "SUCCESS",
			},
		},
		{
			name: "PR without status check",
			pr: &PullRequest{
				Number: 456,
				Title:  "Add feature",
				Url:    "https://github.com/test/repo/pull/456",
				UpdatedAt: time.Date(2024, 11, 30, 13, 0, 0, 0, time.UTC),
				IsDraft: true,
				Author: struct{ Login string }{Login: "bob"},
				Repository: struct{ NameWithOwner string }{NameWithOwner: "test/repo"},
				Commits: Commits{
					Nodes: []struct {
						Commit struct {
							StatusCheckRollup struct {
								State string
							}
						}
					}{},
				},
			},
			want: PullRequestItem{
				Number:         456,
				Title:          "Add feature",
				Author:         "bob",
				UpdatedAt:      time.Date(2024, 11, 30, 13, 0, 0, 0, time.UTC),
				IsDraft:        true,
				Url:            "https://github.com/test/repo/pull/456",
				RepositoryName: "test/repo",
				CheckStatus:    "PENDING",
			},
		},
		{
			name: "PR with failure check status",
			pr: &PullRequest{
				Number: 789,
				Title:  "WIP",
				Url:    "https://github.com/test/repo/pull/789",
				UpdatedAt: time.Date(2024, 11, 30, 14, 0, 0, 0, time.UTC),
				IsDraft: false,
				Author: struct{ Login string }{Login: "charlie"},
				Repository: struct{ NameWithOwner string }{NameWithOwner: "test/repo"},
				Commits: Commits{
					Nodes: []struct {
						Commit struct {
							StatusCheckRollup struct {
								State string
							}
						}
					}{
						{
							Commit: struct {
								StatusCheckRollup struct {
									State string
								}
							}{
								StatusCheckRollup: struct {
									State string
								}{
									State: "FAILURE",
								},
							},
						},
					},
				},
			},
			want: PullRequestItem{
				Number:         789,
				Title:          "WIP",
				Author:         "charlie",
				UpdatedAt:      time.Date(2024, 11, 30, 14, 0, 0, 0, time.UTC),
				IsDraft:        false,
				Url:            "https://github.com/test/repo/pull/789",
				RepositoryName: "test/repo",
				CheckStatus:    "FAILURE",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pr.toPullRequestItem()

			if result.Number != tt.want.Number {
				t.Errorf("Number: got %d, want %d", result.Number, tt.want.Number)
			}
			if result.Title != tt.want.Title {
				t.Errorf("Title: got %q, want %q", result.Title, tt.want.Title)
			}
			if result.Author != tt.want.Author {
				t.Errorf("Author: got %q, want %q", result.Author, tt.want.Author)
			}
			if result.IsDraft != tt.want.IsDraft {
				t.Errorf("IsDraft: got %v, want %v", result.IsDraft, tt.want.IsDraft)
			}
			if result.Url != tt.want.Url {
				t.Errorf("Url: got %q, want %q", result.Url, tt.want.Url)
			}
			if result.RepositoryName != tt.want.RepositoryName {
				t.Errorf("RepositoryName: got %q, want %q", result.RepositoryName, tt.want.RepositoryName)
			}
			if result.CheckStatus != tt.want.CheckStatus {
				t.Errorf("CheckStatus: got %q, want %q", result.CheckStatus, tt.want.CheckStatus)
			}
			if result.UpdatedAt != tt.want.UpdatedAt {
				t.Errorf("UpdatedAt: got %v, want %v", result.UpdatedAt, tt.want.UpdatedAt)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
