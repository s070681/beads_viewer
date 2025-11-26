package ui

import (
	"fmt"
	"sort"

	"beads_viewer/pkg/model"

	"github.com/charmbracelet/lipgloss"
)

// BoardModel represents the Kanban board view with 4 columns
type BoardModel struct {
	columns     [4][]model.Issue
	focusedCol  int
	selectedRow [4]int // Store selection for each column
	ready       bool
	width       int
	height      int
	theme       Theme
}

// Column indices for the Kanban board
const (
	ColOpen       = 0
	ColInProgress = 1
	ColBlocked    = 2
	ColClosed     = 3
)

// sortIssuesByPriorityAndDate sorts issues by priority (ascending) then by creation date (descending)
func sortIssuesByPriorityAndDate(issues []model.Issue) {
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].Priority != issues[j].Priority {
			return issues[i].Priority < issues[j].Priority
		}
		return issues[i].CreatedAt.After(issues[j].CreatedAt)
	})
}

// NewBoardModel creates a new Kanban board from the given issues
func NewBoardModel(issues []model.Issue, theme Theme) BoardModel {
	var cols [4][]model.Issue

	// Distribute issues into columns by status
	for _, issue := range issues {
		switch issue.Status {
		case model.StatusOpen:
			cols[ColOpen] = append(cols[ColOpen], issue)
		case model.StatusInProgress:
			cols[ColInProgress] = append(cols[ColInProgress], issue)
		case model.StatusBlocked:
			cols[ColBlocked] = append(cols[ColBlocked], issue)
		case model.StatusClosed:
			cols[ColClosed] = append(cols[ColClosed], issue)
		}
	}

	// Sort each column
	for i := 0; i < 4; i++ {
		sortIssuesByPriorityAndDate(cols[i])
	}

	return BoardModel{
		columns:    cols,
		focusedCol: 0,
		theme:      theme,
	}
}

// SetIssues updates the board data, typically after filtering
func (b *BoardModel) SetIssues(issues []model.Issue) {
	var cols [4][]model.Issue

	for _, issue := range issues {
		switch issue.Status {
		case model.StatusOpen:
			cols[ColOpen] = append(cols[ColOpen], issue)
		case model.StatusInProgress:
			cols[ColInProgress] = append(cols[ColInProgress], issue)
		case model.StatusBlocked:
			cols[ColBlocked] = append(cols[ColBlocked], issue)
		case model.StatusClosed:
			cols[ColClosed] = append(cols[ColClosed], issue)
		}
	}

	// Sort each column
	for i := 0; i < 4; i++ {
		sortIssuesByPriorityAndDate(cols[i])
	}

	b.columns = cols

	// Sanitize selection to prevent out-of-bounds
	for i := 0; i < 4; i++ {
		if b.selectedRow[i] >= len(b.columns[i]) {
			if len(b.columns[i]) > 0 {
				b.selectedRow[i] = len(b.columns[i]) - 1
			} else {
				b.selectedRow[i] = 0
			}
		}
	}
}

// Navigation methods
func (b *BoardModel) MoveDown() {
	count := len(b.columns[b.focusedCol])
	if count == 0 {
		return
	}
	if b.selectedRow[b.focusedCol] < count-1 {
		b.selectedRow[b.focusedCol]++
	}
}

func (b *BoardModel) MoveUp() {
	if b.selectedRow[b.focusedCol] > 0 {
		b.selectedRow[b.focusedCol]--
	}
}

func (b *BoardModel) MoveRight() {
	if b.focusedCol < 3 {
		b.focusedCol++
	}
}

func (b *BoardModel) MoveLeft() {
	if b.focusedCol > 0 {
		b.focusedCol--
	}
}

func (b *BoardModel) MoveToTop() {
	b.selectedRow[b.focusedCol] = 0
}

func (b *BoardModel) MoveToBottom() {
	count := len(b.columns[b.focusedCol])
	if count > 0 {
		b.selectedRow[b.focusedCol] = count - 1
	}
}

func (b *BoardModel) PageDown(visibleRows int) {
	count := len(b.columns[b.focusedCol])
	if count == 0 {
		return
	}
	newRow := b.selectedRow[b.focusedCol] + visibleRows/2
	if newRow >= count {
		newRow = count - 1
	}
	b.selectedRow[b.focusedCol] = newRow
}

func (b *BoardModel) PageUp(visibleRows int) {
	newRow := b.selectedRow[b.focusedCol] - visibleRows/2
	if newRow < 0 {
		newRow = 0
	}
	b.selectedRow[b.focusedCol] = newRow
}

// SelectedIssue returns the currently selected issue, or nil if none
func (b *BoardModel) SelectedIssue() *model.Issue {
	col := b.columns[b.focusedCol]
	row := b.selectedRow[b.focusedCol]
	if len(col) > 0 && row < len(col) {
		return &col[row]
	}
	return nil
}

