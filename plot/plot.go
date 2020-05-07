package plot

import (
	"image/color"
	"math/rand"

	zl "github.com/rs/zerolog/log"

	"github.com/KudinovKV/golang-aco/aco"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func Draw(c *aco.Cluster) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Ant Colony Optimization Algorithm"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"
	pts := make(plotter.XYs, len(c.BestPath))

	for i := 0; i != len(c.BestPath); i++ {
		n := 0
		flag := false
		for j := 0; j != len(c.City); j++ {
			for k := 0; k != len(c.City[j]); k++ {
				if c.BestPath[i] == n {
					pts[i].X = float64(c.City[j][k].X)
					pts[i].Y = float64(c.City[j][k].Y)
					flag = true
					break
				}
				n++
			}
			if flag == true {
				break
			}
		}
	}
	for j := 0; j != len(c.City); j++ {
		clusterPts := make(plotter.XYs, len(c.City[j]))
		for k := 0; k != len(c.City[j]); k++ {
			clusterPts[k].X = float64(c.City[j][k].X)
			clusterPts[k].Y = float64(c.City[j][k].Y)
		}
		// Make a scatter plotter and set its style.
		s, err := plotter.NewScatter(clusterPts)
		if err != nil {
			panic(err)
		}
		s.GlyphStyle.Color = color.RGBA{
			R: uint8(rand.Intn(255)),
			G: uint8(rand.Intn(255)),
			B: uint8(rand.Intn(255)),
			A: 255}

		p.Add(s)
	}

	line, _, err := plotter.NewLinePoints(pts)
	if err != nil {
		zl.Fatal().Err(err).Msg("Can't create line")
	}
	line.Color = color.RGBA{R: 255, A: 255}
	p.Add(line)
	// Save the plot to a PNG file.
	if err := p.Save(16*vg.Inch, 16*vg.Inch, "points.png"); err != nil {
		zl.Fatal().Err(err).Msg("Can't save")
	}
}
