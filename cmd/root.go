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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	follow          bool
	imageName       string
	detailLevel     uint
	mongoURI        string
	mongoDataBase   string
	mongoCollection string
)

var rootCmd = &cobra.Command{
	Use: "simulator-view",
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.BoolVarP(&follow, "follow", "f", false, "Specify if the logs should be streamed")
	flags.StringVarP(&imageName, "image-name", "i", "", "Image path and name pattern like hoge/foo@.png (@ will be replace by index like 001, 002...)")
	flags.UintVarP(&detailLevel, "detail-leval", "l", 0, "Whether to draw detailed information")
	flags.StringVarP(&mongoURI, "uri", "u", "mongodb://localhost:27017", "URI of mongoDB to get source data")
	flags.StringVarP(&mongoDataBase, "database", "d", "simulation", "database name of mongoDB to get source data")
	flags.StringVarP(&mongoCollection, "collection", "c", "logs", "collection name of mongoDB to get source data")
}

// Execute is entry point for all commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
