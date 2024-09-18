package ui

import (
	termui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func UpdateUI(contractsList *widgets.List, functionsList *widgets.List, detailsParagraph *widgets.Paragraph, contractsListSelected bool, functionsListSelected bool) {
	termWidth, termHeight := termui.TerminalDimensions()

	// Set sizes and positions
	contractsList.SetRect(0, 0, termWidth/4, termHeight)
	functionsList.SetRect(termWidth/4, 0, termWidth/2, termHeight)
	detailsParagraph.SetRect(termWidth/2, 0, termWidth, termHeight)

	// Highlight selected widget
	if contractsListSelected {
			contractsList.BorderStyle = termui.NewStyle(termui.ColorWhite)
			functionsList.BorderStyle = termui.NewStyle(termui.ColorWhite)
	} else if functionsListSelected {
			contractsList.BorderStyle = termui.NewStyle(termui.ColorWhite)
			functionsList.BorderStyle = termui.NewStyle(termui.ColorWhite)
	}

	// Ensure SelectedRow is valid
	validateSelectedRow(contractsList)
	validateSelectedRow(functionsList)

	termui.Render(contractsList, functionsList, detailsParagraph)
}

func validateSelectedRow( list *widgets.List) {
	if len(list.Rows) == 0 {
		list.SelectedRow = 0
	} else if list.SelectedRow >= len(list.Rows) {
		list.SelectedRow = len(list.Rows) - 1
	} else if list.SelectedRow < 0 {
		list.SelectedRow = 0
	}
}
