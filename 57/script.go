// +build js

package main

import (
	"errors"
	"fmt"

	"honnef.co/go/js/dom"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/webgl"
)

var gl *webgl.Context

const (
	vertexSource = `
attribute vec3 aVertexPosition;

uniform mat4 uMVMatrix;
uniform mat4 uPMatrix;

void main(void) {
	gl_Position = uPMatrix * uMVMatrix * vec4(aVertexPosition, 1.0);
}
`
	fragmentSource = `
precision mediump float;

void main(void) {
	gl_FragColor = vec4(1.0, 1.0, 1.0, 1.0);
}
`
)

var program js.Object

var vertexPositionAttribute int
var pMatrixUniform js.Object
var mvMatrixUniform js.Object

var mvMatrix mgl32.Mat4
var pMatrix mgl32.Mat4

var triangleVertexPositionBuffer js.Object
var itemSize int
var numItems int

func initShaders() error {
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(vertexShader, vertexSource)
	gl.CompileShader(vertexShader)
	defer gl.DeleteShader(vertexShader)

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(fragmentShader, fragmentSource)
	gl.CompileShader(fragmentShader)
	defer gl.DeleteShader(fragmentShader)

	program = gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	if !gl.GetProgramParameterb(program, gl.LINK_STATUS) {
		return errors.New("LINK_STATUS")
	}

	gl.ValidateProgram(program)
	if !gl.GetProgramParameterb(program, gl.VALIDATE_STATUS) {
		return errors.New("VALIDATE_STATUS")
	}

	gl.UseProgram(program)

	vertexPositionAttribute = gl.GetAttribLocation(program, "aVertexPosition")
	gl.EnableVertexAttribArray(vertexPositionAttribute)

	pMatrixUniform = gl.GetUniformLocation(program, "uPMatrix")
	mvMatrixUniform = gl.GetUniformLocation(program, "uMVMatrix")

	return nil
}

func createVbo() {
	triangleVertexPositionBuffer = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, triangleVertexPositionBuffer)
	vertices := []float32{
		0, 0, 0,
		300, 100, 0,
		0, 100, 0,
	}
	gl.BufferData(gl.ARRAY_BUFFER, vertices, gl.STATIC_DRAW)
	itemSize = 3
	numItems = 3
}

const viewportWidth = 400
const viewportHeight = 400

func main() {
	var document = dom.GetWindow().Document().(dom.HTMLDocument)
	canvas := document.CreateElement("canvas").(*dom.HTMLCanvasElement)
	devicePixelRatio := js.Global.Get("devicePixelRatio").Float()
	canvas.Width = int(viewportWidth*devicePixelRatio + 0.5)   // Nearest int.
	canvas.Height = int(viewportHeight*devicePixelRatio + 0.5) // Nearest int.
	canvas.Style().SetProperty("width", fmt.Sprintf("%vpx", viewportWidth), "")
	canvas.Style().SetProperty("height", fmt.Sprintf("%vpx", viewportHeight), "")
	document.Body().AppendChild(canvas)
	text := document.CreateElement("div")
	textContent := fmt.Sprintf("%v %v (%v) @%v", dom.GetWindow().InnerWidth(), canvas.Width, viewportWidth*devicePixelRatio, devicePixelRatio)
	text.SetTextContent(textContent)
	document.Body().AppendChild(text)

	document.Body().Style().SetProperty("margin", "0px", "")

	attrs := webgl.DefaultAttributes()
	attrs.Alpha = false
	attrs.Antialias = false

	var err error
	gl, err = webgl.NewContext(canvas.Underlying(), attrs)
	if err != nil {
		js.Global.Call("alert", "Error: "+err.Error())
	}

	err = initShaders()
	if err != nil {
		panic(err)
	}
	createVbo()

	gl.ClearColor(0.8, 0.3, 0.01, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// Draw scene.
	{
		gl.Viewport(0, 0, viewportWidth, viewportHeight)

		pMatrix = mgl32.Ortho2D(0, float32(viewportWidth), float32(viewportHeight), 0)

		mvMatrix = mgl32.Translate3D(50, 100, 0)

		gl.BindBuffer(gl.ARRAY_BUFFER, triangleVertexPositionBuffer)
		gl.VertexAttribPointer(vertexPositionAttribute, itemSize, gl.FLOAT, false, 0, 0)
		gl.UniformMatrix4fv(pMatrixUniform, false, pMatrix[:])
		gl.UniformMatrix4fv(mvMatrixUniform, false, mvMatrix[:])
		gl.DrawArrays(gl.TRIANGLES, 0, numItems)
	}
}
