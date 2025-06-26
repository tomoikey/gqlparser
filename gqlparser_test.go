package gqlparser_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

func TestLoadQueryWithRules(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{Input: `
		type Query {
			hello: String!
			user(id: ID!): User
		}
		
		type User {
			id: ID!
			name: String!
		}
	`})

	t.Run("valid query", func(t *testing.T) {
		query, errs := gqlparser.LoadQueryWithRules(schema, `{ hello }`, nil)
		require.Empty(t, errs)
		require.NotNil(t, query)
	})

	t.Run("invalid query - unknown field", func(t *testing.T) {
		_, errs := gqlparser.LoadQueryWithRules(schema, `{ unknown }`, nil)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Message, "Cannot query field")
	})

	t.Run("syntax error", func(t *testing.T) {
		_, errs := gqlparser.LoadQueryWithRules(schema, `{ hello `, nil)
		require.NotEmpty(t, errs)
	})

	t.Run("with custom rules", func(t *testing.T) {
		customRules := rules.NewDefaultRules()
		query, errs := gqlparser.LoadQueryWithRules(schema, `{ hello }`, customRules)
		require.Empty(t, errs)
		require.NotNil(t, query)
	})

	t.Run("nil schema", func(t *testing.T) {
		_, errs := gqlparser.LoadQueryWithRules(nil, `{ hello }`, nil)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Message, "cannot validate as Schema is nil")
	})
}

func TestMustLoadQueryWithRules(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{Input: `
		type Query {
			hello: String!
			user(id: ID!): User
		}
		
		type User {
			id: ID!
			name: String!
		}
	`})

	t.Run("valid query", func(t *testing.T) {
		query := gqlparser.MustLoadQueryWithRules(schema, `{ hello }`, nil)
		require.NotNil(t, query)
		require.Equal(t, 1, len(query.Operations))
	})

	t.Run("invalid query - panics", func(t *testing.T) {
		require.Panics(t, func() {
			gqlparser.MustLoadQueryWithRules(schema, `{ unknown }`, nil)
		})
	})

	t.Run("syntax error - panics", func(t *testing.T) {
		require.Panics(t, func() {
			gqlparser.MustLoadQueryWithRules(schema, `{ hello `, nil)
		})
	})

	t.Run("with custom rules", func(t *testing.T) {
		customRules := rules.NewDefaultRules()
		query := gqlparser.MustLoadQueryWithRules(schema, `{ hello }`, customRules)
		require.NotNil(t, query)
	})

	t.Run("nil schema - panics", func(t *testing.T) {
		require.Panics(t, func() {
			gqlparser.MustLoadQueryWithRules(nil, `{ hello }`, nil)
		})
	})
}

// Tests identical to LoadQuery
func TestLoadQuery(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{Input: `
		type Query {
			hello: String!
			user(id: ID!): User
		}
		
		type User {
			id: ID!
			name: String!
		}
	`})

	t.Run("valid query", func(t *testing.T) {
		query, errs := gqlparser.LoadQuery(schema, `{ hello }`)
		require.Empty(t, errs)
		require.NotNil(t, query)
	})

	t.Run("invalid query - unknown field", func(t *testing.T) {
		_, errs := gqlparser.LoadQuery(schema, `{ unknown }`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Message, "Cannot query field")
	})

	t.Run("syntax error", func(t *testing.T) {
		_, errs := gqlparser.LoadQuery(schema, `{ hello `)
		require.NotEmpty(t, errs)
	})

	t.Run("nil schema", func(t *testing.T) {
		_, errs := gqlparser.LoadQuery(nil, `{ hello }`)
		require.NotEmpty(t, errs)
		require.Contains(t, errs[0].Message, "cannot validate as Schema is nil")
	})
}

// Tests identical to MustLoadQuery
func TestMustLoadQuery(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{Input: `
		type Query {
			hello: String!
			user(id: ID!): User
		}
		
		type User {
			id: ID!
			name: String!
		}
	`})

	t.Run("valid query", func(t *testing.T) {
		query := gqlparser.MustLoadQuery(schema, `{ hello }`)
		require.NotNil(t, query)
		require.Equal(t, 1, len(query.Operations))
	})

	t.Run("invalid query - panics", func(t *testing.T) {
		require.Panics(t, func() {
			gqlparser.MustLoadQuery(schema, `{ unknown }`)
		})
	})

	t.Run("syntax error - panics", func(t *testing.T) {
		require.Panics(t, func() {
			gqlparser.MustLoadQuery(schema, `{ hello `)
		})
	})

	t.Run("nil schema - panics", func(t *testing.T) {
		require.Panics(t, func() {
			gqlparser.MustLoadQuery(nil, `{ hello }`)
		})
	})
}