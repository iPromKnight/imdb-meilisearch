package cmd

import (
	"fmt"
	mellisearchclient "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	imdbMeilisearch "github.com/ipromknight/imdb-meilisearch/pkg/imdb-meilisearch"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
	"time"
)

type searchCommandConfig struct {
	imdbMeilisearch.SearchQuery
	mellisearchclient.ClientOptions
}

var searchOptions searchCommandConfig

func RegisterSearchCommand(rootCmd *cobra.Command) {
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search for a title.",
		Long:  `Perform a query for a title based on either filename, or (title, category and year).`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			searchOptions.ClientOptions = searchOptions.ClientOptions.PopulateFromEnv()

			if searchOptions.Title == "" && searchOptions.Filename == "" {
				return fmt.Errorf("required flag 'search-title' or 'search-filename' is not set")
			}

			if searchOptions.Title != "" && searchOptions.Filename != "" {
				return fmt.Errorf("only one of 'search-title' or 'search-filename' can be set")
			}

			if searchOptions.Title != "" && searchOptions.TitleType != "movie" && searchOptions.TitleType != "series" {
				return fmt.Errorf("search-type must be either 'movie' or 'series' if title is set")
			}

			if searchOptions.Year < 0 {
				return fmt.Errorf("search-year must be a positive integer")
			}

			if searchOptions.Host == "" {
				return fmt.Errorf("required flag 'meili-host' is not set and the fallback environment variable 'MEILISEARCH_HOST' is not set")
			}
			if searchOptions.ApiKey == "" {
				return fmt.Errorf("required flag 'meili-api-key' is not set and the fallback environment variable 'MEILI_MASTER_KEY' is not set")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			debugEnabled, _ := cmd.Flags().GetBool("debug")
			logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
			if debugEnabled {
				logger = zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Logger()
			}

			logger.Debug().Str("Host", searchOptions.Host).Msg("Using Meilisearch Host")
			logger.Debug().Str("ApiKey", searchOptions.ApiKey).Msg("Using Meilisearch Api Key")

			performSearch(searchOptions, logger)
		},
	}

	searchCmd.PersistentFlags().StringVar(&searchOptions.Title, "search-title", "", "Search Title")
	searchCmd.PersistentFlags().StringVar(&searchOptions.TitleType, "search-type", "", "Search Category type - can be movie or series.")
	searchCmd.PersistentFlags().IntVar(&searchOptions.Year, "search-year", 0, "Search Year")
	searchCmd.PersistentFlags().StringVar(&searchOptions.Filename, "search-filename", "", "Search Filename")

	rootCmd.AddCommand(searchCmd)
}

func performSearch(options searchCommandConfig, logger zerolog.Logger) {
	imdbClient, err := imdbMeilisearch.NewSearchClient(imdbMeilisearch.SearchClientConfig{
		MeiliSearchConfig: options.ClientOptions,
		Logger:            logger,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create IMDB Search Client")
		return
	}

	start := time.Now()

	var imdbMinimal imdbMeilisearch.ImdbMinimalTitle
	if options.Filename != "" {
		imdbMinimal = imdbClient.GetClosestImdbTitleForFilename(options.Filename)
	} else {
		imdbMinimal = imdbClient.GetClosestImdbTitleForTitleAndYear(options.Title, options.TitleType, options.Year)
	}

	if imdbMinimal.Title == "" {
		logger.Info().Msg("No title found")
		return
	}

	logger.Info().Str("Title", imdbMinimal.Title).Str("Type", imdbMinimal.Type).Str("Imdb Id", imdbMinimal.Id).Float64("score", imdbMinimal.Score).Msg("Best Match")

	elapsed := time.Since(start)

	fmt.Printf("The query took %s\n", elapsed)
}
