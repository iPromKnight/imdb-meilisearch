package meilisearch_client

import (
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	"github.com/meilisearch/meilisearch-go"
)

func InitMeiliSearchClient(clientConfig meilisearchConfiguration.ClientOptions) (meilisearch.IndexManager, error) {
	client := meilisearch.New(clientConfig.Host, meilisearch.WithAPIKey(clientConfig.ApiKey))
	index := client.Index("imdb")
	return index, nil
}
