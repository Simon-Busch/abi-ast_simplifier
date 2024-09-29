// main.go
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Simon-Busch/abi_simplifier/parser"
	"github.com/Simon-Busch/abi_simplifier/ui"
	termui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func main() {
	dataFolder := "data"

	contracts, err := parser.ParseAllContracts(dataFolder)
	if err != nil {
		fmt.Println("Error parsing contract files:", err)
		return
	}

	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer termui.Close()

	// Create widgets
	contractsList := widgets.NewList()
	contractsList.Title = "Contracts"
	contractsList.TextStyle = termui.NewStyle(termui.ColorGreen)
	contractsList.WrapText = false

	detailsList := widgets.NewList()
	detailsList.Title = "Details"
	detailsList.TextStyle = termui.NewStyle(termui.ColorGreen)
	detailsList.WrapText = false

	codeParagraph := widgets.NewParagraph()
	codeParagraph.Title = "Code"
	codeParagraph.WrapText = true

	// Populate contracts list
	var contractNames []string
	for name := range contracts {
		contractNames = append(contractNames, name)
	}
	contractsList.Rows = contractNames
	contractsList.SelectedRow = 0

	// Variables to keep track of selections
	var selectedContract *parser.Contract
	var contractsListSelected = true  // Initially, contracts list is selected
	var detailsListSelected = false   // Details list is not selected

	ui.UpdateUI(
		contractsList,
		detailsList,
		codeParagraph,
		contractsListSelected,
		detailsListSelected,
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
				detailsList,
				codeParagraph,
				contractsListSelected,
				detailsListSelected,
			)
		case "<Down>":
			if contractsListSelected {
				if len(contractsList.Rows) > 0 {
					contractsList.ScrollDown()
				}
			} else if detailsListSelected {
				if len(detailsList.Rows) > 0 {
					detailsList.ScrollDown()
				}
			}
		case "<Up>":
			if contractsListSelected {
				if len(contractsList.Rows) > 0 {
					contractsList.ScrollUp()
				}
			} else if detailsListSelected {
				if len(detailsList.Rows) > 0 {
					detailsList.ScrollUp()
				}
			}
		case "<Right>":
			if contractsListSelected {
				// Contract selected, show details
				if len(contractsList.Rows) == 0 {
					continue
				}
				contractName := contractsList.Rows[contractsList.SelectedRow]
				contract := contracts[contractName]
				selectedContract = contract

				// Populate details list with functions, variables, events, structs, enums
				var details []string

				// Constructor
				if selectedContract.Constructor != nil {
					details = append(details, "[Constructor](fg:cyan)")
					details = append(details, "  - Constructor")
				}

				// Functions
				details = append(details, "[Functions](fg:cyan)")
				for _, function := range contract.Functions {
					details = append(details, "  "+function.Name)
				}

				// Mappings
				details = append(details, "[Mappings](fg:cyan)")
				for _, mapping := range selectedContract.Mappings {
					details = append(details, "  "+mapping.Name)
				}

				// Constants
				details = append(details, "[Constants](fg:cyan)")
				for _, constant := range selectedContract.Constants {
					details = append(details, "  "+constant.Name)
				}
				// Variables
				details = append(details, "[Variables](fg:cyan)")
				for _, variable := range contract.Variables {
					details = append(details, "  "+variable.Name)
				}

				// Events
				details = append(details, "[Events](fg:cyan)")
				for _, event := range contract.Events {
					details = append(details, "  "+event.Name)
				}

				// Structs
				details = append(details, "[Structs](fg:cyan)")
				for _, strct := range contract.Structs {
					details = append(details, "  "+strct.Name)
				}

				// Enums
				details = append(details, "[Enums](fg:cyan)")
				for _, enum := range contract.Enums {
					details = append(details, "  "+enum.Name)
				}

				detailsList.Rows = details
				detailsList.SelectedRow = 0 // Reset SelectedRow

				detailsList.Title = "Details of " + contract.Name

				// Update code paragraph with contract summary
				codeText := fmt.Sprintf("Contract: %s\n", contract.Name)
				codeText += fmt.Sprintf("Pragma: %s\n", contract.Pragma)
				if len(contract.Inherits) > 0 {
					codeText += fmt.Sprintf("Inherits: %v\n", contract.Inherits)
				} else {
					codeText += "Inherits: None\n"
				}
				codeParagraph.Text = codeText

				// Switch selection to details list
				detailsListSelected = true
				contractsListSelected = false
				ui.UpdateUI(
						contractsList,
						detailsList,
						codeParagraph,
						contractsListSelected,
						detailsListSelected,
				)
			} else if detailsListSelected {
				// Item selected, show code/details
				if selectedContract == nil || len(detailsList.Rows) == 0 {
					continue
				}
				selectedRow := detailsList.Rows[detailsList.SelectedRow]
				if strings.HasPrefix(selectedRow, "[") {
					continue
				}

				// Determine the section type based on previous headers
				index := detailsList.SelectedRow
				itemType := ""
				for i := index; i >= 0; i-- {
					row := detailsList.Rows[i]
					if strings.HasPrefix(row, "[") {
						// Remove formatting to get the item type
						itemType = strings.Trim(row, "[]()fg:cyan")
						break
					}
				}

				itemName := strings.TrimSpace(selectedRow)
				switch itemType {
				case "Constructor":
					// Display constructor details
					constructor := selectedContract.Constructor
					constructorDetails := "Constructor\n"
					if len(constructor.Parameters) > 0 {
							constructorDetails += "Parameters:\n"
							for _, param := range constructor.Parameters {
									constructorDetails += fmt.Sprintf("  - %s: %s\n", param.Name, param.Type)
							}
					}
					if len(constructor.Modifiers) > 0 {
							constructorDetails += "Modifiers:\n"
							for _, mod := range constructor.Modifiers {
									constructorDetails += fmt.Sprintf("  - %s\n", mod)
							}
					}
					constructorDetails += fmt.Sprintf("Visibility: %s\n", constructor.Visibility)
					constructorDetails += fmt.Sprintf("State Mutability: %s\n", constructor.StateMutability)
					codeParagraph.Text = constructorDetails
				case "Functions":
					var selectedFunction parser.Function
					for _, fn := range selectedContract.Functions {
						if fn.Name == itemName {
							selectedFunction = fn
							break
						}
					}
					// Display function details
					functionDetails := fmt.Sprintf("Function: %s\n", selectedFunction.Name)
					if len(selectedFunction.Parameters) > 0 {
						functionDetails += "Parameters:\n"
						for _, param := range selectedFunction.Parameters {
							functionDetails += fmt.Sprintf("  - %s: %s\n", param.Name, param.Type)
						}
					}
					if len(selectedFunction.ReturnParameters) > 0 {
						functionDetails += "Returns:\n"
						for _, param := range selectedFunction.ReturnParameters {
							functionDetails += fmt.Sprintf("  - %s: %s\n", param.Name, param.Type)
						}
					}
					if len(selectedFunction.Modifiers) > 0 {
						functionDetails += "Modifiers:\n"
						for _, mod := range selectedFunction.Modifiers {
							functionDetails += fmt.Sprintf("  - %s\n", mod)
						}
					}
					functionDetails += fmt.Sprintf("Visibility: %s\n", selectedFunction.Visibility)
					functionDetails += fmt.Sprintf("State Mutability: %s\n", selectedFunction.StateMutability)
					codeParagraph.Text = functionDetails
				case "Constants":
					var selectedConstant parser.Variable
					for _, c := range selectedContract.Constants {
						if c.Name == itemName {
							selectedConstant = c
							break
						}
					}
					// Display constant details
					constantDetails := fmt.Sprintf("Constant: %s\n", selectedConstant.Name)
					constantDetails += fmt.Sprintf("Type: %s\n", selectedConstant.Type)
					constantDetails += fmt.Sprintf("Visibility: %s\n", selectedConstant.Visibility)
					if selectedConstant.Value != "" {
						constantDetails += fmt.Sprintf("Value: %s\n", selectedConstant.Value)
					}
					codeParagraph.Text = constantDetails
				case "Variables":
					var selectedVariable parser.Variable
					for _, v := range selectedContract.Variables {
						if v.Name == itemName {
							selectedVariable = v
							break
						}
					}
					// Display variable details
					variableDetails := fmt.Sprintf("Variable: %s\n", selectedVariable.Name)
					variableDetails += fmt.Sprintf("Type: %s\n", selectedVariable.Type)
					variableDetails += fmt.Sprintf("Visibility: %s\n", selectedVariable.Visibility)
					if selectedVariable.Constant {
						variableDetails += "Constant: true\n"
					}
					if selectedVariable.Mutability != "" {
						variableDetails += fmt.Sprintf("Mutability: %s\n", selectedVariable.Mutability)
					}
					if selectedVariable.Value != "" {
						variableDetails += fmt.Sprintf("Value: %s\n", selectedVariable.Value)
					}
					codeParagraph.Text = variableDetails
				case "Events":
					var selectedEvent parser.Event
					for _, e := range selectedContract.Events {
						if e.Name == itemName {
							selectedEvent = e
							break
						}
					}
					// Display event details
					eventDetails := fmt.Sprintf("Event: %s\n", selectedEvent.Name)
					if len(selectedEvent.Parameters) > 0 {
						eventDetails += "Parameters:\n"
						for _, param := range selectedEvent.Parameters {
							indexedStr := ""
							if param.Indexed {
								indexedStr = "(indexed)"
							}
							eventDetails += fmt.Sprintf("  - %s %s %s\n", param.Type, param.Name, indexedStr)
						}
					}
					codeParagraph.Text = eventDetails
				case "Structs":
					var selectedStruct parser.Struct
					for _, s := range selectedContract.Structs {
						if s.Name == itemName {
							selectedStruct = s
							break
						}
					}
					// Display struct details
					structDetails := fmt.Sprintf("Struct: %s\n", selectedStruct.Name)
					if len(selectedStruct.Members) > 0 {
						structDetails += "Members:\n"
						for _, member := range selectedStruct.Members {
							structDetails += fmt.Sprintf("  - %s: %s\n", member.Name, member.Type)
						}
					}
					codeParagraph.Text = structDetails
				case "Enums":
					var selectedEnum parser.Enum
					for _, e := range selectedContract.Enums {
						if e.Name == itemName {
							selectedEnum = e
							break
						}
					}
					// Display enum details
					enumDetails := fmt.Sprintf("Enum: %s\n", selectedEnum.Name)
					if len(selectedEnum.Values) > 0 {
						enumDetails += "Values:\n"
						for _, value := range selectedEnum.Values {
							enumDetails += fmt.Sprintf("  - %s\n", value)
						}
					}
					codeParagraph.Text = enumDetails
				case "Mappings":
					var selectedMapping parser.Variable
					for _, m := range selectedContract.Mappings {
						if m.Name == itemName {
							selectedMapping = m
							break
						}
					}
					// Display mapping details
					mappingDetails := fmt.Sprintf("Mapping: %s\n", selectedMapping.Name)
					mappingDetails += fmt.Sprintf("Type: %s\n", selectedMapping.Type)
					mappingDetails += fmt.Sprintf("Visibility: %s\n", selectedMapping.Visibility)
					codeParagraph.Text = mappingDetails
				}
				ui.UpdateUI(
					contractsList,
					detailsList,
					codeParagraph,
					contractsListSelected,
					detailsListSelected,
				)
			}
		case "<Left>":
			if detailsListSelected {
				// Go back to contracts list
				detailsListSelected = false
				contractsListSelected = true
				selectedContract = nil
				detailsList.Rows = []string{}
				detailsList.Title = "Details"
				detailsList.SelectedRow = 0
				codeParagraph.Text = ""
				ui.UpdateUI(
					contractsList,
					detailsList,
					codeParagraph,
					contractsListSelected,
					detailsListSelected,
				)
			}
		}
		ui.UpdateUI(
			contractsList,
			detailsList,
			codeParagraph,
			contractsListSelected,
			detailsListSelected,
		)
	}
}
