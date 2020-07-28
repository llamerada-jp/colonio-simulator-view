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

package sphere

import (
	"log"
	"math"
	"runtime"
	"time"

	"github.com/llamerada-jp/simulator-view/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	messageCurrentPosition   = "current position"
	messageLinks             = "links"
	messageRouting1DRequired = "routing 1d required"
	messageRouting2DRequired = "routing 2d required"
)

// Sphere is the instance for sphere module
type Sphere struct {
	accessor *utils.Accessor
	nodes    map[string]*Node
	gl       *utils.GL
}

// Node containes last information for each time
type Node struct {
	nid        string
	x          float64
	y          float64
	links      []string
	required1D []string
	required2D []string
	timestamp  time.Time
}

// ParameterCurrentPosition is for decoding parameter of `current position` log
type ParameterCurrentPosition struct {
	Coordinate struct {
		X float64 `bson:"x"`
		Y float64 `bson:"y"`
	} `bson:"coordinate"`
}

// ParameterLinks is for decodeing paramter of `links` log
type ParameterLinks struct {
	Nids []string `bson:"nids"`
}

// ParameterRouting2DRequired is for decodeing paramter of `routing 2d required` log
type ParameterRouting2DRequired struct {
	Nids map[string]struct {
		X float64 `bson:"x"`
		Y float64 `bson:"y"`
	}
}

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

// NewInstance makes a new instance of Sphere
func NewInstance(acc *utils.Accessor, gl *utils.GL) *Sphere {
	return &Sphere{
		accessor: acc,
		nodes:    make(map[string]*Node),
		gl:       gl,
	}
}

// Run is an entory point for sphere module
func (s *Sphere) Run() error {
	// get time range from mongodb
	current, err := s.accessor.GetEarliestTime()
	if err != nil {
		return err
	}
	if current == nil {
		log.Fatalln("nothing data")
	}

	last, err := s.accessor.GetLastTime()
	if err != nil {
		return err
	}

	if err = s.updateByLogs(current); err != nil {
		return err
	}

	// setup opengl
	s.gl.Setup()
	defer s.gl.Quit()

	s.gl.SetImageDigit(int(math.Log10(last.Sub(*current).Seconds()) + 1.0))

	// main loop until closing the window or existing data
	for s.gl.Loop() {
		*current = current.Add(time.Second)
		if current.UnixNano() > last.UnixNano() {
			break
		}

		// update data
		if err = s.updateByLogs(current); err != nil {
			return err
		}

		// draw data
		s.draw(current)
	}

	return nil
}

func (s *Sphere) updateByLogs(current *time.Time) error {
	records, err := s.accessor.GetByTime(current)
	if err != nil {
		return err
	}

	for _, record := range records {
		switch record.Message {
		case messageCurrentPosition:
			var p ParameterCurrentPosition
			if err = bson.Unmarshal(record.Param, &p); err != nil {
				return err
			}
			node := s.getNode(&record)
			node.x = p.Coordinate.X
			node.y = p.Coordinate.Y

		case messageLinks:
			var p ParameterLinks
			if err = bson.Unmarshal(record.Param, &p); err != nil {
				return err
			}
			node := s.getNode(&record)
			node.links = p.Nids

		case messageRouting1DRequired:
			var p ParameterLinks
			if err = bson.Unmarshal(record.Param, &p); err != nil {
				return err
			}
			node := s.getNode(&record)
			node.required1D = p.Nids

		case messageRouting2DRequired:
			var p ParameterRouting2DRequired
			if err = bson.Unmarshal(record.Param, &p); err != nil {
				return err
			}
			node := s.getNode(&record)
			node.required2D = make([]string, len(p.Nids))
			idx := 0
			for k := range p.Nids {
				node.required2D[idx] = k
				idx++
			}
		}
	}
	return nil
}

func (s *Sphere) getNode(record *utils.Record) *Node {
	nid := record.NID
	if _, ok := s.nodes[nid]; !ok {
		s.nodes[nid] = &Node{
			nid: nid,
		}
	}
	node := s.nodes[nid]
	node.timestamp = record.TimeNtv
	return node
}

func (s *Sphere) draw(current *time.Time) {
	for _, node := range s.nodes {
		s.gl.SetRGB(0.0, 0.8, 0.2)
		s.gl.Point3(s.convertCoordinate(node.x, node.y))

		for _, link := range node.links {
			if pair, ok := s.nodes[link]; ok {
				if pair.hasLink(node.nid) {
					if node.hasRequired2D(pair.nid) {
						s.gl.SetRGB(0.0, 1.0, 0.2)
					} else {
						s.gl.SetRGB(0.6, 0.6, 0.6)
					}
				} else {
					s.gl.SetRGB(0.8, 0.0, 0.0)
				}
				x1, y1, z1 := s.convertCoordinate(node.x, node.y)
				x2, y2, z2 := s.convertCoordinate(pair.x, pair.y)
				s.gl.Line3(x1, y1, z1, x2, y2, z2)
			}
		}
	}
}

func (s *Sphere) convertCoordinate(xi, yi float64) (xo, yo, zo float64) {
	xo = math.Cos(xi) * math.Cos(yi)
	yo = math.Sin(yi)
	zo = math.Sin(xi) * math.Cos(yi)
	return
}

func (n *Node) hasLink(nid string) bool {
	for _, v := range n.links {
		if v == nid {
			return true
		}
	}
	return false
}

func (n *Node) hasRequired2D(nid string) bool {
	for _, v := range n.required2D {
		if v == nid {
			return true
		}
	}
	return false
}
