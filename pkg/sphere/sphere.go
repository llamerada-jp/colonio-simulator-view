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
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/llamerada-jp/simulator-view/pkg/accessor"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	width                  = 800
	height                 = 600
	messageCurrentPosition = "current position"
)

// Sphere is the instance for sphere module
type Sphere struct {
	acc   *accessor.Accessor
	nodes map[string]*Node
}

// Node containes last information for each time
type Node struct {
	NID       string
	x         float64
	y         float64
	Timestamp time.Time
}

// ParameterCurrentPosition is for decoding parameter of `current position` log
type ParameterCurrentPosition struct {
	Coordinate struct {
		X float64 `bson:"x"`
		Y float64 `bson:"y"`
	} `bson:"coordinate"`
}

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

// NewInstance makes a new instance of Sphere
func NewInstance(acc *accessor.Accessor) *Sphere {
	return &Sphere{
		acc:   acc,
		nodes: make(map[string]*Node),
	}
}

// Run is an entory point for sphere module
func (s *Sphere) Run() error {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	window, err := glfw.CreateWindow(width, height, "simulator-view", nil, nil)
	if err != nil {
		log.Fatalln("failed to CreateWindow:", err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatalln("failed to initialize gl:", err)
	}

	current, err := s.acc.GetEarliestTime()
	if err != nil {
		return err
	}
	if current == nil {
		log.Fatalln("nothing data")
	}

	last, err := s.acc.GetLastTime()
	if err != nil {
		return err
	}

	if err = s.updateByLogs(current); err != nil {
		return err
	}

	s.setupScene()
	for !window.ShouldClose() {
		*current = current.Add(time.Second)
		if current.UnixNano() > last.UnixNano() {
			break
		}

		gl.Clear(gl.COLOR_BUFFER_BIT)
		if err = s.updateByLogs(current); err != nil {
			return err
		}
		//drawScene()
		window.SwapBuffers()
		glfw.PollEvents()
	}

	return nil
}

func (s *Sphere) setupScene() {
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.LIGHTING)

	gl.ClearColor(0.5, 0.5, 0.5, 0.0)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)

	ambient := []float32{0.5, 0.5, 0.5, 1}
	diffuse := []float32{1, 1, 1, 1}
	lightPosition := []float32{-5, 5, 10, 0}
	gl.Lightfv(gl.LIGHT0, gl.AMBIENT, &ambient[0])
	gl.Lightfv(gl.LIGHT0, gl.DIFFUSE, &diffuse[0])
	gl.Lightfv(gl.LIGHT0, gl.POSITION, &lightPosition[0])
	gl.Enable(gl.LIGHT0)

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	f := ((float64(width) / height) - 1) / 2
	gl.Frustum(-1-f, 1+f, -1, 1, 1.0, 10.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

func (s *Sphere) updateByLogs(current *time.Time) error {
	records, err := s.acc.GetByTime(current)
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
		}
	}
	return nil
}

func (s *Sphere) getNode(record *accessor.Record) *Node {
	nid := record.NID
	if _, ok := s.nodes[nid]; !ok {
		s.nodes[nid] = &Node{
			NID: nid,
		}
	}
	node := s.nodes[nid]
	node.Timestamp = record.TimeNtv
	return node
}
