package main

import "github.com/graphql-go/graphql"

// newSchema membangun skema GraphQL yang menutup atas store (sumber data).
func newSchema(s *store) (graphql.Schema, error) {
	bookType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Book",
		Fields: graphql.Fields{
			"id":       &graphql.Field{Type: graphql.Int},
			"title":    &graphql.Field{Type: graphql.String},
			"authorId": &graphql.Field{Type: graphql.Int},
		},
	})

	authorType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Author",
		Fields: graphql.Fields{
			"id":   &graphql.Field{Type: graphql.Int},
			"name": &graphql.Field{Type: graphql.String},
			// RELASI: field "books" di-resolve saat diminta -> ini titik rawan
			// N+1 (satu query per author). DataLoader mengatasinya (lihat README).
			"books": &graphql.Field{
				Type: graphql.NewList(bookType),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					a := p.Source.(Author)
					return s.booksByAuthor(a.ID), nil
				},
			},
		},
	})

	// Query root: entry point untuk membaca data.
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"authors": &graphql.Field{
				Type:    graphql.NewList(authorType),
				Resolve: func(p graphql.ResolveParams) (any, error) { return s.allAuthors(), nil },
			},
		},
	})

	// Mutation root: entry point untuk mengubah data.
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"addBook": &graphql.Field{
				Type: bookType,
				Args: graphql.FieldConfigArgument{
					"title":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"authorId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					title := p.Args["title"].(string)
					authorID := p.Args["authorId"].(int)
					return s.addBook(title, authorID), nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}
