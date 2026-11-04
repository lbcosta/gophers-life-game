package gopher

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ellipseSegments é o número de lados do polígono que aproxima uma elipse.
const ellipseSegments = 32

type point struct{ x, y float32 }

// ellipsePoints devolve os vértices do polígono que aproxima uma elipse de
// centro (cx, cy) e raios rx (horizontal) e ry (vertical), girada por angle
// graus (sentido horário na tela).
//
// É a única definição do contorno de uma elipse: tanto o desenho quanto o
// cálculo do alcance do Gopher usam estes pontos, então os dois sempre concordam.
func ellipsePoints(cx, cy, rx, ry float32, angle float64) [ellipseSegments]point {
	sin, cos := math.Sincos(radians(angle))
	sinA, cosA := float32(sin), float32(cos)

	var pts [ellipseSegments]point
	for i := range pts {
		sinT, cosT := math.Sincos(2 * math.Pi * float64(i) / ellipseSegments)
		// Ponto na elipse sem rotação, em relação ao centro.
		px := rx * float32(cosT)
		py := ry * float32(sinT)
		// Rotação e translação.
		pts[i] = point{
			x: cx + px*cosA - py*sinA,
			y: cy + px*sinA + py*cosA,
		}
	}
	return pts
}

// fillEllipse preenche, em dst, a elipse descrita em ellipsePoints.
func fillEllipse(dst *ebiten.Image, cx, cy, rx, ry float32, angle float64, clr color.Color) {
	pts := ellipsePoints(cx, cy, rx, ry, angle)

	var path vector.Path
	path.MoveTo(pts[0].x, pts[0].y)
	for _, p := range pts[1:] {
		path.LineTo(p.x, p.y)
	}
	path.Close()

	var opts vector.DrawPathOptions
	opts.AntiAlias = true
	opts.ColorScale.ScaleWithColor(clr)
	vector.FillPath(dst, &path, nil, &opts)
}
