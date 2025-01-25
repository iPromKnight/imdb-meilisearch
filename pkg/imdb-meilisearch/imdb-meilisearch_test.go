package imdb_meilisearch

import (
	"context"
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	meilisearchClient "github.com/ipromknight/imdb-meilisearch/internal/pkg/search/meilisearch"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"testing"
)

var searchClient *ImdbSearchClient

func TestMain(m *testing.M) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "getmeili/meilisearch:latest",
		ExposedPorts: []string{"7700/tcp"},
		Env:          map[string]string{"MEILI_ENV": "development"},
		WaitingFor:   wait.ForExposedPort(),
	}
	meiliC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Could not start meilisearch: %s", err)
	}

	endpoint, err := meiliC.Endpoint(ctx, "http")
	if err != nil {
		log.Fatalf("Could not get endpoint: %s", err)
	}

	clientConfig := meilisearchConfiguration.ClientOptions{
		Host: endpoint,
	}
	client, err := meilisearchClient.InitMeiliSearchClient(clientConfig)
	if err != nil {
		log.Fatalf("could not connect to meilisearch: %s", err)
	}
	res, err := client.AddDocuments([]map[string]interface{}{
		{"id": 38650, "imdb_id": "tt0038650", "title": "It's a Wonderful Life", "year": 1946, "category": "movie"},
		{"id": 63350, "imdb_id": "tt0063350", "title": "Night Of The Living Dead", "year": 1968, "category": "movie"},
		{"id": 140738, "imdb_id": "tt0140738", "title": "Flash Gordon", "year": 1954, "category": "series"},
		{"id": 123, "imdb_id": "tt0123", "title": "Flash Gordon", "year": 1953, "category": "series"},
		{"id": 124, "imdb_id": "tt01234", "title": "Flash Gordon", "year": 1954, "category": "movie"},
	}, "id")
	if err != nil {
		log.Fatalf("could not add documents: %s", err)
	}
	task, err := client.WaitForTask(res.TaskUID, 20)
	if err != nil || task.Status != "succeeded" {
		log.Fatalf("could not add documents: %s,%v", err, task)
	}

	searchClient, err = NewSearchClient(SearchClientConfig{MeiliSearchConfig: clientConfig})
	if err != nil {
		log.Fatalf("could not create search client: %s", err)
	}

	code := m.Run()
	// You can't defer this because os.Exit doesn't care for defer
	if err := meiliC.Terminate(ctx); err != nil {
		log.Fatalf("Could not stop meilisearch: %s", err)
	}

	os.Exit(code)
}

func TestImdbSearchClient_GetClosestImdbTitleMovieByFilename(t *testing.T) {
	imdbTitle, _ := searchClient.GetClosestImdbTitleForFilename("Night Of The Living Dead (1968) [BluRay] [720p] [YTS.AM]/Night.Of.The.Living.Dead.1968.720p.BluRay.x264-[YTS.AM].mp4")
	if imdbTitle.Title != "Night Of The Living Dead" {
		t.Errorf("expected title to be Night Of The Living Dead, got %s", imdbTitle.Title)
	}
}

func TestImdbSearchClient_GetClosestImdbTitleEmptyStringByFilename(t *testing.T) {
	imdbTitle, _ := searchClient.GetClosestImdbTitleForFilename("")
	if imdbTitle.Title != "" {
		t.Errorf("expected title to be empty, got %s", imdbTitle.Title)
	}
}

func TestImdbSearchClient_GetClosestImdbTitleSeriesByFilename(t *testing.T) {
	imdbTitle, _ := searchClient.GetClosestImdbTitleForFilename("FlashGordon.S01E01.ThePlanetOfDeath_512kb.mp4")
	if imdbTitle.Title != "Flash Gordon" {
		t.Errorf("expected title to be Flash Gordon, got %s", imdbTitle.Title)
	}
}

func TestImdbSearchClient_GetClosestImdbTitleSeriesWithYearByFilename(t *testing.T) {
	imdbTitle, _ := searchClient.GetClosestImdbTitleForFilename("FlashGordon.S01E01.1954.ThePlanetOfDeath_512kb.mp4")
	if imdbTitle.Title != "Flash Gordon" || imdbTitle.Id != "tt0140738" {
		t.Errorf("expected title to be Flash Gordon, got %s ,Id:%s", imdbTitle.Title, imdbTitle.Id)
	}
}

func TestImdbSearchClient_GetClosestImdbTitleMovie(t *testing.T) {
	imdbTitle, _ := searchClient.GetClosestImdbTitleForTitleAndYear("Night Of The Living Dead", "movie", 1968)
	if imdbTitle.Title != "Night Of The Living Dead" {
		t.Errorf("expected title to be Night Of The Living Dead, got %s", imdbTitle.Title)
	}
}

func TestImdbSearchClient_GetClosestImdbTitleEmptyString(t *testing.T) {
	imdbTitle, _ := searchClient.GetClosestImdbTitleForTitleAndYear("", "", 0)
	if imdbTitle.Title != "" {
		t.Errorf("expected title to be empty, got %s", imdbTitle.Title)
	}
}

func TestImdbSearchClient_GetClosestImdbTitleSeries(t *testing.T) {
	imdbTitle, _ := searchClient.GetClosestImdbTitleForTitleAndYear("FlashGordon", "series", 0)
	if imdbTitle.Title != "Flash Gordon" {
		t.Errorf("expected title to be Flash Gordon, got %s", imdbTitle.Title)
	}
}

func TestImdbSearchClient_GetClosestImdbTitleSeriesWithYear(t *testing.T) {
	imdbTitle, _ := searchClient.GetClosestImdbTitleForTitleAndYear("FlashGordon", "series", 1954)
	if imdbTitle.Title != "Flash Gordon" || imdbTitle.Id != "tt0140738" {
		t.Errorf("expected title to be Flash Gordon, got %s ,Id:%s", imdbTitle.Title, imdbTitle.Id)
	}
}
