package model2d

import (
	"math"
	"time"

	"github.com/llamerada-jp/simulator-view/pkg/utils"
)

type Sphere struct{}

func (s *Sphere) draw(gl *utils.GL, nodes map[string]*Node, current *time.Time) error {
	for _, node := range nodes {
		gl.SetRGB(0.0, 0.8, 0.2)
		gl.Point3(s.convertCoordinate(node.x, node.y))

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
				x1, y1, z1 := s.convertCoordinate(node.x, node.y)
				x2, y2, z2 := s.convertCoordinate(pair.x, pair.y)
				gl.Line3(x1, y1, z1, x2, y2, z2)
			}
		}
	}

	return nil
}

func (s *Sphere) convertCoordinate(xi, yi float64) (xo, yo, zo float64) {
	xo = math.Cos(xi) * math.Cos(yi)
	yo = math.Sin(yi)
	zo = math.Sin(xi) * math.Cos(yi)
	return
}
