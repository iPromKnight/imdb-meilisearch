package api

import (
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"strconv"
)

func mapSearchByTitleEndpoint(app *fiber.App) {
	app.Get("/search/title", func(c fiber.Ctx) error {
		title := c.Query("title")
		category := c.Query("category")
		year := c.Query("year")
		if title == "" {
			errorResponse, _ := json.Marshal(ResponseWrapper{
				Success: false,
				Error:   "Missing title query parameter",
			})
			return c.SendString(string(errorResponse))
		}

		if category == "" {
			errorResponse, _ := json.Marshal(ResponseWrapper{
				Success: false,
				Error:   "Missing category query parameter",
			})
			return c.SendString(string(errorResponse))
		}

		if year == "" {
			errorResponse, _ := json.Marshal(ResponseWrapper{
				Success: false,
				Error:   "Missing year query parameter",
			})
			return c.SendString(string(errorResponse))
		}

		if category != "movie" && category != "series" {
			errorResponse, _ := json.Marshal(ResponseWrapper{
				Success: false,
				Error:   "Invalid category query parameter. Must be 'movie' or 'series'",
			})
			return c.SendString(string(errorResponse))
		}

		yearInt, err := strconv.Atoi(year)
		if err != nil {
			errorResponse, _ := json.Marshal(ResponseWrapper{
				Success: false,
				Error:   "Invalid year query parameter",
			})
			return c.SendString(string(errorResponse))
		}

		imdbMinimalTitle := imdbSearchClient.GetClosestImdbTitleForTitleAndYear(title, category, yearInt)
		if imdbMinimalTitle.Title == "" {
			errorResponse, _ := json.Marshal(ResponseWrapper{
				Success: false,
				Error:   "No title found for title and year",
			})
			return c.SendString(string(errorResponse))
		}

		successResponse, _ := json.Marshal(ResponseWrapper{
			Success: true,
			Result:  imdbMinimalTitle,
		})

		return c.SendString(string(successResponse))
	})
}
