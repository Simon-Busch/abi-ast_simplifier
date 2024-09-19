// main.go
package main

import (
	"fmt"
	"log"

	"github.com/Simon-Busch/abi_simplifier/parser"
	"github.com/Simon-Busch/abi_simplifier/ui"
	termui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func main() {
	dataFolder := "data"

	contracts, err := parser.ParseAllABIs(dataFolder)
	if err != nil {
			fmt.Println("Error parsing ABI files:", err)
			return
	}

	if err := termui.Init(); err != nil {
			log.Fatalf("failed to initialize termui: %v", err)
	}
	defer termui.Close()

	// Create widgets
	contractsList := widgets.NewList()
	contractsList.Title = "Contracts"
	contractsList.TextStyle = termui.NewStyle(termui.ColorYellow)
	contractsList.WrapText = false

	functionsList := widgets.NewList()
	functionsList.Title = "Functions"
	functionsList.TextStyle = termui.NewStyle(termui.ColorGreen)
	functionsList.WrapText = false


	// TODO 0XSI finish this
	constructorData := widgets.NewList()
	constructorData.Title = "Constructor Data"
	constructorData.TextStyle = termui.NewStyle(termui.ColorGreen)
	constructorData.WrapText = false

	detailsParagraph := widgets.NewParagraph()
	detailsParagraph.Title = "Details"
	detailsParagraph.WrapText = true

	// Populate contracts list
	var contractNames []string
	for name := range contracts {
			contractNames = append(contractNames, name)
	}
	contractsList.Rows = contractNames
	contractsList.SelectedRow = 0

	// Variables to keep track of selections
	var selectedContract *parser.Contract
	var selectedFunction *parser.Function
	var contractsListSelected = true  // Initially, contracts list is selected
	var functionsListSelected = false // Functions list is not selected

	ui.UpdateUI(
		contractsList,
		functionsList,
		detailsParagraph,
		contractsListSelected,
		functionsListSelected,
	)

	// Event handling
	uiEvents := termui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "<Resize>":
			ui.UpdateUI(
				contractsList,
				functionsList,
				detailsParagraph,
				contractsListSelected,
				functionsListSelected,
			)
		case "<Down>":
			if contractsListSelected {
				if len(contractsList.Rows) > 0 {
					contractsList.ScrollDown()
				}
			} else if functionsListSelected {
				if len(functionsList.Rows) > 0 {
					functionsList.ScrollDown()
				}
			}
		case "<Up>":
			if contractsListSelected {
				if len(contractsList.Rows) > 0 {
					contractsList.ScrollUp()
				}
			} else if functionsListSelected {
				if len(functionsList.Rows) > 0 {
					functionsList.ScrollUp()
				}
			}
		case "<Right>":
			if contractsListSelected {
				// Contract selected, show functions
				if len(contractsList.Rows) == 0 {
					continue
				}
				contractName := contractsList.Rows[contractsList.SelectedRow]
				contract := contracts[contractName]
				selectedContract = contract

				// Populate functions list
				var functionNames []string
				for _, function := range contract.Functions {
					functionNames = append(functionNames, function.Name)
				}
				functionsList.Rows = functionNames
				functionsList.SelectedRow = 0 // Reset SelectedRow

				functionsList.Title = "Functions of " + contract.Name

				// Update details
				inheritsText := ""
				if len(contract.Inherits) > 0 {
					inheritsText = fmt.Sprintf("Inherits:[%v](fg:green)\n", contract.Inherits)
				} else {
					inheritsText = "No inheritance\n"
				}
				constructorText := ""
				if contract.Constructor != nil {
					constructorText = "Constructor:\n"
					for _, input := range contract.Constructor.Inputs {
						constructorText += fmt.Sprintf("  - %s: %s\n", input.Name, input.Type)
					}
					constructorText += fmt.Sprintf("State Mutability: [%s](fg:green)\n", contract.Constructor.StateMutability)
				} else {
					constructorText = "No constructor\n"
				}
				detailsParagraph.Text = fmt.Sprintf("%s\n%s \n%s",inheritsText, contract.Pragma, constructorText)


				// Switch selection to functions list
				functionsListSelected = true
				contractsListSelected = false
				ui.UpdateUI(
					contractsList,
					functionsList,
					detailsParagraph,
					contractsListSelected,
					functionsListSelected,
				)
			} else if functionsListSelected {
				// Function selected, show details
				if selectedContract == nil || len(functionsList.Rows) == 0 {
					continue
				}
				functionName := functionsList.Rows[functionsList.SelectedRow]
				for _, fn := range selectedContract.Functions {
					if fn.Name == functionName {
						selectedFunction = fn
						break
					}
				}
				if selectedFunction != nil {
					functionDetails := fmt.Sprintf("Function: [%s](fg:green)\n", selectedFunction.Name)
					if len(selectedFunction.Inputs) > 0 {
						functionDetails += "Inputs:\n"
						for _, input := range selectedFunction.Inputs {
							functionDetails += fmt.Sprintf("  - %s: %s\n", input.Name, input.Type)
						}
					}
					functionDetails += fmt.Sprintf("State Mutability: [%s](fg:green)\n", selectedFunction.StateMutability)
					if len(selectedFunction.Outputs) > 0 {
						functionDetails += "Outputs:\n"
						for _, output := range selectedFunction.Outputs {
							functionDetails += fmt.Sprintf("  - %s: %s\n", output.Name, output.Type)
						}
					}
					detailsParagraph.Text = functionDetails
					ui.UpdateUI(
						contractsList,
						functionsList,
						detailsParagraph,
						contractsListSelected,
						functionsListSelected,
					)
				}
			}
		case "<Left>":
			if functionsListSelected {
				// Go back to contracts list
				functionsListSelected = false
				contractsListSelected = true
				selectedContract = nil
				functionsList.Rows = []string{}
				functionsList.Title = "Functions"
				functionsList.SelectedRow = 0
				detailsParagraph.Text = ""
				ui.UpdateUI(
					contractsList,
					functionsList,
					detailsParagraph,
					contractsListSelected,
					functionsListSelected,
				)
			}
		}
		ui.UpdateUI(
			contractsList,
			functionsList,
			detailsParagraph,
			contractsListSelected,
			functionsListSelected,
		)
	}
}
