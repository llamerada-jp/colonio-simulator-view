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
	"sort"
	"time"

	"github.com/llamerada-jp/simulator-view/pkg/utils"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	AuthStatusNone    = 0
	AuthStatusSuccess = 1
	AuthStatusFailure = 2

	LinkStatusOffline    = 0
	LinkStatusConnecting = 1
	LinkStatusOnline     = 2
	LinkStatusClosing    = 3
)

const (
	timeout                  = 4 * time.Second
	messageCurrentPosition   = "current position"
	messageLinks             = "links"
	messageRouting1DRequired = "routing 1d required"
	messageRouting2DRequired = "routing 2d required"
	messageLinkStatus        = "link status"
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
	follow   bool
	tail     bool
}

// Node contains last information for each time
type Node struct {
	enable         bool
	group          int
	nid            string
	x              float64
	y              float64
	links          []string
	required1D     []string
	required2D     []string
	timestamp      time.Time
	seedLinkStatus int
	nodeLinkStatus int
	authStatus     int
	isOnlyone      bool
}

// ParameterCurrentPosition is for decoding parameter of `current position` log
type ParameterCurrentPosition struct {
	Coordinate struct {
		X float64 `bson:"x"`
		Y float64 `bson:"y"`
	} `bson:"coordinate"`
}

// ParameterLinks is for decodeing parameter of `links` log
type ParameterLinks struct {
	Nids []string `bson:"nids"`
}

// ParameterRouting2DRequired is for decodeing parameter of `routing 2d required` log
type ParameterRouting2DRequired struct {
	Nids map[string]struct {
		X float64 `bson:"x"`
		Y float64 `bson:"y"`
	}
}

// ParameterLinkStatus is for decodeing parameter of `link status` log
type ParameterLinkStatus struct {
	Seed    int  `bson:"seed"`
	Node    int  `bson:"node"`
	Auth    int  `bson:"auth"`
	Onlyone bool `bson:"onlyone"`
}

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

// NewInstance makes a new instance of Sphere
func NewInstance(accessor *utils.Accessor, drawer Drawer, gl *utils.GL, follow, tail bool) *Model2D {
	return &Model2D{
		accessor: accessor,
		drawer:   drawer,
		nodes:    make(map[string]*Node),
		gl:       gl,
		follow:   follow,
		tail:     tail,
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

	// tail option
	if s.tail {
		current, err = s.accessor.GetLastTime()
		if err != nil {
			return err
		}
		*current = current.Add(-10 * time.Second)
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

	if s.follow {
		s.gl.SetImageDigit(6)
	} else {
		s.gl.SetImageDigit(int(math.Log10(last.Sub(*current).Seconds()) + 1.0))
	}

	// main loop until closing the window or existing data
	for s.gl.Loop() {
		*current = current.Add(time.Second)

		if s.follow {
			if current.UnixNano() > time.Now().Add(-5*time.Second).UnixNano() {
				time.Sleep(1 * time.Second)
			}

		} else {
			if current.UnixNano() > last.UnixNano() {
				break
			}
		}

		// update data
		if err = s.updateByLogs(current); err != nil {
			return err
		}
		s.disableTimeoutNode(current)
		s.setGroupNumber()

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

		case messageLinkStatus:
			var p ParameterLinkStatus
			if err = bson.Unmarshal(record.Param, &p); err != nil {
				return err
			}
			node := s.getNode(&record)
			node.seedLinkStatus = p.Seed
			node.nodeLinkStatus = p.Node
			node.authStatus = p.Auth
			node.isOnlyone = p.Onlyone
		}
	}

	return nil
}

func (s *Model2D) disableTimeoutNode(current *time.Time) {
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
}

type group struct {
	count  int
	assign int
}

type groups []*group

func (g groups) Len() int {
	return len(g)
}

func (g groups) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

func (g groups) Less(i, j int) bool {
	return g[i].count > g[j].count
}

func (s *Model2D) setGroupNumber() {
	for _, node := range s.nodes {
		node.group = 0
	}

	// assign temp ID
	groupMap := make(map[int]*group)
	for _, node := range s.nodes {
		if node.group == 0 {
			idx := len(groupMap) + 1
			groupMap[idx] = &group{
				count: s.findGroup(idx, node),
			}
		}
	}

	// order by member count
	groupOrder := make(groups, 0)
	for _, v := range groupMap {
		groupOrder = append(groupOrder, v)
	}
	sort.Sort(groupOrder)
	for idx, v := range groupOrder {
		v.assign = idx + 1
	}

	// assign ID 0 for very small group
	for _, g := range groupOrder {
		if g.count < 3 {
			g.assign = 0
		}
	}

	// assign new ID
	for _, node := range s.nodes {
		node.group = groupMap[node.group].assign
	}
}

func (s *Model2D) findGroup(group int, node *Node) int {
	node.group = group
	count := 1
	for _, linkNid := range node.links {
		if link, ok := s.nodes[linkNid]; ok {
			if link.enable && link.group == 0 {
				count += s.findGroup(group, link)
			}
		}
	}
	return count
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
