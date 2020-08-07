package model2d

import (
	"time"

	"github.com/llamerada-jp/simulator-view/pkg/utils"
)

type Plane struct{}

func (s *Plane) draw(gl *utils.GL, nodes map[string]*Node, current *time.Time) error {
	for _, node := range nodes {
		if !node.enable {
			continue
		}
		gl.SetRGB(0.0, 0.8, 0.2)
		gl.Point3(node.x, node.y, 0)

		for _, link := range node.links {
			if pair, ok := nodes[link]; ok {
				if pair.hasLink(node.nid) {
					if node.hasRequired2D(pair.nid) {
						gl.SetRGB(0.0, 1.0, 0.2)
					} else {
						gl.SetRGB(0.6, 0.6, 0.6)
					}
				} else {
					gl.SetRGB(0.8, 0.0, 0.0)
				}
				gl.Line3(node.x, node.y, 0, pair.x, pair.y, 0)
			}
		}
	}

	return nil
}
