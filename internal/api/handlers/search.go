package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/ipromknight/imdb-meilisearch/internal/api/types"
	imdbMeilisearch "github.com/ipromknight/imdb-meilisearch/pkg/imdb-meilisearch"
	"net/http"
	"strconv"
)

func SearchByFilename(service *imdbMeilisearch.ImdbSearchClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var filename = c.Query("filename")
		if filename == "" {
			c.Status(http.StatusBadRequest)
			return c.JSON(types.SearchErrorResponse(fmt.Errorf("missing filename query parameter")))
		}

		result, err := service.GetClosestImdbTitleForFilename(filename)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(types.SearchErrorResponse(err))
		}
		return c.JSON(types.SearchSuccessResponse(result))
	}
}

func SearchByTitle(service *imdbMeilisearch.ImdbSearchClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var title = c.Query("title")
		var category = c.Query("category")
		var year = c.Query("year")

		if title == "" {
			c.Status(http.StatusBadRequest)
			return c.JSON(types.SearchErrorResponse(fmt.Errorf("missing title query parameter")))
		}

		if category == "" {
			c.Status(http.StatusBadRequest)
			return c.JSON(types.SearchErrorResponse(fmt.Errorf("missing category query parameter")))
		}

		if year == "" {
			c.Status(http.StatusBadRequest)
			return c.JSON(types.SearchErrorResponse(fmt.Errorf("missing year query parameter")))
		}

		yearInt, err := strconv.Atoi(year)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return c.JSON(types.SearchErrorResponse(fmt.Errorf("invalid year query parameter")))
		}

		result, err := service.GetClosestImdbTitleForTitleAndYear(title, category, yearInt)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(types.SearchErrorResponse(err))
		}
		return c.JSON(types.SearchSuccessResponse(result))
	}
}
