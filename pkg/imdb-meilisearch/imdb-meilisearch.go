package imdb_meilisearch

import (
	"fmt"
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	"github.com/ipromknight/imdb-meilisearch/internal/pkg/search"
	meilisearchClient "github.com/ipromknight/imdb-meilisearch/internal/pkg/search/meilisearch"
	"github.com/meilisearch/meilisearch-go"
	"github.com/razsteinmetz/go-ptn"
	"github.com/rs/zerolog"
	"os"
)

type ImdbSearchClient struct {
	index  meilisearch.IndexManager
	logger zerolog.Logger
}
type ImdbMinimalTitle struct {
	Id       string  `json:"imdb_id"`
	Title    string  `json:"title"`
	Year     float64 `json:"year"`
	Category string  `json:"category"`
	Score    float64 `json:"score"`
}

type SearchClientConfig struct {
	MeiliSearchConfig     meilisearchConfiguration.ClientOptions
	RankingScoreThreshold float64
	Logger                zerolog.Logger
}

type SearchQuery struct {
	Title                 string
	TitleType             string
	Year                  int
	Filename              string
	RankingScoreThreshold float64
}

var clientOptions SearchClientConfig

func NewSearchClient(searchClientConfig SearchClientConfig) (*ImdbSearchClient, error) {
	if searchClientConfig.Logger.GetLevel() == zerolog.NoLevel {
		searchClientConfig.Logger = zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	}
	searchClientConfig.Logger.Debug().Str("Host", searchClientConfig.MeiliSearchConfig.Host).Msg("Initializing MeiliSearch client")
	index, err := meilisearchClient.InitMeiliSearchClient(searchClientConfig.MeiliSearchConfig)
	if err != nil {
		return nil, err
	}
	searchClient := &ImdbSearchClient{index: index, logger: searchClientConfig.Logger}
	clientOptions = searchClientConfig
	return searchClient, nil
}

func (searchClient *ImdbSearchClient) GetClosestImdbTitleForFilename(filename string) (*ImdbMinimalTitle, error) {
	info, err := ptn.Parse(filename)
	if err != nil {
		return &ImdbMinimalTitle{}, err
	}
	titleType := "series"
	if info.IsMovie {
		titleType = "movie"
	}
	return searchClient.GetClosestImdbTitleForTitleAndYear(info.Title, titleType, info.Year)
}

func (searchClient *ImdbSearchClient) GetClosestImdbTitleForTitleAndYear(title string, titleType string, year int) (*ImdbMinimalTitle, error) {
	if len(title) < 2 {
		return &ImdbMinimalTitle{}, fmt.Errorf("title must be at least 2 characters long")
	}
	var imdbMinimal ImdbMinimalTitle
	if title == "" {
		return &ImdbMinimalTitle{}, fmt.Errorf("title must be provided")
	}

	var filters []string
	handleYear(year, &filters)
	handleSeriesType(titleType, &filters)
	handleMovieType(titleType, &filters)

	searchRequest := &meilisearch.SearchRequest{
		Limit:                 1,
		AttributesToSearchOn:  []string{"title", "year"},
		AttributesToHighlight: []string{"title", "year"},
		Filter:                filters,
		ShowRankingScore:      true,
	}

	if clientOptions.RankingScoreThreshold > 0 {
		searchRequest.RankingScoreThreshold = clientOptions.RankingScoreThreshold
	}

	searchRes, err := searchClient.index.Search(search.NormalizeString(title), searchRequest)
	if err != nil {
		searchClient.logger.Error().AnErr("error", err).Msg("could not search meilisearch")
		return &ImdbMinimalTitle{}, err
	}
	var hit map[string]interface{}
	for _, result := range searchRes.Hits {
		hit = result.(map[string]interface{})
		imdbMinimal.Id = hit["imdb_id"].(string)
		imdbMinimal.Title = hit["title"].(string)
		imdbMinimal.Year = hit["year"].(float64)
		imdbMinimal.Category = hit["category"].(string)
		imdbMinimal.Score = hit["_rankingScore"].(float64)
		break
	}

	return &imdbMinimal, nil
}

func handleSeriesType(titleType string, filters *[]string) {
	if titleType == "series" {
		*filters = append(*filters, `category = "series"`)
	}
}

func handleMovieType(titleType string, filters *[]string) {
	if titleType == "movie" {
		*filters = append(*filters, `category = "movie"`)
	}
}

func handleYear(year int, filters *[]string) {
	if year != 0 {
		*filters = append(*filters, fmt.Sprintf("year < %d AND year > %d", year+1, year-1))
	}
}
