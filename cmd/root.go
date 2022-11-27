package cmd

import (
	"os"

	"github.com/ant31/bungen/generators/model"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "bungen",
	Short: "Bungen is model generator for Bun package [Postgres Driver]",
	Long: `This application is a tool to generate the needed files
to quickly create a models for Bun [Postgres driver] https://github.com/uptrace/bun`,
	Version: "0.1.0",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			panic("help not found")
		}
	},
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

func init() {
	root.AddCommand(
		model.CreateCommand(),
	)
}

// Execute runs root cmd
func Execute() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
