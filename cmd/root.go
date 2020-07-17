package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	mongoURI string
)

var rootCmd = &cobra.Command{
	Use: "simulator-view",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&mongoURI, "uri", "u", "", "URI of mongoDB to get source data")
}

// Execute is entry point for all commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
