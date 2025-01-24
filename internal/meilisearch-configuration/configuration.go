package meilisearch_configuration

import "os"

type ClientOptions struct {
	Host   string
	ApiKey string
}

func (clientOptions ClientOptions) PopulateFromEnv() ClientOptions {
	if clientOptions.Host == "" {
		clientOptions.Host = os.Getenv("MEILISEARCH_HOST")

		if clientOptions.Host == "" {
			clientOptions.Host = "http://127.0.0.1:7700"
		}
	}
	if clientOptions.ApiKey == "" {
		clientOptions.ApiKey = os.Getenv("MEILI_MASTER_KEY")
	}

	return clientOptions
}
