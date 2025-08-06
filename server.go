package main

import (
	"log"
	"net/http"
	"os"
	"quizfreely/api/graph"
	"quizfreely/api/dbpool"

	"github.com/joho/godotenv"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8008"

func main() {
	_ = godotenv.Load()
	/* godotenv means go dotenv, not godot env*/

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal(`DB_URL is not set
Copy .env.example to .env and/or
check your environment variables`,
		)
	}

	dbpool.Init(dbUrl)
	defer dbpool.Pool.Close()

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/graphiql", playground.Handler("Quizfreely API GraphiQL", "/graphql"))
	http.Handle("/graphql", srv)

	log.Printf("http://localhost:%s/graphiql for GraphiQL", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
