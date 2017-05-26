package graphqlquery

import (
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
			Episode int `graphql:"EMPIRE"`
		}
		Name string
	} `graphql:"aliasof=hero"`
	JediHero struct {
		GraphQLArguments struct {
			Episode int `graphql:"JEDI"`
		}
		Name string
	} `graphql:"aliasof=hero"`
}

type Episode int

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

func TestToString(t *testing.T) {
	tests := []interface{}{
		&simple{},
		&withArguments{},
		&withArguments2{},
		&withArgumentsScalar{},
		&withAliases{},
		&withVariables{},
	}

	for _, test := range tests {
		s, err := New(test).String()
		t.Log(s, err)
	}
}

func TestParseTags(t *testing.T) {
	t.Logf("%v", parseTags("a,b,c"))
	t.Logf("%v", parseTags("(a,b),c"))
	t.Logf("%v", parseTags("a,(b,c)"))
	t.Logf("%v", parseTags("a,(b,c),d"))
	t.Logf("%v", parseTags("a,(b,c,d),e"))
	t.Logf("%v", parseTags("a,(b,c,d,e"))
}
