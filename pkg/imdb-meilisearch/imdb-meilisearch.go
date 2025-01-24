package imdb_meilisearch

import (
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	"github.com/ipromknight/imdb-meilisearch/internal/pkg/search"
	meilisearchClient "github.com/ipromknight/imdb-meilisearch/internal/pkg/search/meilisearch"
	"github.com/meilisearch/meilisearch-go"
	"github.com/razsteinmetz/go-ptn"
	"github.com/rs/zerolog"
	"os"
	"strconv"
)

type ImdbSearchClient struct {
	index  meilisearch.IndexManager
	logger zerolog.Logger
}
type ImdbMinimalTitle struct {
	Id    string
	Type  string
	Title string
	Score float64
}

type SearchClientConfig struct {
	MeiliSearchConfig meilisearchConfiguration.ClientOptions
	Logger            zerolog.Logger
}

type SearchQuery struct {
	Title     string
	TitleType string
	Year      int
	Filename  string
}

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
	return searchClient, nil
}

func (searchClient *ImdbSearchClient) GetClosestImdbTitleForFilename(filename string) ImdbMinimalTitle {
	info, err := ptn.Parse(filename)
	if err != nil {
		return ImdbMinimalTitle{}
	}
	titleType := "series"
	if info.IsMovie {
		titleType = "movie"
	}
	return searchClient.GetClosestImdbTitleForTitleAndYear(info.Title, titleType, info.Year)
}

func (searchClient *ImdbSearchClient) GetClosestImdbTitleForTitleAndYear(title string, titleType string, year int) ImdbMinimalTitle {
	if len(title) < 2 {
		return ImdbMinimalTitle{}
	}
	var imdbMinimal ImdbMinimalTitle
	if title == "" {
		return ImdbMinimalTitle{}
	}
	var filters interface{}
	if year != 0 {
		filters = "year < " + strconv.Itoa(year+1) + " AND year > " + strconv.Itoa(year-1)
	} else {
		filters = nil
	}
	if titleType == "series" {
		if filters == nil {
			filters = "title_type = series"
		} else {
			filters = filters.(string) + " AND title_type = series"
		}

	}

	searchRes, err := searchClient.index.Search(search.NormalizeString(title),
		&meilisearch.SearchRequest{
			Limit:                1,
			AttributesToSearchOn: []string{"title"},
			Filter:               filters,
			ShowRankingScore:     true,
		})
	if err != nil {
		searchClient.logger.Error().AnErr("error", err).Msg("could not search meilisearch")
		return ImdbMinimalTitle{}
	}
	var hit map[string]interface{}
	for _, result := range searchRes.Hits {
		hit = result.(map[string]interface{})
		imdbMinimal.Id = hit["imdb_id"].(string)
		imdbMinimal.Type = hit["title_type"].(string)
		imdbMinimal.Title = hit["title"].(string)
		imdbMinimal.Score = hit["_rankingScore"].(float64)
		break
	}

	return imdbMinimal

}
