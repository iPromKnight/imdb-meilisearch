package cmd

import (
	cmd "github.com/ipromknight/imdb-meilisearch/cmd/commands"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "imdb-meilisearch",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")

	cmd.RegisterSeedCommand(rootCmd)
	cmd.RegisterSearchCommand(rootCmd)
	cmd.RegisterDaemonCommand(rootCmd)
}
