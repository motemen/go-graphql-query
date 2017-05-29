package graphqlquery_test

import (
	"fmt"

	"github.com/motemen/go-graphql-query"
)

// a struct simply generates GraphQL query generating the result
// suitable for the struct
type simpleExample struct {
	Hero struct {
		Name    string
		Friends []struct {
			Name string
		}
	}
}

type complexExample struct {
	EmpireHero struct {
		// arguments can be specified by the special field GraphQLArguments
		// or "(..)" tag (see Hero below)
		GraphQLArguments struct {
			// arguments in GraphQLArguments are automatically shown in the query arguments
			Episode Episode `graphql:"$ep"`
		}
		Name string
	} `graphql:"aliasof=hero"` // use "aliasof=" tag to use alias

	Hero struct {
		Name string

		// use embedding and "..." tag to build inline fragments query
		DroidFields `graphql:"... on Droid"`
		// or "..." for a field
		Height int `graphql:"... on Human"`

		Friends []struct {
			Name string
		} `graphql:"@include(if: $withFriends)"` // directives
	} `graphql:"(episode: $ep)"` // you can use "(..)" tag to specify arguments

	// GraphQLArguments at toplevel stands for query arguments
	GraphQLArguments struct {
		// should include arguments appeared in struct tags
		WithFriends bool `graphql:"$withFriends,notnull"`
	}
}

type DroidFields struct {
	PrimaryFunction string
}

type Episode string // string types starting with capital letter are treated as custom types (TODO: custom object types, etc)

func ExampleBuild() {
	s, _ := graphqlquery.Build(&simpleExample{})
	c, _ := graphqlquery.Build(&complexExample{})
	fmt.Println(string(s))
	fmt.Println(string(c))
	// Output:
	// query {
	//   hero {
	//     name
	//     friends {
	//       name
	//     }
	//   }
	// }
	// query($ep: Episode, $withFriends: Boolean!) {
	//   empireHero: hero(episode: $ep) {
	//     name
	//   }
	//   hero(episode: $ep) {
	//     name
	//     ... on Droid {
	//       primaryFunction
	//     }
	//     ... on Human {
	//       height
	//     }
	//     friends @include(if: $withFriends) {
	//       name
	//     }
	//   }
	// }
}
