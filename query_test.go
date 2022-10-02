package graphqlquery

import (
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
)

type simple struct {
	Hero struct {
		Name string
	}
}

type arguments struct {
	Human []struct {
		GraphQLArguments struct {
			ID string `graphql:"\"1000\""`
		}
		Name   string
		Height int
	}
}

type arguments2 struct {
	Human []struct {
		Name   string
		Height int
	} `graphql:"(id: \"1000\")"`
}

type argumentsScalar struct {
	Human struct {
		GraphQLArguments struct {
			ID string `graphql:"\"1000\""`
		}
		Name   string
		Height int `graphql:"(unit: FOOT)"`
	}
}

type aliases struct {
	EmpireHero struct {
		GraphQLArguments struct {
			Episode string `graphql:"EMPIRE"`
		}
		Name string
	} `graphql:"alias=hero"`
	JediHero struct {
		GraphQLArguments struct {
			Episode string `graphql:"JEDI"`
		}
		Name string
	} `graphql:"alias=hero"`
}

type variables struct {
	Hero struct {
		GraphQLArguments struct {
			Episode Episode `graphql:"$episode"`
		}
		Friends []struct {
			Name string
		}
	}
}

type Episode string

type inlineFragments struct {
	GraphQLArguments struct {
		Episode Episode `graphql:"$ep,notnull"`
	}
	Hero struct {
		Name        string
		Height      int `graphql:"... on Human"`
		DroidFields `graphql:"... on Droid"`
	} `graphql:"(episode: $ep)"`
}

type DroidFields struct {
	PrimaryFunction string
}

type directives struct {
	GraphQLArguments struct {
		WithFriends bool `graphql:"$withFriends"`
	}
	Hero struct {
		Name    string
		Friends []struct {
			Name string
		} `graphql:"@include(if: $withFriends)"`
	}
}

type pointers struct {
	EmpireHero *struct {
		GraphQLArguments struct {
			Episode string `graphql:"EMPIRE"`
		}
		Name string
	} `graphql:"alias=hero"`
	JediHero *struct {
		Name string
	} `graphql:"alias=hero,(episode: JEDI)"`
}

type jsonTag struct {
	HeroObject struct {
		Name string
	} `json:"hero"`
}

func TestBuildQuery(t *testing.T) {
	tests := []struct {
		query interface{}
		name  string
	}{
		{&simple{}, "simple"},
		{&arguments{}, "arguments"},
		{&arguments2{}, "arguments2"},
		{&argumentsScalar{}, "argumentsScalar"},
		{&aliases{}, "aliases"},
		{&variables{}, "variables"},
		{&inlineFragments{}, "inlineFragments"},
		{&directives{}, "directives"},
		{&pointers{}, "pointers"},
		{&jsonTag{}, "jsonTag"},
	}

	for _, test := range tests {
		s, err := BuildQuery(test.query)
		if err != nil {
			t.Error(err)
			continue
		}
		golden.AssertBytes(t, s, "query."+test.name+".golden")
	}
}

func TestBuild_Mutation(t *testing.T) {
	type mutation struct {
		CreateReview struct {
			Stars      int
			Commentary string
		} `graphql:"(episode: $ep)"`

		GraphQLArguments struct {
			Episode Episode `graphql:"$ep"`
		}
	}

	s, err := Build(
		&mutation{},
		OperationTypeMutation,
		OperationName("CreateReviewForEpisode"),
	)
	if err != nil {
		t.Fatal(err)
	}

	expectedMutation := `mutation CreateReviewForEpisode($ep: Episode) {
  createReview(episode: $ep) {
    stars
    commentary
  }
}`
	assert.Equal(t, string(s), expectedMutation)
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		in  string
		out []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{"(a,b),c", []string{"(a,b)", "c"}},
		{"a,(b,c)", []string{"a", "(b,c)"}},
		{"a,(b,c),d", []string{"a", "(b,c)", "d"}},
		{"a,(b,c,d", []string{"a", "(b", "c", "d"}},
	}

	for _, test := range tests {
		assert.DeepEqual(t, parseTags(test.in), test.out)
	}
}

func TestBuilder_ToName(t *testing.T) {
	var b builder
	tests := []struct {
		in  string
		out string
	}{
		{"Foo", "foo"},
		{"FooBar", "fooBar"},
		{"URL", "url"},
		{"URLIn", "urlIn"},
		{"XMLName", "xmlName"},
		{"fooBar", "fooBar"},
	}

	for _, test := range tests {
		assert.Equal(t, b.toName(test.in), test.out)
	}
}
