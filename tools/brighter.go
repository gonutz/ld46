//+build ignore

package main

import "github.com/gonutz/img"

var (
	blue0 = rgb(255, 255, 255)
	blue1 = rgb(200, 240, 255)
	blue2 = rgb(140, 220, 255)
	blue3 = rgb(65, 200, 255)
	blue4 = rgb(0, 175, 255)
	blue5 = rgb(0, 155, 235)
	blue6 = rgb(0, 140, 210)
	blue7 = rgb(0, 120, 180)
	blue8 = rgb(0, 105, 155)
	blues = []RGBA{
		blue0, blue1, blue2, blue3, blue4, blue5, blue6, blue7, blue8,
	}
)

type RGBA struct {
	r, g, b, a uint8
}

func rgb(r, g, b uint8) RGBA {
	return RGBA{r, g, b, 255}
}

func rgba(r, g, b, a uint8) RGBA {
	return RGBA{r, g, b, a}
}

func main() {
	img.Run(func(p *img.Pixel) {
		c := rgba(p.R, p.G, p.B, p.A)
		for i := 1; i < len(blues); i++ {
			if c == blues[i] {
				c = blues[i-1]
			}
		}
		p.R, p.G, p.B, p.A = c.r, c.g, c.b, c.a
	})
}
