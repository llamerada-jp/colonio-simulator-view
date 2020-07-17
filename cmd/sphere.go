package cmd

import (
	"context"
	"time"

	"github.com/llamerada-jp/simulator-view/pkg/accessor"
	"github.com/spf13/cobra"
)

var sphereCmd = &cobra.Command{
	Use:   "sphere",
	Short: "View data for sphere",
	RunE: func(cmd *cobra.Command, args []string) error {
		// make context
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// make accessor
		acc, err := accessor.NewAccessor(ctx, mongoURI, mongoDataBase, mongoCollection)
		if err != nil {
			return err
		}
		defer acc.Disconnect()

		// make sphere instance

		return nil
	},
}

func init() {
	sphereCmd.PersistentFlags().StringVarP(&mongoCollection, "collection", "c", "sphere", "collection name of mongoDB to get source data")
	rootCmd.AddCommand(sphereCmd)
}
