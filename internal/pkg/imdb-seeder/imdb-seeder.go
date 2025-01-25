package imdb_seeder

import (
	"compress/gzip"
	"crypto/tls"
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	meilisearchclient "github.com/ipromknight/imdb-meilisearch/internal/pkg/search/meilisearch"
	"github.com/ipromknight/imdb-meilisearch/internal/pkg/tsv_reader"
	"github.com/meilisearch/meilisearch-go"
	"github.com/rs/zerolog"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var client = &http.Client{Transport: &http.Transport{
	TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	DisableCompression: false,
	DisableKeepAlives:  true,
	IdleConnTimeout:    20 * time.Second}}

func Seed(meilisearchConfig meilisearchConfiguration.ClientOptions, logger zerolog.Logger) error {
	index, err := meilisearchclient.InitMeiliSearchClient(meilisearchConfig)
	if err != nil {
		logger.Fatal().AnErr("error", err).Msg("could not connect to meilisearch")
		return err
	}

	var taskIds []int64

	titlemap := map[string]string{
		"movie":        "movie",
		"tvMovie":      "movie",
		"tvSeries":     "series",
		"tvShort":      "series",
		"tvMiniSeries": "series",
		"tvSpecial":    "series",
	}

	logger.Info().Msg("Writing titles to meilisearch...")

	req, _ := http.NewRequest("GET", "https://datasets.imdbws.com/title.basics.tsv.gz", nil)

	resp, err := client.Do(req)
	if err != nil {
		logger.Fatal().AnErr("error", err).Msg("failed to fetch imdb data")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Fatal().AnErr("error", err).Msg("failed to close response body")
		}
	}(resp.Body)

	gzreader, err := gzip.NewReader(resp.Body)
	if err != nil {
		logger.Fatal().AnErr("error", err).Msg("failed to read gzip data")
		return err
	}
	defer func(gzreader *gzip.Reader) {
		err := gzreader.Close()
		if err != nil {
			logger.Fatal().AnErr("error", err).Msg("failed to close gzip reader")
		}
	}(gzreader)
	parsertitle := tsv_reader.NewTabNewlineReader(gzreader)
	_, _ = parsertitle.Read()

	valueArgs := make([]map[string]interface{}, 0)
	var record []string
	var tsverr error
	var ok bool
	rowCount, insertCount := 0, 0
	for {
		record, tsverr = parsertitle.Read()
		if tsverr == io.EOF {
			break
		}
		if tsverr != nil {
			logger.Error().AnErr("error", tsverr).Msg("failed to read tsv record")
			continue
		}
		typeOfMedia := record[1]
		if _, ok = titlemap[typeOfMedia]; ok && typeOfMedia != "" {
			rowCount++
			id := record[0]
			idWithoutPrefix := csvgetint(strings.TrimLeft(strings.TrimPrefix(id, "tt"), "0"))
			title := html.UnescapeString(record[2])
			year := csvgetint(record[5])
			imdbRecord := map[string]interface{}{
				"id":       idWithoutPrefix,
				"imdb_id":  id,
				"title":    title,
				"year":     year,
				"category": titlemap[typeOfMedia],
			}
			insertCount++
			valueArgs = append(valueArgs, imdbRecord)
		}
		if len(valueArgs) > 9998 {
			taskInfo, err := index.AddDocuments(valueArgs, "id")
			if err != nil {
				logger.Error().AnErr("error", err).Msg("failed to add documents to meilisearch")
			}
			valueArgs = make([]map[string]interface{}, 0)
			taskIds = append(taskIds, taskInfo.TaskUID)
		}
	}

	if len(valueArgs) > 1 {
		taskInfo, err := index.AddDocuments(valueArgs, "id")
		if err != nil {
			logger.Error().AnErr("error", err).Msg("failed to add documents to meilisearch")
		}
		taskIds = append(taskIds, taskInfo.TaskUID)
	}
	for _, id := range taskIds {
		task, _ := index.WaitForTask(id, 30)
		if task.Status != "succeeded" {
			logger.Error().Str("status", string(task.Status)).Msg("task failed")
		}
	}
	logger.Info().Int("rows", rowCount).Int("inserted", insertCount).Msg("Finished writing titles to meilisearch")

	logger.Info().Msg("Creating index settings...")

	err = setIndexSettings(index, logger)
	if err != nil {
		logger.Error().AnErr("error", err).Msg("Failed to update index")
		return err
	}

	return nil
}
func csvgetint(instr string) int {
	getint, err := strconv.Atoi(instr)
	if err != nil {
		return 0
	}
	return getint
}

func setIndexSettings(index meilisearch.IndexManager, logger zerolog.Logger) error {
	distinctAttribute := "imdb_id"
	settings := meilisearch.Settings{
		RankingRules: []string{
			"words",
			"typo",
			"proximity",
			"attribute",
			"sort",
			"exactness",
		},
		DistinctAttribute: &distinctAttribute,
		SearchableAttributes: []string{
			"title",
			"year",
		},
		DisplayedAttributes: []string{
			"title",
			"category",
			"year",
			"imdb_id",
		},
		//StopWords: search.GetStopWords(),
		StopWords: make([]string, 0),
		SortableAttributes: []string{
			"title",
		},
		FilterableAttributes: []string{
			"category",
			"year",
		},
		Synonyms: map[string][]string{},
		TypoTolerance: &meilisearch.TypoTolerance{
			Enabled: true,
		},
		Pagination: &meilisearch.Pagination{
			MaxTotalHits: 1000,
		},
		Faceting: &meilisearch.Faceting{
			MaxValuesPerFacet: 200,
		},
		SearchCutoffMs:  200,
		SeparatorTokens: []string{".", "-", "_"},
	}

	taskInfo, err := index.UpdateSettings(&settings)
	if err != nil {
		return err
	}

	task, taskWaitError := index.WaitForTask(taskInfo.TaskUID, 30)
	if taskWaitError != nil {
		return taskWaitError
	}
	if task.Status != "succeeded" {
		logger.Error().Str("status", string(task.Status)).Msg("task failed")
	}
	return nil
}
