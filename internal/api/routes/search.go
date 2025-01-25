package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ipromknight/imdb-meilisearch/internal/api/handlers"
	imdbMeilisearch "github.com/ipromknight/imdb-meilisearch/pkg/imdb-meilisearch"
)

func SearchRouter(app fiber.Router, service *imdbMeilisearch.ImdbSearchClient) {
	app.Get("/filename", handlers.SearchByFilename(service))
	app.Get("/title", handlers.SearchByTitle(service))
}
