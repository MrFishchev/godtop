package termui

import (
	"godtop/consoleui/termui/drawille-go"
	"image"
	"sort"
	"strconv"
	"unicode"

	ui "github.com/gizak/termui/v3"
)

// LineGraph draws a graph like data points
type LineGraph struct {
	*ui.Block

	// Data is a size-managed data set for the graph. Each entry is a line
	// each sub-array are points in the line. The maximum size of the sub-arrays is controlled
	// by the size of the canvas. This array is **not** thread-safe. Do not modify this array,
	// or it's sub-arrays in threads different than the thread that calls Draw().
	Data map[string][]float64

	// The labels drawn on the graph for each of the lines. The key is shared by Data.
	// The value is the text that will be rendered.
	Labels map[string]string

	HorizontalScale  int
	LineColors       map[string]ui.Color
	LabelStyles      map[string]ui.Modifier
	DefaultLineColor ui.Color
	seriesList       numbered
}

func NewLineGraph() *LineGraph {
	return &LineGraph{
		Block:           ui.NewBlock(),
		Data:            make(map[string][]float64),
		Labels:          make(map[string]string),
		HorizontalScale: 5,
		LineColors:      make(map[string]ui.Color),
		LabelStyles:     make(map[string]ui.Modifier),
	}
}

func (g *LineGraph) Draw(buf *ui.Buffer) {
	g.Block.Draw(buf)

	// render each data point on to the canvas then copy over the braille to the buffer at the end
	// fyi braille characters have 2x4 dots for each character
	canvas := drawille.NewCanvas()

	// used to keep track of the braille colors until the end when we render the braille to the buffer
	colors := make([][]ui.Color, g.Inner.Dx()+2)
	for i := range colors {
		colors[i] = make([]ui.Color, g.Inner.Dy()+2)
	}

	if len(g.seriesList) != len(g.Data) {
		// sort the series so that overlapping data will overlap the same way each time
		g.seriesList = make(numbered, len(g.Data))
		i := 0
		for seriesName := range g.Data {
			g.seriesList[i] = seriesName
			i++
		}
		sort.Sort(g.seriesList)
	}

	// draw lines in reverse order so that the first color defined in the colorscheme is on top
	for i := len(g.seriesList) - 1; i >= 0; i-- {
		seriesName := g.seriesList[i]
		seriesData := g.Data[seriesName]
		seriesLineColor, ok := g.LineColors[seriesName]
		if !ok {
			seriesLineColor = g.DefaultLineColor
			g.LineColors[seriesName] = seriesLineColor
		}

		// coordinates of last point
		lastY, lastX := -1, -1
		// assign colors to `colors` and lines/points to the canvas
		dx := g.Inner.Dx()
		for i := len(seriesData) - 1; i >= 0; i-- {
			x := ((dx + 1) * 2) - 1 - (((len(seriesData) - 1) - i) * g.HorizontalScale)
			y := ((g.Inner.Dy() + 1) * 4) - 1 - int((float64((g.Inner.Dy())*4)-1)*(seriesData[i]/100))
			if x < 0 {
				// render the line to the last point up to the wall
				if x > -g.HorizontalScale {
					for _, p := range drawille.Line(lastX, lastY, x, y) {
						if p.X > 0 {
							canvas.Set(p.X, p.Y)
							colors[p.X/2][p.Y/4] = seriesLineColor
						}
					}
				}
				if len(seriesData) > 4*dx {
					g.Data[seriesName] = seriesData[dx-1:]
				}
				break
			}

			if lastY == -1 { // this is the first point
				canvas.Set(x, y)
				colors[x/2][y/4] = seriesLineColor
			} else {
				canvas.DrawLine(lastX, lastY, x, y)
				for _, p := range drawille.Line(lastX, lastY, x, y) {
					colors[p.X/2][p.Y/4] = seriesLineColor
				}
			}
			lastX, lastY = x, y
		}

		// copy braille and colors to buffer
		for y, line := range canvas.Rows(canvas.MinX(), canvas.MinY(), canvas.MaxX(), canvas.MaxY()) {
			for x, char := range line {
				x /= 3 // idk why but it works
				if x == 0 {
					continue
				}

				if char != 10240 { // empty braille character
					buf.SetCell(
						ui.NewCell(char, ui.NewStyle(colors[x][y])),
						image.Pt(g.Inner.Min.X+x-1, g.Inner.Min.Y+y-1),
					)
				}
			}
		}
	}

	// renders key/label ontop
	maxWid := 0
	xoff := 0 // X offset for additional columns of text
	yoff := 0 // Y offset for resetting column to top of widget
	for i, seriesName := range g.seriesList {
		if yoff+i+2 > g.Inner.Dy() {
			xoff += maxWid + 2
			yoff = -i
			maxWid = 0
		}
		seriesLineColor, ok := g.LineColors[seriesName]
		if !ok {
			seriesLineColor = g.DefaultLineColor
		}
		seriesLabelStyle, ok := g.LabelStyles[seriesName]
		if !ok {
			seriesLabelStyle = ui.ModifierClear
		}

		// render key ontop, but let braille be drawn over space characters
		str := seriesName + " " + g.Labels[seriesName]
		if len(str) > maxWid {
			maxWid = len(str)
		}
		for k, char := range str {
			if char != ' ' {
				buf.SetCell(
					ui.NewCell(char, ui.NewStyle(seriesLineColor, ui.ColorClear, seriesLabelStyle)),
					image.Pt(xoff+g.Inner.Min.X+2+k, yoff+g.Inner.Min.Y+i+1),
				)
			}
		}
	}
}

// A string containing an integer
type numbered []string

func (n numbered) Len() int      { return len(n) }
func (n numbered) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n numbered) Less(i, j int) bool {
	a := n[i]
	b := n[j]
	for i := 0; i < len(a); i++ {
		ac := a[i]
		if unicode.IsDigit(rune(ac)) {
			j := i + 1
			for ; j < len(a); j++ {
				if !unicode.IsDigit(rune(a[j])) {
					break
				}
				if j >= len(b) {
					return false
				}
				if !unicode.IsDigit(rune(b[j])) {
					return false
				}
			}
			an, err := strconv.Atoi(a[i:j])
			if err != nil {
				return true
			}
			if j > len(b) {
				return false
			}
			for ; j < len(b); j++ {
				if !unicode.IsDigit(rune(b[j])) {
					break
				}
			}
			bn, err := strconv.Atoi(b[i:j])
			if err != nil {
				return true
			}
			if an < bn {
				return true
			} else if bn < an {
				return false
			}
			i = j
		}
		if i >= len(a) {
			return true
		} else if i >= len(b) {
			return false
		}
		if ac < b[i] {
			return true
		} else if b[i] < ac {
			return false
		}
	}
	return true
}
