package graphqlquery

import (
	"testing"
)

type graphqlVariables struct {
	Owner        string
	Repo         string
	Number       int
	CommitsAfter string
}

type graphqlResult struct {
	GraphQLArguments struct {
		WithCommits bool `graphql:"$withCommits,notnull"`
	}
	Repository struct {
		GraphQLArguments struct {
			Owner string `graphql:"$owner,notnull"`
			Name  string `graphql:"$repo,notnull"`
		}
		DefaultBranchRef struct {
			Name string
		}
		PullRequest struct {
			GraphQLArguments struct {
				Number int `graphql:"$number,notnull"`
			}
			Title   string
			Number  int
			BaseRef struct {
				Name string
			}
			Commits struct {
				GraphQLArguments struct {
					First int    `graphql:"100"`
					After string `graphql:"$commitsAfter"`
				}
				Edges []struct {
					Node struct {
						Commit struct {
							Message string
						}
					}
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
				TotalCount int
			} `graphql:"@include(if: $withCommits)"`
		}
	}
	RateLimit struct {
		Remaining int
	}
}

func TestToString(t *testing.T) {
	var r graphqlResult
	s, err := New(&r).String()
	t.Log(s, err)
}