// ColumnCount returns the number of issues in a column
func (b *BoardModel) ColumnCount(col int) int {
	if col >= 0 && col < 4 {
		return len(b.columns[col])
	}
	return 0
}

// TotalCount returns the total number of issues across all columns
func (b *BoardModel) TotalCount() int {
	total := 0
	for i := 0; i < 4; i++ {
		total += len(b.columns[i])
	}
	return total
}

// View renders the Kanban board
func (b BoardModel) View(width, height int) string {
	colWidth := (width - 8) / 4 // Account for borders and gaps
	if colWidth < 20 {
		colWidth = 20
	}

	colHeight := height - 3 // Account for header and border
	if colHeight < 5 {
		colHeight = 5
	}

	t := b.theme

	columnTitles := []string{"OPEN", "IN PROGRESS", "BLOCKED", "CLOSED"}
	columnColors := []lipgloss.AdaptiveColor{t.Open, t.InProgress, t.Blocked, t.Closed}

	var renderedCols []string

	for colIdx := 0; colIdx < 4; colIdx++ {
		isFocused := b.focusedCol == colIdx
		issueCount := len(b.columns[colIdx])

		// Header with issue count
		headerText := fmt.Sprintf("%s (%d)", columnTitles[colIdx], issueCount)
		headerStyle := t.Renderer.NewStyle().
			Width(colWidth).
			Align(lipgloss.Center).
			Bold(true)

		if isFocused {
			headerStyle = headerStyle.
				Background(columnColors[colIdx]).
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#282A36"})
		} else {
			headerStyle = headerStyle.
				Background(lipgloss.AdaptiveColor{Light: "#EEEEEE", Dark: "#333333"}).
				Foreground(columnColors[colIdx])
		}

		header := headerStyle.Render(headerText)

		// Calculate visible rows and scrolling
		visibleRows := colHeight - 2
		if visibleRows < 1 {
			visibleRows = 1
		}

		sel := b.selectedRow[colIdx]
		if sel >= issueCount && issueCount > 0 {
			sel = issueCount - 1
		}

		// Simple scrolling: keep selected row visible
		start := 0
		if sel >= visibleRows {
			start = sel - visibleRows + 1
		}

		end := start + visibleRows
		if end > issueCount {
			end = issueCount
		}

		// Render issue rows
		var rows []string
		for rowIdx := start; rowIdx < end; rowIdx++ {
			issue := b.columns[colIdx][rowIdx]

			rowStyle := t.Renderer.NewStyle().
				Width(colWidth).
				Padding(0, 1).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(t.Border)

			if isFocused && rowIdx == sel {
				rowStyle = rowStyle.
					Background(t.Highlight).
					BorderForeground(t.Primary)
			}

			icon, iconColor := t.GetTypeIcon(string(issue.IssueType))
			prio := GetPriorityIcon(issue.Priority)

			// Line 1: Icon, ID, Priority
			line1 := fmt.Sprintf("%s %s %s",
				t.Renderer.NewStyle().Foreground(iconColor).Render(icon),
				t.Renderer.NewStyle().Bold(true).Foreground(t.Secondary).Render(issue.ID),
				prio,
			)

			// Line 2: Truncated title (UTF-8 safe)
			titleMaxWidth := colWidth - 4
			if titleMaxWidth < 10 {
				titleMaxWidth = 10
			}
			truncatedTitle := truncateRunesHelper(issue.Title, titleMaxWidth, "…")
			line2 := t.Renderer.NewStyle().
				Foreground(t.Base.GetForeground()).
				Render(truncatedTitle)

			rows = append(rows, rowStyle.Render(line1+"\n"+line2))
		}

		// Show scroll indicators if needed
		if issueCount > visibleRows {
			scrollInfo := fmt.Sprintf("↕ %d/%d", sel+1, issueCount)
			scrollStyle := t.Renderer.NewStyle().
				Width(colWidth).
				Align(lipgloss.Center).
				Foreground(t.Secondary)
			rows = append(rows, scrollStyle.Render(scrollInfo))
		}

		// Assemble column content
		content := lipgloss.JoinVertical(lipgloss.Left, rows...)

		colStyle := t.Renderer.NewStyle().
			Width(colWidth).
			Height(colHeight).
			Border(lipgloss.RoundedBorder())

		if isFocused {
			colStyle = colStyle.BorderForeground(t.Primary)
		} else {
			colStyle = colStyle.BorderForeground(t.Secondary)
		}

		renderedCols = append(renderedCols, lipgloss.JoinVertical(lipgloss.Center, header, colStyle.Render(content)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)
}
