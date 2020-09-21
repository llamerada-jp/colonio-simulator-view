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

	"github.com/llamerada-jp/simulator-view/pkg/model2d"
	"github.com/llamerada-jp/simulator-view/pkg/utils"
	"github.com/spf13/cobra"
)

var planeCmd = &cobra.Command{
	Use:   "plane",
	Short: "View data for plane",
	Run: func(cmd *cobra.Command, args []string) {
		// make accessor
		accessor, err := utils.NewAccessor(mongoURI, mongoDataBase, mongoCollection)
		if err != nil {
			fmt.Fprintf(os.Stderr, "accessor:%v", err)
		}
		defer accessor.Disconnect()

		// make drawer
		drawer := &model2d.Plane{}

		model := model2d.NewInstance(accessor, drawer, utils.NewGL(imageName), follow)
		err = model.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "plane:%v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(planeCmd)
}
