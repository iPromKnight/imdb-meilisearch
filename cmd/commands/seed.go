package cmd

import (
	"fmt"
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	seeder "github.com/ipromknight/imdb-meilisearch/internal/pkg/imdb-seeder"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
)

type seederCommandConfig struct {
	meilisearchConfiguration.ClientOptions
}

var seedOptions seederCommandConfig

func RegisterSeedCommand(rootCmd *cobra.Command) {
	seedCmd := &cobra.Command{
		Use:   "seed",
		Short: "This is a CLI tool to seed the IMDB database into MeiliSearch.",
		Long:  `This should be run as a cron job to keep the MeiliSearch database up to date with the latest IMDB data.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			seedOptions.ClientOptions = seedOptions.ClientOptions.PopulateFromEnv()

			if seedOptions.Host == "" {
				return fmt.Errorf("required flag 'host' is not set and the fallback environment variable 'MEILISEARCH_HOST' is not set")
			}
			if seedOptions.ApiKey == "" {
				return fmt.Errorf("required flag 'api-key' is not set and the fallback environment variable 'MEILI_MASTER_KEY' is not set")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			debugEnabled, _ := cmd.Flags().GetBool("debug")
			logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
			if debugEnabled {
				logger = zerolog.New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Logger()
			}

			logger.Debug().Str("Host", seedOptions.Host).Msg("Using Meilisearch Host")
			logger.Debug().Str("ApiKey", seedOptions.ApiKey).Msg("Using Meilisearch Api Key")

			err := seeder.Seed(seedOptions.ClientOptions, logger)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to seed MeiliSearch")
			} else {
				logger.Info().Msg("Successfully seeded MeiliSearch")
			}
		},
	}

	seedCmd.PersistentFlags().StringVar(&seedOptions.Host, "host", "", "Host of your Meilisearch database")
	seedCmd.PersistentFlags().StringVar(&seedOptions.ApiKey, "api-key", "", "API Key for accessing Meilisearch")
	rootCmd.AddCommand(seedCmd)
}
