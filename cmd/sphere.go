/**
 * Copyright 2020-2020 Yuji Ito <llamerada.jp@gmail.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cmd

import (
	"github.com/llamerada-jp/simulator-view/pkg/accessor"
	"github.com/spf13/cobra"
)

var sphereCmd = &cobra.Command{
	Use:   "sphere",
	Short: "View data for sphere",
	RunE: func(cmd *cobra.Command, args []string) error {
		// make accessor
		acc, err := accessor.NewAccessor(mongoURI, mongoDataBase, mongoCollection)
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
