package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipromknight/imdb-meilisearch/internal/api/types"
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	seeder "github.com/ipromknight/imdb-meilisearch/internal/pkg/imdb-seeder"
	"github.com/rs/zerolog"
	"net/http"
)

func IngestImdbDataSet(options meilisearchConfiguration.ClientOptions, logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		seedError := seeder.Seed(options, logger)
		if seedError != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(types.SearchErrorResponse(seedError))
		}

		return c.JSON(types.IngestSuccessResponse())
	}
}
