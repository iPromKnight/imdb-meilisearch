package cmd

import (
	"fmt"
	daemonApi "github.com/ipromknight/imdb-meilisearch/internal/api"
	meilisearchConfiguration "github.com/ipromknight/imdb-meilisearch/internal/meilisearch-configuration"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
)

type daemonCommandConfig struct {
	meilisearchConfiguration.ClientOptions
}

var daemonOptions daemonCommandConfig

func RegisterDaemonCommand(rootCmd *cobra.Command) {
	daemonCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Run the api on port 8080.",
		Long:  `This will run up a rest api on port 8080 to perform searches, and on demand ingestion.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			daemonOptions.ClientOptions = daemonOptions.ClientOptions.PopulateFromEnv()

			if daemonOptions.Host == "" {
				return fmt.Errorf("required flag 'host' is not set and the fallback environment variable 'MEILISEARCH_HOST' is not set")
			}
			if daemonOptions.ApiKey == "" {
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

			err := daemonApi.ServeApi(daemonOptions.ClientOptions, logger)
			if err != nil {
				logger.Fatal().AnErr("error", err).Msg("Failed to start api")
			}
		},
	}

	daemonCmd.PersistentFlags().StringVar(&daemonOptions.Host, "host", "", "Host of your Meilisearch database")
	daemonCmd.PersistentFlags().StringVar(&daemonOptions.ApiKey, "api-key", "", "API Key for accessing Meilisearch")
	rootCmd.AddCommand(daemonCmd)
}
