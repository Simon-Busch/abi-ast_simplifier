// ui.go
package ui

import (
    termui "github.com/gizak/termui/v3"
    "github.com/gizak/termui/v3/widgets"
)

func UpdateUI(contractsList *widgets.List, detailsList *widgets.List, codeParagraph *widgets.Paragraph, contractsListSelected bool, detailsListSelected bool) {
	termWidth, termHeight := termui.TerminalDimensions()

	// Set sizes and positions
	contractsList.SetRect(0, 0, termWidth/4, termHeight)
	detailsList.SetRect(termWidth/4, 0, termWidth/2, termHeight)
	codeParagraph.SetRect(termWidth/2, 0, termWidth, termHeight)

	// Highlight selected widget
	if contractsListSelected {
		contractsList.BorderStyle = termui.NewStyle(termui.ColorGreen)
		detailsList.BorderStyle = termui.NewStyle(termui.ColorWhite)
	} else if detailsListSelected {
		contractsList.BorderStyle = termui.NewStyle(termui.ColorWhite)
		detailsList.BorderStyle = termui.NewStyle(termui.ColorGreen)
	}

	// Ensure SelectedRow is valid
	validateSelectedRow(contractsList)
	validateSelectedRow(detailsList)

	termui.Render(contractsList, detailsList, codeParagraph)
}

func validateSelectedRow(list *widgets.List) {
	if len(list.Rows) == 0 {
		list.SelectedRow = 0
	} else if list.SelectedRow >= len(list.Rows) {
		list.SelectedRow = len(list.Rows) - 1
	} else if list.SelectedRow < 0 {
		list.SelectedRow = 0
	}
}
