package imdb_seeder

import (
	"bytes"
	"compress/gzip"
	"context"
	mellisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	meilisearchClient "github.com/ipromknight/imdb-meilisearch/internal/pkg/search/meilisearch"
	"github.com/jarcoal/httpmock"
	"github.com/meilisearch/meilisearch-go"
	"github.com/rs/zerolog"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"net/http"
	"os"
	"testing"
)

var clientConfig mellisearchConfiguration.ClientOptions

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
		log.Fatalf("Could not start redis: %s", err)
	}

	endpoint, err := meiliC.Endpoint(ctx, "http")
	if err != nil {
		log.Fatalf("Could not get endpoint: %s", err)
	}

	clientConfig = mellisearchConfiguration.ClientOptions{
		Host: endpoint,
	}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	compressedString, err := compressString("tconst\ttitleType\tprimaryTitle\toriginalTitle\tisAdult\tstartYear\tendYear\truntimeMinutes\tgenres\ntt0000002   short\tLe clown et ses chiens\tLe clown et ses chiens\t0\t1892\t\\N\t5\tAnimation,Short\ntt0063350\tmovie\tNight of the Living Dead\tNight of the Living Dead\t0\t1968\t\\N\t96\tHorror,Thriller\ntt0038650\tmovie\tIt's a Wonderful Life\tIt's a Wonderful Life\t0\t1946\t\\N\t130\tDrama,Family,Fantasy\ntt0140738\ttvSeries\tFlash Gordon\tFlash Gordon\t0\t1954\t1955\t30\tAction,Adventure,Family")
	if err != nil {
		log.Fatalf("Could not compress string: %s", err)
	}

	httpmock.RegisterResponder("GET", "https://datasets.imdbws.com/title.basics.tsv.gz",
		httpmock.NewBytesResponder(200, compressedString).HeaderAdd(http.Header{"Content-Encoding": {"gzip"}}).HeaderAdd(http.Header{"Content-Length": {string(rune(len(compressedString)))}}))

	code := m.Run()
	// You can't defer this because os.Exit doesn't care for defer
	if err := meiliC.Terminate(ctx); err != nil {
		log.Fatalf("Could not stop meilisearch: %s", err)
	}

	os.Exit(code)
}

func TestSeed(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Logger()
	err := Seed(clientConfig, logger)
	if err != nil {
		return
	}

	searchClient, _ := meilisearchClient.InitMeiliSearchClient(clientConfig)
	var result meilisearch.DocumentsResult
	searchErr := searchClient.GetDocuments(&meilisearch.DocumentsQuery{
		Fields: []string{"imdb_id", "title", "year", "category"},
	}, &result)
	if searchErr != nil {
		t.Errorf("Could not get documents: %s", err)
	}

	if len(result.Results) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(result.Results))
	}

	for _, doc := range result.Results {
		switch doc["imdb_id"] {
		case "tt0140738":
			if doc["title"] != "Flash Gordon" {
				t.Errorf("Expected title to be 'Flash Gordon', got %s", doc["title"])
			}
			if doc["year"] != float64(1954) {
				t.Errorf("Expected year to be '1954', got %s", doc["year"])
			}
			if doc["category"] != "series" {
				t.Errorf("Expected category to be 'series', got %s", doc["category"])
			}
		case "tt0063350":
			if doc["title"] != "Night of the Living Dead" {
				t.Errorf("Expected title to be 'Night of the Living Dead', got %s", doc["title"])
			}
			if doc["year"] != float64(1968) {
				t.Errorf("Expected year to be '1968', got %s", doc["year"])
			}
			if doc["category"] != "movie" {
				t.Errorf("Expected category to be 'movie', got %s", doc["category"])
			}
		case "tt0038650":
			if doc["title"] != "It's a Wonderful Life" {
				t.Errorf("Expected title to be 'It's a Wonderful Life', got %s", doc["title"])
			}
			if doc["year"] != float64(1946) {
				t.Errorf("Expected year to be '1946', got %s", doc["year"])
			}
			if doc["category"] != "movie" {
				t.Errorf("Expected category to be 'movie', got %s", doc["category"])
			}
		default:
			t.Errorf("Unexpected imdb_id: %s", doc["imdb_id"])
		}
	}

}

func compressString(text string) ([]byte, error) {
	var buf bytes.Buffer

	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write([]byte(text))
	if err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
