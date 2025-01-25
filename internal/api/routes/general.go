package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipromknight/imdb-meilisearch/internal/api/handlers"
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	"github.com/rs/zerolog"
)

func GeneralRouter(app fiber.Router, options meilisearchConfiguration.ClientOptions, logger zerolog.Logger) {
	app.Get("/ingest", handlers.IngestImdbDataSet(options, logger))
}
