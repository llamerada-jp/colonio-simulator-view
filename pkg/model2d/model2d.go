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

package model2d

import (
	"log"
	"math"
	"runtime"
	"time"

	"github.com/llamerada-jp/simulator-view/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	timeout                  = 4 * time.Second
	messageCurrentPosition   = "current position"
	messageLinks             = "links"
	messageRouting1DRequired = "routing 1d required"
	messageRouting2DRequired = "routing 2d required"
)

type Drawer interface {
	draw(*utils.GL, map[string]*Node, *time.Time) error
}

// Model2D is the instance for sphere module
type Model2D struct {
	accessor *utils.Accessor
	drawer   Drawer
	nodes    map[string]*Node
	gl       *utils.GL
}

// Node containes last information for each time
type Node struct {
	enable     bool
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
func NewInstance(accessor *utils.Accessor, drawer Drawer, gl *utils.GL) *Model2D {
	return &Model2D{
		accessor: accessor,
		drawer:   drawer,
		nodes:    make(map[string]*Node),
		gl:       gl,
	}
}

// Run is an entory point for sphere module
func (s *Model2D) Run() error {
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
		if err = s.drawer.draw(s.gl, s.nodes, current); err != nil {
			return err
		}
	}

	return nil
}

func (s *Model2D) updateByLogs(current *time.Time) error {
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

	// disable timeout node
	for nid, node := range s.nodes {
		if node.timestamp.Add(timeout).After(*current) {
			node.enable = true
			continue
		}

		node.enable = false
		for _, nextNid := range node.links {
			if next, ok := s.nodes[nextNid]; ok {
				if next.timestamp.Add(timeout).After(*current) && next.hasLink(nid) {
					node.enable = true
					break
				}
			}
		}
	}

	return nil
}

func (s *Model2D) getNode(record *utils.Record) *Node {
	nid := record.NID
	if _, ok := s.nodes[nid]; !ok {
		s.nodes[nid] = &Node{
			enable: true,
			nid:    nid,
		}
	}
	node := s.nodes[nid]
	node.timestamp = record.TimeNtv
	return node
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
