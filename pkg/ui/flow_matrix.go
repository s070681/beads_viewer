package ui

import (
	"fmt"
	"strings"

	"github.com/Dicklesworthstone/beads_viewer/pkg/analysis"
)

// FlowMatrixView renders a simple label->label dependency matrix
// Rows = from/blocking labels, Cols = to/blocked labels
func FlowMatrixView(flow analysis.CrossLabelFlow, width int) string {
	if len(flow.Labels) == 0 {
		return "No label flows available"
	}
	labels := flow.Labels
	maxLabel := 0
	for _, l := range labels {
		if len(l) > maxLabel {
			maxLabel = len(l)
		}
	}
	cellWidth := 4
	leftWidth := maxLabel
	if leftWidth < 6 {
		leftWidth = 6
	}
	// header
	var b strings.Builder
	pad := func(s string, w int) string {
		if len(s) >= w {
			return s
		}
		return s + strings.Repeat(" ", w-len(s))
	}
	truncate := func(s string, w int) string {
		if len(s) <= w {
			return s
		}
		if w <= 1 {
			return s[:w]
		}
		return s[:w-1] + "…"
	}

	// header row
	b.WriteString(pad("", leftWidth))
	b.WriteString(" | ")
	for _, l := range labels {
		b.WriteString(pad(truncate(l, cellWidth), cellWidth))
	}
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", leftWidth+3+cellWidth*len(labels)))
	b.WriteString("\n")

	for i, row := range flow.FlowMatrix {
		b.WriteString(pad(truncate(labels[i], leftWidth), leftWidth))
		b.WriteString(" | ")
		for _, v := range row {
			b.WriteString(fmt.Sprintf("%*d", cellWidth, v))
		}
		b.WriteString("\n")
	}

	return b.String()
}
