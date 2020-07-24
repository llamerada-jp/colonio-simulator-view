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

package utils

import (
	"log"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	width  = 1024
	height = 1024
)

// GL containing any instances of OpenGL
type GL struct {
	program uint32
	window  *glfw.Window
	colorR  float32
	colorG  float32
	colorB  float32
}

// NewGL makes new utility instance of OpenGL
func NewGL() *GL {
	return &GL{}
}

// Setup OpenGL and create a new window
func (g *GL) Setup() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "simulator-view", nil, nil)
	if err != nil {
		log.Fatalln("failed to CreateWindow:", err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	g.window = window

	if err := gl.Init(); err != nil {
		log.Fatalln("failed to initialize gl:", err)
	}

	g.setupProgram()

	gl.ClearColor(1.0, 1.0, 1.0, 1.0)
}

// Quit OpenGL
func (g *GL) Quit() {
	glfw.Terminate()
}

// Loop swap and clear buffer, and poll events. return false is program should quit
func (g *GL) Loop() bool {
	g.window.SwapBuffers()
	// time.Sleep(time.Second)

	// clear and draw
	defer func() {
		glfw.PollEvents()
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	}()

	return !g.window.ShouldClose()
}

// SetRGB set fill color
func (g *GL) SetRGB(red, green, blue float32) {
	g.colorR = red
	g.colorG = green
	g.colorB = blue
}

// Point3 draw point
func (g *GL) Point3(x, y, z float32) {
	vertices := []float32{
		x - 0.1, y - 0.1, z,
		x + 0.1, y - 0.1, z,
		x + 0.1, y + 0.1, z,
		x - 0.1, y - 0.1, z,
		x - 0.1, y + 0.1, z,
	}

	fragments := []float32{
		g.colorR, g.colorG, g.colorB,
		g.colorR, g.colorG, g.colorB,
		g.colorR, g.colorG, g.colorB,
		g.colorR, g.colorG, g.colorB,
		g.colorR, g.colorG, g.colorB,
	}

	var vertexArrayObject uint32
	gl.GenVertexArrays(1, &vertexArrayObject)
	defer gl.DeleteVertexArrays(1, &vertexArrayObject)
	gl.BindVertexArray(vertexArrayObject)

	var vertexBuffer uint32
	gl.GenBuffers(1, &vertexBuffer)
	defer gl.DeleteBuffers(1, &vertexBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STREAM_DRAW)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	var fragmentBuffer uint32
	gl.GenBuffers(1, &fragmentBuffer)
	defer gl.DeleteBuffers(1, &fragmentBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, fragmentBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(fragments)*4, gl.Ptr(fragments), gl.STREAM_DRAW)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(1)

	gl.DrawArrays(gl.TRIANGLES, 0, 3)
	gl.DrawArrays(gl.TRIANGLES, 2, 3)

	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

func (g *GL) setupProgram() {
	program := gl.CreateProgram()
	vertexShader := g.setupShader(`
#version 330 core

layout(location = 0) in vec3 vertexPosition_modelspace;
layout(location = 1) in vec3 vertexColor;

out vec3 fragmentColor;
uniform mat4 MVP;

void main(){
	//gl_Position =  MVP * vec4(vertexPosition_modelspace,1);
	gl_Position = vec4(vertexPosition_modelspace.x, vertexPosition_modelspace.y, vertexPosition_modelspace.z, 1.0);
	fragmentColor = vertexColor;
}
`+"\x00", gl.VERTEX_SHADER)
	defer gl.DeleteShader(vertexShader)

	fragmentShader := g.setupShader(`
#version 330 core

in vec3 fragmentColor;
out vec3 color;

void main(){
	color = fragmentColor;
	// color = vec3(1.0, 1.0, 1.0);
}
`+"\x00", gl.FRAGMENT_SHADER)
	defer gl.DeleteShader(fragmentShader)

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)

	gl.LinkProgram(program)
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		lotText := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(lotText))
		log.Fatalf("failed to compile at setupProgram %v", lotText)
	}

	g.program = program
	gl.UseProgram(g.program)
}

func (g *GL) setupShader(source string, shaderType uint32) uint32 {
	shader := gl.CreateShader(shaderType)
	sourceChars, freeFunc := gl.Strs(source)
	defer freeFunc()
	gl.ShaderSource(shader, 1, sourceChars, nil)
	gl.CompileShader(shader)
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		lotText := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(lotText))
		log.Fatalf("failed to compile at setupShader %v: %v", source, lotText)
	}

	return shader
}
