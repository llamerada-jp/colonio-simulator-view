package model2d

import (
	"log"
	"math"
	"time"

	"github.com/llamerada-jp/simulator-view/pkg/utils"
)

type Sphere struct{}

var sphereScale = 0.95
var sphereColorMap = [][]float32{
	{0.8, 0.0, 0.8},
	{0.0, 0.2, 1.0},
	{0.0, 0.8, 0.2},
	{1.0, 0.6, 0.0},
}

func (s *Sphere) draw(gl *utils.GL, nodes map[string]*Node, current *time.Time) error {
	nodeCount := 0
	seedCount := 0
	onlyoneCount := 0

	for _, node := range nodes {
		if !node.enable {
			continue
		}
		nodeCount++

		colorIdx := node.group
		if colorIdx >= len(colorMap) {
			colorIdx = 0
		}
		x, y, z := s.convertCoordinate(node.x, node.y)
		gl.SetRGB(s.reduceColorByZ(colorMap[colorIdx], z))
		gl.Point3(x, y, z)

		if node.seedLinkStatus == LinkStatusOnline {
			x, y, z := s.convertCoordinate(node.x, node.y)
			gl.SetRGB(s.reduceColorByZ([]float32{1.0, 0.0, 0.0}, z))
			gl.Box3(x, y, z, 6.0)
			seedCount++
		}
		if node.isOnlyone {
			x, y, z := s.convertCoordinate(node.x, node.y)
			gl.SetRGB(s.reduceColorByZ([]float32{1.0, 0.0, 0.0}, z))
			gl.Box3(x, y, z, 10.0)
			onlyoneCount++
		}

		for _, link := range node.links {
			if pair, ok := nodes[link]; ok {
				var rgb []float32
				if pair.hasLink(node.nid) {
					if node.hasRequired2D(pair.nid) {
						rgb = colorMap[colorIdx]
					} else {
						rgb = []float32{0.8, 0.8, 0.8}
					}
				} else {
					rgb = []float32{0.8, 0.0, 0.0}
				}
				x1, y1, z1 := s.convertCoordinate(node.x, node.y)
				x2, y2, z2 := s.convertCoordinate(pair.x, pair.y)
				gl.SetRGB(s.reduceColorByZ(rgb, (z1+z2)/2.0))
				gl.Line3(x1, y1, z1, x2, y2, z2)
			}
		}
	}

	log.Printf("node: %d/%d  seed: %d/%d", nodeCount, len(nodes), onlyoneCount, seedCount)

	return nil
}

func (s *Sphere) reduceColorByZ(ci []float32, z float64) (r, g, b float32) {
	rate := (float32(-z) + 1.0) / 2.0
	r = 1.0 - ((1.0 - ci[0]) * rate)
	g = 1.0 - ((1.0 - ci[1]) * rate)
	b = 1.0 - ((1.0 - ci[2]) * rate)
	return
}

func (s *Sphere) convertCoordinate(xi, yi float64) (xo, yo, zo float64) {
	xo = math.Cos(xi) * math.Cos(yi)
	yo = math.Sin(yi)
	zo = math.Sin(xi) * math.Cos(yi)
	return
}
