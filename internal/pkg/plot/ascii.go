package plot

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func LineChartString(values []float64, width, height int) string {
	// three characters of the width go to chart borders and newline
	plotWidth := width - 3
	// two characters of the height go to chart borders
	plotHeight := height - 2

	maxAbsVal := math.Inf(-1)
	for _, v := range values {
		maxAbsVal = math.Max(maxAbsVal, math.Abs(v))
	}
	if maxAbsVal == 0 {
		maxAbsVal = 1
	}

	grid := make([][]rune, plotHeight)
	for i := 0; i < len(grid); i++ {
		grid[i] = make([]rune, width)
		for j := 0; j < len(grid[i]); j++ {
			grid[i][j] = ' '
		}
		grid[i][0] = '|'
		grid[i][len(grid[i])-2] = '|'
		grid[i][len(grid[i])-1] = '\n'
	}

	// by default, zero is always at the center
	zero := plotHeight / 2
	// draw a dashed line at zero
	for i := 1; i < len(grid[zero])-2; i++ {
		if i%2 == 1 {
			grid[zero][i] = '-'
		}
	}

	var lastRow int
	step := float64(len(values)-1) / float64(plotWidth)
	for i := 0; i < plotWidth; i++ {
		// interpolate the value at (step*i + step/2)
		t := step*float64(i) + step/2

		// find the two closest values
		i1 := int(math.Floor(t))
		i2 := int(math.Ceil(t))
		val := values[i1] + (values[i2]-values[i1])*(t-float64(i1))

		row := int(float64(plotHeight-1) * (maxAbsVal - val) / (2 * maxAbsVal))
		grid[row][i+1] = 'O'
		if i > 0 && row != lastRow {
			// draw a line from the last point to this point
			if row > lastRow {
				for j := lastRow + 1; j <= row; j++ {
					grid[j][i+1] = 'o'
				}
			} else {
				for j := row + 1; j <= lastRow; j++ {
					grid[j][i+1] = 'o'
				}
			}
		}
		lastRow = row
	}

	maxStrInt := fmt.Sprintf("%d", int(maxAbsVal))
	maxStrDec := strconv.FormatFloat(maxAbsVal-float64(int(maxAbsVal)), 'f', -1, 64)
	if strings.HasPrefix(maxStrDec, "0.") {
		// at most 3 decimal places
		maxStrDec = maxStrDec[1:4]
	} else {
		maxStrDec = ""
	}
	maxStr := maxStrInt + maxStrDec + " "

	grid[len(grid)-1][1] = '-'
	for i, rune := range maxStr {
		grid[0][i+1] = rune
		grid[len(grid)-1][i+2] = rune
	}

	builder := strings.Builder{}
	writeChartHLine(&builder, width)
	builder.WriteRune('\n')
	for _, row := range grid {
		for _, cell := range row {
			builder.WriteRune(cell)
		}
	}
	writeChartHLine(&builder, width)
	return builder.String()
}
