package layout

import (
	"godtop/consoleui/config"
	"godtop/consoleui/widgets"
	"log"
	"sort"
	"strings"

	ui "github.com/gizak/termui/v3"
)

type widgetRule struct {
	Widget string
	Weight float64
	Height int
}

type layout struct {
	Rows [][]widgetRule
}

type Grid struct {
	*ui.Grid
	Volumes *widgets.VolumesWidget
}

var widgetNames []string = []string{"volumes", "network", "cpu"}

func GenerateGrid(wl layout, c config.Config) (*Grid, error) {
	rowDefs := wl.Rows
	uiRows := make([][]interface{}, 0)
	numRows := countNumRows(wl.Rows)

	var uiRow []interface{}
	maxHeight := 0
	heights := make([]int, 0)
	var h int
	for len(rowDefs) > 0 {
		h, uiRow, rowDefs = processRow(c, numRows, rowDefs)
		maxHeight += h
		uiRows = append(uiRows, uiRow)
		heights = append(heights, h)
	}

	rgs := make([]interface{}, 0)
	for i, ur := range uiRows {
		rh := float64(heights[i]) / float64(maxHeight)
		rgs = append(rgs, ui.NewRow(rh, ur...))
	}
	grid := &Grid{ui.NewGrid(), nil}
	grid.Set(rgs...)

	return grid, nil
}

func countNumRows(rs [][]widgetRule) int {
	var ttl int
	for len(rs) > 0 {
		ttl++
		line := rs[0]
		h := 1
		for _, c := range line {
			if c.Height > h {
				h = c.Height
			}
		}
		if h < len(rs) {
			rs = rs[h:]
		} else {
			break
		}
	}
	return ttl
}

// processRow eats a single row from the input list of rows and returns a UI
// row (GridItem) representation of the specification, along with a slice
// without that row.
//
// It does more than that, actually, because it may consume more than one row
// if there's a row span widget in the row; in this case, it'll consume as many
// rows as the largest row span object in the row, and produce an uber-row
// containing all that stuff. It returns a slice without the consumed elements.
func processRow(c config.Config, numRows int, rowDefs [][]widgetRule) (int, []interface{}, [][]widgetRule) {
	// Recursive function #3.  See the comment in deepFindProc.
	if len(rowDefs) < 1 {
		return 0, nil, [][]widgetRule{}
	}
	// The height of the tallest widget in this row; the number of rows that
	// will be consumed, and the overall height of the row that will be
	// produced.
	maxHeight := countMaxHeight([][]widgetRule{rowDefs[0]})
	var processing [][]widgetRule
	if maxHeight < len(rowDefs) {
		processing = rowDefs[0:maxHeight]
		rowDefs = rowDefs[maxHeight:]
	} else {
		processing = rowDefs[0:]
		rowDefs = [][]widgetRule{}
	}
	var colWeights []float64
	var columns [][]interface{}
	numCols := len(processing[0])
	if numCols < 1 {
		numCols = 1
	}
	for _, rd := range processing[0] {
		colWeights = append(colWeights, rd.Weight)
		columns = append(columns, make([]interface{}, 0))
	}
	colHeights := make([]int, numCols)
outer:
	for i, row := range processing {
		// A definition may fill up the columns before all rows are consumed,
		// e.g. cpu/2 net/2.  This block checks for that and, if it occurs,
		// prepends the remaining rows to the "remainder" return value.
		full := true
		for _, ch := range colHeights {
			if ch <= maxHeight {
				full = false
				break
			}
		}
		if full {
			rowDefs = append(processing[i:], rowDefs...)
			break
		}
		// Not all rows have been consumed, so go ahead and place the row's
		// widgets in columns
		for w, widg := range row {
			placed := false
			for k := w; k < len(colHeights); k++ { // there are enough columns
				ch := colHeights[k]
				if ch+widg.Height <= maxHeight {
					widget := makeWidget(c, widg)
					columns[k] = append(columns[k], ui.NewRow(float64(widg.Height)/float64(maxHeight), widget))
					colHeights[k] += widg.Height
					placed = true
					break
				}
			}
			// If all columns are full, break out, return the row, and continue processing
			if !placed {
				rowDefs = append(processing[i:], rowDefs...)
				break outer
			}
		}
	}
	var uiColumns []interface{}
	for i, widgets := range columns {
		if len(widgets) > 0 {
			uiColumns = append(uiColumns, ui.NewCol(float64(colWeights[i]), widgets...))
		}
	}

	return maxHeight, uiColumns, rowDefs
}

// Counts the height of the window so rows can be proportionally scaled.
func countMaxHeight(rs [][]widgetRule) int {
	var ttl int
	for len(rs) > 0 {
		line := rs[0]
		h := 1
		for _, c := range line {
			if c.Height > h {
				h = c.Height
			}
		}
		ttl += h
		if h < len(rs) {
			rs = rs[h:]
		} else {
			break
		}
	}
	return ttl
}

func makeWidget(c config.Config, widRule widgetRule) interface{} {
	switch widRule.Widget {
	case "volumes":
		return widgets.NewVolumesWidget()
	case "network":
		return widgets.NewNetworkWidget()
	case "cpu":
		cpu := widgets.NewCpuWidget()
		assignColors(cpu.Data, c.Colorscheme.CpuLines, cpu.LineColors)
		return cpu
	default:
		log.Printf("The widget %v doesn't exist (%v).", widRule.Widget, strings.Join(widgetNames, ","))
		return ui.NewBlock()
	}
}

func assignColors(data map[string][]float64, colors []int, assign map[string]ui.Color) {
	// Make sure the data is always processed in the same order so that
	// colors are assigned to devices consistently
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	i := 0 // For looping around if we run out of colors
	for _, v := range keys {
		if i >= len(colors) {
			i = 0
		}
		assign[v] = ui.Color(colors[i])
		i++
	}
}
