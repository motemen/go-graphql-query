package graphqlquery

import (
	"fmt"
	"strings"
	"testing"
)

type simple struct {
	Hero struct {
		Name string
	}
}

var simpleQuery = `
query {
  hero {
    name
  }
}`

type withArguments struct {
	Human struct {
		GraphQLArguments struct {
			Id string `graphql:"\"1000\""`
		}
		Name   string
		Height int
	}
}

var withArgumentsQuery = `
query {
  human(id: "1000") {
    name
    height
  }
}`

type withArguments2 struct {
	Human struct {
		Name   string
		Height int
	} `graphql:"(id: \"1000\")"`
}

var withArguments2Query = `
query {
  human(id: "1000") {
    name
    height
  }
}`

type withArgumentsScalar struct {
	Human struct {
		GraphQLArguments struct {
			Id string `graphql:"\"1000\""`
		}
		Name   string
		Height int `graphql:"(unit: FOOT)"`
	}
}

var withArgumentsScalarQuery = `
query {
  human(id: "1000") {
    name
    height(unit: FOOT)
  }
}`

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

var withAliasesQuery = `
query {
  empireHero: hero(episode: EMPIRE) {
    name
  }
  jediHero: hero(episode: JEDI) {
    name
  }
}`

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

var withVariablesQuery = `
query($episode: Episode) {
  hero(episode: $episode) {
    friends {
      name
    }
  }
}`

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

var withInlineFragmentsQuery = `
query($ep: Episode!) {
  hero(episode: $ep) {
    name
    ... on Droid {
      primaryFunction
    }
    ... on Human {
      height
    }
  }
}`

type withPointers struct {
	EmpireHero *struct {
		Name string
	} `graphql:"aliasof=hero,(episode: EMPIRE)"`
	JediHero *struct {
		Name string
	} `graphql:"aliasof=hero,(episode: JEDI)"`
}

var withPointersQuery = `
query {
  empireHero: hero(episode: EMPIRE) {
    name
  }
  jediHero: hero(episode: JEDI) {
    name
  }
}`

func TestToString(t *testing.T) {
	tests := []struct {
		query  interface{}
		result string
	}{
		{&simple{}, simpleQuery},
		{&withArguments{}, withArgumentsQuery},
		{&withArguments2{}, withArguments2Query},
		{&withArgumentsScalar{}, withArgumentsScalarQuery},
		{&withAliases{}, withAliasesQuery},
		{&withVariables{}, withVariablesQuery},
		{&withInlineFragments{}, withInlineFragmentsQuery},
		{&withPointers{}, withPointersQuery},
	}

	for _, test := range tests {
		s, err := Build(test.query)
		if err != nil {
			t.Error(err)
			continue
		}
		if expected := strings.TrimSpace(test.result); string(s) != expected {
			t.Errorf("===== got:\n%s\n===== but expected:\n%s\n", string(s), expected)
		}
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
		if got := fmt.Sprintf("%v", b.toName(test.in)); got != test.out {
			t.Errorf("expected %v but got %v", test.out, got)
		}
	}
}
