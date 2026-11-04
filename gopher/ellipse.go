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

// rotate gira p em torno da origem pelo ângulo de seno sin e cosseno cos
// (sentido horário na tela, a mesma rotação de ebiten.GeoM.Rotate).
func (p point) rotate(sin, cos float32) point {
	return point{x: p.x*cos - p.y*sin, y: p.x*sin + p.y*cos}
}

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
		// Ponto na elipse sem rotação, em relação ao centro; depois gira e translada.
		p := point{rx * float32(cosT), ry * float32(sinT)}.rotate(sinA, cosA)
		pts[i] = point{cx + p.x, cy + p.y}
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
