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
	Repository struct {
		GraphQLParams struct {
			Owner string
			Name  string
		}
		DefaultBranchRef struct {
			Name string
		}
		PullRequest struct {
			GraphQLParams struct {
				Number int
			}
			Title   string
			Number  int
			BaseRef struct {
				Name string
			}
			Commits struct {
				GraphQLParams struct {
					First int `graphql:"100"`
					After string
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
			}
		}
	}
	RateLimit struct {
		Remaining int
	}
}

func TestToString(t *testing.T) {
	var r graphqlResult
	vars := graphqlVariables{
		Owner:  "motemen",
		Repo:   "test-repository",
		Number: 2,
	}
	s, err := New(&r, &vars).
		Bind(&r.Repository.GraphQLParams.Owner, Var{&vars.Owner}).
		Bind(&r.Repository.GraphQLParams.Name, Var{&vars.Repo}).
		Bind(&r.Repository.PullRequest.GraphQLParams.Number, Var{&vars.Number}).
		Bind(&r.Repository.PullRequest.Commits.GraphQLParams.First, 100).
		Bind(&r.Repository.PullRequest.Commits.GraphQLParams.After, Var{&vars.CommitsAfter}).
		String()
	t.Log(s, err)
}
