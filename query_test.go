package graphqlquery

import (
	"fmt"
	"testing"
)

type simple struct {
	Hero struct {
		Name string
	}
}

type withArguments struct {
	Human struct {
		GraphQLArguments struct {
			Id string `graphql:"\"1000\""`
		}
		Name   string
		Height int
	}
}

type withArguments2 struct {
	Human struct {
		Name   string
		Height int
	} `graphql:"(id: \"1000\")"`
}

type withArgumentsScalar struct {
	Human struct {
		GraphQLArguments struct {
			Id string `graphql:"\"1000\""`
		}
		Name   string
		Height int `graphql:"(unit: FOOT)"`
	}
}

type withAliases struct {
	EmpireHero struct {
		GraphQLArguments struct {
			Episode string `graphql:"EMPIRE"`
		}
		Name string
	} `graphql:"aliasof=hero"`
	JediHero struct {
		GraphQLArguments struct {
			Episode string `graphql:"JEDI"`
		}
		Name string
	} `graphql:"aliasof=hero"`
}

type Episode string

type withVariables struct {
	Hero struct {
		GraphQLArguments struct {
			Episode Episode `graphql:"$episode"`
		}
		Friends []struct {
			Name string
		}
	}
}

type withInlineFragments struct {
	GraphQLArguments struct {
		Episode Episode `graphql:"$ep,notnull"`
	}
	Hero struct {
		Name        string
		DroidFields `graphql:"... on Droid"`
		HumanFields `graphql:"... on Human"`
	} `graphql:"(episode: $ep)"`
}

type DroidFields struct {
	PrimaryFunction string
}

type HumanFields struct {
	Height int
}

func TestToString(t *testing.T) {
	tests := []interface{}{
		&simple{},
		&withArguments{},
		&withArguments2{},
		&withArgumentsScalar{},
		&withAliases{},
		&withVariables{},
		&withInlineFragments{},
	}

	for _, test := range tests {
		s, err := New(test).String()
		t.Log(s, err)
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"a,b,c", "[a b c]"},
		{"(a,b),c", "[(a,b) c]"},
		{"a,(b,c)", "[a (b,c)]"},
		{"a,(b,c),d", "[a (b,c) d]"},
		{"a,(b,c,d", "[a (b c d]"},
	}

	for _, test := range tests {
		if got := fmt.Sprintf("%v", parseTags(test.in)); got != test.out {
			t.Errorf("expected %v but got %v", test.out, got)
		}
	}
}
