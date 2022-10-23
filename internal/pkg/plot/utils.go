package plot

import "strings"

func writeChartHLine(builder *strings.Builder, width int) {
	builder.WriteRune('+')
	builder.WriteString(strings.Repeat("-", width-3))
	builder.WriteRune('+')
}
