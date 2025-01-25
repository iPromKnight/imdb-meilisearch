package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipromknight/imdb-meilisearch/internal/api/routes"
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	imdbMeilisearch "github.com/ipromknight/imdb-meilisearch/pkg/imdb-meilisearch"
	"github.com/rs/zerolog"
)

func ServeApi(options meilisearchConfiguration.ClientOptions, logger zerolog.Logger) error {
	app := fiber.New()
	client, createClientError := createSearchClient(options, logger)
	if createClientError != nil {
		return createClientError
	}

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Send([]byte("Imdb-Meilisearch!"))
	})

	search := app.Group("/search")
	general := app.Group("/general")

	routes.SearchRouter(search, client)
	routes.GeneralRouter(general, options, logger)

	logger.Info().Msg("Starting API server")
	return app.Listen(":8080")
}

func createSearchClient(options meilisearchConfiguration.ClientOptions, logger zerolog.Logger) (*imdbMeilisearch.ImdbSearchClient, error) {
	imdbClient, err := imdbMeilisearch.NewSearchClient(imdbMeilisearch.SearchClientConfig{
		MeiliSearchConfig: options,
		Logger:            logger,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create IMDB Search Client")
		return nil, err
	}

	return imdbClient, nil
}
