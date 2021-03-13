package layout

import (
	"bufio"
	"io"
	"log"
	"strconv"
	"strings"
)

/*************************

The syntax for the layout grid is: (rowspan:)?widget(/weight)?

1. Each line is a row.
2. Empty lines are skipped.
3. Spaces are compressed.
4. Legal widget names are: volumes, network.
5. Names are not case sensitive.
6. The simplest row is a single widget: volumes.
7. Widgets with no weights have a weight of 1.
8. If multiple widgets are put on a row with no weights, they will all have the same width.
9. Weights are integers.
10. A widget will have a width proportional to its weight devided by the total weight count of the row:
	volumes network			-- row
	disk/2 	mem/4			-- row
	...
	The first row will have two widgets, each will be 50% of the total width wide.
	The second row will have two widgets: disk and memory; the first will be 2/6 ~= 33% wide,
	and the second will be 5/6 ~= 83% wide (or, memory will be twice as wide as disk).
11. If the prefix is a number and colon, the widget will span that number of rows downward.
	2:volumes
	mem
	The volumes widget will be twice as high as the memory widget.
	````
	mem 	2:volumes
	network
	The memory and network will be in the same row as volumes, one over the other,
	and each half as hight as volumes
12. Negative, 0, or non-integer weights will be recorded as 1. Same for row spans.
13. Unrecognized widgets will cause the application to abort.
14. Widgets are filled in top down, left-to-right order.
15. The largers row span in a row defines the top-level row span; all smaller row spans
	constitude sub-rows in the row. For example:
	````
	cpu mem/3 net/5
	````
	Means that net/5 will be 5 rows tall overall, and mem will compose 3 of them.
	If following rows do not have enough widgets to fill the gaps, spacers will be used.
16. Lines beginning with '#' will be ignored. It must be the first character of the line.

*************************/

func ParseLayout(reader io.Reader) layout {
	scanner := bufio.NewScanner(reader)
	layout := layout{Rows: make([][]widgetRule, 0)}

	for scanner.Scan() { // go through lines

		// ignore empty and comments
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}

		// split input row to widget paramss
		layoutRow := make([]widgetRule, 0)
		lineWidgets := strings.Fields(line)

		weightTotal := 0
		for _, widgetParam := range lineWidgets {
			wRule := widgetRule{Weight: 1}

			// get widget's weight, rowspan and name
			var widget string
			splitRules := strings.Split(widgetParam, "/")
			splitRowspan := strings.Split(splitRules[0], ":")
			// validate rowspan and widget name part
			if len(splitRowspan) > 1 {
				wRowspan, err := strconv.Atoi(splitRowspan[0])
				if err != nil {
					log.Printf("layout.error.format (INT:)?STRING: %v (%v)", splitRowspan[0], widgetParam)
					wRowspan = 1
				}
				if wRowspan < 1 {
					wRowspan = 1
				}
				wRule.Height = wRowspan
				widget = splitRowspan[1]
			} else {
				wRule.Height = 1
				widget = splitRowspan[0]
			}
			wRule.Widget = strings.ToLower(widget)

			// validate widget's weight
			if len(splitRules) > 1 {
				wWeight, err := strconv.Atoi(splitRules[1])
				if err != nil {
					log.Printf("layout.error.format STRING(/INT)?: %v (%v)", splitRules[1], widgetParam)
					wWeight = 1
				}
				if wWeight < 1 {
					wWeight = 1
				}
				wRule.Weight = float64(wWeight)
				if len(splitRules) > 2 {
					log.Printf("layout.error.slashes: %v", widgetParam)
				}
				weightTotal += wWeight
			} else {
				weightTotal++
			}

			layoutRow = append(layoutRow, wRule)
		}

		// prevent tricksy users from breaking their own computers
		if weightTotal < 1 {
			weightTotal = 1
		}

		// calculate weight of each row
		for i, w := range layoutRow {
			layoutRow[i].Weight = w.Weight / float64(weightTotal)
		}

		layout.Rows = append(layout.Rows, layoutRow)
	}
	return layout
}
