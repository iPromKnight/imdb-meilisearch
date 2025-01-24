package api

import (
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	seeder "github.com/ipromknight/imdb-meilisearch/internal/pkg/imdb-seeder"
	"github.com/rs/zerolog"
)

func mapIngestEndpoint(app *fiber.App, options meilisearchConfiguration.ClientOptions, logger zerolog.Logger) {
	app.Get("/ingest", func(c fiber.Ctx) error {

		seedError := seeder.Seed(options, logger)
		if seedError != nil {
			errorResponse, _ := json.Marshal(ResponseWrapper{
				Success: false,
				Error:   seedError.Error(),
			})
			return c.SendString(string(errorResponse))
		}

		successResponse, _ := json.Marshal(ResponseWrapper{
			Success: true,
		})

		return c.SendString(string(successResponse))
	})
}
