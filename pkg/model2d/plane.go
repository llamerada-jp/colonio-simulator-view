package model2d

import (
	"time"

	"github.com/llamerada-jp/simulator-view/pkg/utils"
)

type Plane struct{}

var colorMap = [][]float32{
	{0.8, 0.0, 0.8},
	{0.0, 0.2, 1.0},
	{0.0, 0.8, 0.2},
	{1.0, 0.6, 0.0},
}

func (s *Plane) draw(gl *utils.GL, nodes map[string]*Node, current *time.Time) error {
	for _, node := range nodes {
		if !node.enable {
			continue
		}
		colorIdx := node.group
		if colorIdx >= len(colorMap) {
			colorIdx = 0
		}
		gl.SetRGB(colorMap[colorIdx][0], colorMap[colorIdx][1], colorMap[colorIdx][2])
		gl.Point3(node.x, node.y, -1.0)

		for _, link := range node.links {
			if pair, ok := nodes[link]; ok {
				z := 0.0
				if pair.hasLink(node.nid) {
					if node.hasRequired2D(pair.nid) {
						gl.SetRGB(colorMap[colorIdx][0], colorMap[colorIdx][1], colorMap[colorIdx][2])
					} else {
						gl.SetRGB(0.8, 0.8, 0.8)
						z = 1.0
					}
				} else {
					gl.SetRGB(0.8, 0.0, 0.0)
				}
				gl.Line3(node.x, node.y, z, pair.x, pair.y, 0)
			}
		}
	}

	return nil
}
