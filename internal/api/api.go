package api

import (
	"github.com/gofiber/fiber/v3"
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	imdbMeilisearch "github.com/ipromknight/imdb-meilisearch/pkg/imdb-meilisearch"
	"github.com/rs/zerolog"
)

type ResponseWrapper struct {
	Result  interface{} `json:"results,omitempty"`
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
}

var imdbSearchClient *imdbMeilisearch.ImdbSearchClient

func NewApi(options meilisearchConfiguration.ClientOptions, logger zerolog.Logger) (*fiber.App, error) {
	app := fiber.New()
	client, createClientError := createSearchClient(options, logger)
	if createClientError != nil {
		return nil, createClientError
	}
	imdbSearchClient = client

	mapIngestEndpoint(app, options, logger)
	mapSearchByFilenameEndpoint(app)
	mapSearchByTitleEndpoint(app)

	return app, nil
}

func Serve(app *fiber.App, logger zerolog.Logger) error {
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
