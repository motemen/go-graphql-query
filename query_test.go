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

type arguments struct {
	Human []struct {
		GraphQLArguments struct {
			ID string `graphql:"\"1000\""`
		}
		Name   string
		Height int
	}
}

var argumentsQuery = `
query {
  human(id: "1000") {
    name
    height
  }
}`

type arguments2 struct {
	Human []struct {
		Name   string
		Height int
	} `graphql:"(id: \"1000\")"`
}

var arguments2Query = `
query {
  human(id: "1000") {
    name
    height
  }
}`

type argumentsScalar struct {
	Human struct {
		GraphQLArguments struct {
			ID string `graphql:"\"1000\""`
		}
		Name   string
		Height int `graphql:"(unit: FOOT)"`
	}
}

var argumentsScalarQuery = `
query {
  human(id: "1000") {
    name
    height(unit: FOOT)
  }
}`

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

var aliasesQuery = `
query {
  empireHero: hero(episode: EMPIRE) {
    name
  }
  jediHero: hero(episode: JEDI) {
    name
  }
}`

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

var variablesQuery = `
query($episode: Episode) {
  hero(episode: $episode) {
    friends {
      name
    }
  }
}`

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

var inlineFragmentsQuery = `
query($ep: Episode!) {
  hero(episode: $ep) {
    name
    ... on Human {
      height
    }
    ... on Droid {
      primaryFunction
    }
  }
}`

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

var directivesQuery = `
query($withFriends: Boolean) {
  hero {
    name
    friends @include(if: $withFriends) {
      name
    }
  }
}`

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

var pointersQuery = `
query {
  empireHero: hero(episode: EMPIRE) {
    name
  }
  jediHero: hero(episode: JEDI) {
    name
  }
}`

type jsonTag struct {
	HeroObject struct {
		Name string
	} `json:"hero"`
}

var jsonTagQuery = `
query {
  hero {
    name
  }
}`

func TestToString(t *testing.T) {
	tests := []struct {
		query  interface{}
		result string
	}{
		{&simple{}, simpleQuery},
		{&arguments{}, argumentsQuery},
		{&arguments2{}, arguments2Query},
		{&argumentsScalar{}, argumentsScalarQuery},
		{&aliases{}, aliasesQuery},
		{&variables{}, variablesQuery},
		{&inlineFragments{}, inlineFragmentsQuery},
		{&directives{}, directivesQuery},
		{&pointers{}, pointersQuery},
		{&jsonTag{}, jsonTagQuery},
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
