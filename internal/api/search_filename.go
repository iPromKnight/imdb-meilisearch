package api

import (
	"encoding/json"
	"github.com/gofiber/fiber/v3"
)

func mapSearchByFilenameEndpoint(app *fiber.App) {
	app.Get("/search/filename", func(c fiber.Ctx) error {
		filename := c.Query("filename")
		if filename == "" {
			errorResponse, _ := json.Marshal(ResponseWrapper{
				Success: false,
				Error:   "Missing filename query parameter",
			})
			return c.SendString(string(errorResponse))
		}

		imdbMinimalTitle := imdbSearchClient.GetClosestImdbTitleForFilename(filename)
		if imdbMinimalTitle.Title == "" {
			errorResponse, _ := json.Marshal(ResponseWrapper{
				Success: false,
				Error:   "No title found for filename",
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
