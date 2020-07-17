package cmd

import "github.com/spf13/cobra"

var sphereCmd = &cobra.Command{
	Use:   "sphere",
	Short: "View data for sphere",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sphereCmd)
}
