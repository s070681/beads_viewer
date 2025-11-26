package ui

import (
	"fmt"
	"strings"

	"beads_viewer/pkg/analysis"
	"github.com/charmbracelet/lipgloss"
)

type InsightsModel struct {
	insights analysis.Insights
	ready    bool
	width    int
	height   int
}

func NewInsightsModel(ins analysis.Insights) InsightsModel {
	return InsightsModel{
		insights: ins,
	}
}

func (i *InsightsModel) SetSize(w, h int) {
	i.width = w
	i.height = h
	i.ready = true
}

func (i InsightsModel) View() string {
	if !i.ready {
		return ""
	}

	// Layout:
	// [ Top Bottlenecks ] [ Top Keystones ]
	// [     Cycles      ] [    Stats      ]

	colWidth := (i.width / 3) - 3
	if colWidth < 20 {
		colWidth = 20
	}
	rowHeight := (i.height / 2) - 2
	if rowHeight < 6 {
		rowHeight = 6
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorSecondary).
		Padding(0, 1).
		Width(colWidth).
		Height(rowHeight)

	titleStyle := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)

	// Bottlenecks
	var bnSb strings.Builder
	bnSb.WriteString(titleStyle.Render("🚧 Top Bottlenecks (Betweenness)"))
	bnSb.WriteString("\n\n")
	for _, id := range i.insights.Bottlenecks {
		bnSb.WriteString(fmt.Sprintf("• %s\n", id))
	}
	bnBox := boxStyle.Render(bnSb.String())

	// Keystones
	var ksSb strings.Builder
	ksSb.WriteString(titleStyle.Render("🏛️  Keystones (High Impact)"))
	ksSb.WriteString("\n\n")
	for _, id := range i.insights.Keystones {
		ksSb.WriteString(fmt.Sprintf("• %s\n", id))
	}
	ksBox := boxStyle.Render(ksSb.String())

	// Influencers (Eigenvector)
	var infSb strings.Builder
	infSb.WriteString(titleStyle.Render("🌐 Influencers (Eigenvector)"))
	infSb.WriteString("\n\n")
	for _, id := range i.insights.Influencers {
		infSb.WriteString(fmt.Sprintf("• %s\n", id))
	}
	infBox := boxStyle.Render(infSb.String())

	// Hubs & Authorities
	var haSb strings.Builder
	haSb.WriteString(titleStyle.Render("🛰️  Flow Roles (Hubs / Authorities)"))
	haSb.WriteString("\n\n")
	maxLen := len(i.insights.Hubs)
	if len(i.insights.Authorities) > maxLen {
		maxLen = len(i.insights.Authorities)
	}
	for idx := 0; idx < maxLen; idx++ {
		hub := ""
		auth := ""
		if idx < len(i.insights.Hubs) {
			hub = i.insights.Hubs[idx]
		}
		if idx < len(i.insights.Authorities) {
			auth = i.insights.Authorities[idx]
		}
		haSb.WriteString(fmt.Sprintf("• H: %-12s  A: %s\n", hub, auth))
	}
	haBox := boxStyle.Render(haSb.String())

	// Cycles
	var cySb strings.Builder
	cySb.WriteString(titleStyle.Render("🔄 Structural Risks (Cycles)"))
	cySb.WriteString("\n\n")
	if len(i.insights.Cycles) == 0 {
		cySb.WriteString(lipgloss.NewStyle().Foreground(ColorStatusOpen).Render("No cycles detected. Graph is healthy."))
	} else {
		for _, cycle := range i.insights.Cycles {
			cySb.WriteString(fmt.Sprintf("• %s\n", strings.Join(cycle, " -> ")))
		}
	}
	cyBox := boxStyle.Render(cySb.String())

	// Stats
	var stSb strings.Builder
	stSb.WriteString(titleStyle.Render("📊 Network Health"))
	stSb.WriteString("\n\n")
	stSb.WriteString(fmt.Sprintf("Density: %.4f\n", i.insights.ClusterDensity))
	stBox := boxStyle.Render(stSb.String())

	// Layout: 3 columns top, 2 columns bottom
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, bnBox, ksBox, infBox)
	btmRow := lipgloss.JoinHorizontal(lipgloss.Top, haBox, cyBox, stBox)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, btmRow)
}
