package types

import (
	"github.com/gofiber/fiber/v2"
	imdbMeilisearch "github.com/ipromknight/imdb-meilisearch/pkg/imdb-meilisearch"
)

func SearchSuccessResponse(data *imdbMeilisearch.ImdbMinimalTitle) *fiber.Map {
	return &fiber.Map{
		"status": true,
		"data":   &data,
		"error":  nil,
	}
}

func SearchErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"data":   "",
		"error":  err.Error(),
	}
}

func IngestSuccessResponse() *fiber.Map {
	return &fiber.Map{
		"status": true,
		"data":   "",
		"error":  nil,
	}
}
