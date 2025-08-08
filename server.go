package main

import (
	"net/http"
	"os"
	"quizfreely/api/graph"
	"quizfreely/api/auth"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/go-chi/chi/v5"
)

const defaultPort = "8008"

func main() {
	_ = godotenv.Load()
	/* godotenv means go dotenv, not godot env*/

	if os.Getenv("PRETTY_LOG") == "true" {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{Out: os.Stderr},
		)
	} else {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal().Msgf(
			`DB_URL is not set
Copy .env.example to .env and/or
check your environment variables`,
		)
	}

	var err error
	var dbPool *pgxpool.Pool
	dbPool, err = pgxpool.New(
		context.Background(),
		dbUrl,
	)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error creating database pool")
	}
	defer dbPool.Close()

	router := chi.NewRouter()

	authHandler := &auth.AuthHandler{DB: dbPool}

	router.Post(
		"/auth/sign-up",
		authHandler.SignUpHandler,
	)
	router.Post(
		"/auth/sign-in",
		authHandler.SignInHandler,
	)

	router.Group(func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)

		srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: dbPool}}))

		srv.AddTransport(transport.Options{})
		srv.AddTransport(transport.GET{})
		srv.AddTransport(transport.POST{})

		srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

		srv.Use(extension.Introspection{})
		srv.Use(extension.AutomaticPersistedQuery{
			Cache: lru.New[string](100),
		})

		r.Handle(
			"/graphiql",
			playground.Handler(
				"Quizfreely API GraphiQL",
				"/graphql",
			),
		)
		r.Handle("/graphql", srv)
	})

	log.Info().Msg(
		"http://localhost:" + port + "/graphiql for GraphiQL",
	)
	log.Fatal().Err(
		http.ListenAndServe(":"+port, router),
	).Msgf("Error starting server")
}
