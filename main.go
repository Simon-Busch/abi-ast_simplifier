// main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Data structures to represent ABI entries and contracts
type ABIEntry struct {
	Type          		string        `json:"type"`
	Name          		string        `json:"name,omitempty"`
	Inputs        		[]Param       `json:"inputs,omitempty"`
	Outputs       		[]Param       `json:"outputs,omitempty"`
	BaseContracts 		[]Inheritance `json:"baseContracts,omitempty"` // For inheritance // ??
	StateMutability 	string `json:"stateMutability,omitempty"`
}

type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Inheritance struct {
	Name string `json:"name"`
}

type ABIFile struct {
	ContractName string     `json:"contractName,omitempty"`
	ABI          []ABIEntry `json:"abi"`
	// Add other fields if necessary
}

type Contract struct {
	Name      string
	Functions []Function
	Events    []Event
	Inherits  []string
	Calls     map[string][]string // Function name to called functions
}

type Function struct {
	Name    string
	Inputs  []Param
	Outputs []Param
	StateMutability string
}

type Event struct {
	Name   string
	Inputs []Param
}

func main() {
	// Specify the data folder containing ABI JSON files
	dataFolder := "data" // You can modify this path as needed

	contracts, err := parseAllABIs(dataFolder)
	if err != nil {
		fmt.Println("Error parsing ABI files:", err)
		return
	}

	// Initialize the application
	app := tview.NewApplication()

	// Create lists for contracts and functions
	contractsList := tview.NewList().ShowSecondaryText(false)
	functionsList := tview.NewList().ShowSecondaryText(false)
	detailsText := tview.NewTextView().SetDynamicColors(true)

	// Variable to keep track of the selected contract
	var selectedContract *Contract

	// Populate the contracts list
	for name := range contracts {
		contractsList.AddItem(name, "", 0, nil)
	}

	// Create a flex layout to organize the UI
	// Initially, only the contractsList is displayed
	flex := tview.NewFlex().AddItem(contractsList, 0, 1, true)

	// Function to handle contract selection
	selectContract := func(index int, mainText string) {
		contract := contracts[mainText]
		selectedContract = &contract

		// Display inheritance information
		inheritsText := ""
		if len(contract.Inherits) > 0 {
			inheritsText = fmt.Sprintf("Inherits: %v", contract.Inherits)
		} else {
			inheritsText = "No inheritance"
		}
		detailsText.SetText("[yellow]" + inheritsText + "[white]")

		// Populate the functions list
		functionsList.Clear()
		for _, function := range contract.Functions {
			functionsList.AddItem(function.Name, "", 0, nil)
		}

		// Update the layout to include functionsList and detailsText
		flex.RemoveItem(functionsList) // Ensure functionsList is not already added
		flex.RemoveItem(detailsText)   // Ensure detailsText is not already added
		flex.AddItem(functionsList, 0, 1, false).
			AddItem(detailsText, 0, 2, false)

		app.SetFocus(functionsList)
	}

	// Handle contract selection
	contractsList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		selectContract(index, mainText)
	})

	// Handle function selection
	functionsList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if selectedContract == nil {
			return
		}
		function := selectedContract.Functions[index]

		// Display function details
		functionDetails := fmt.Sprintf("[green]Function: [white]%s\n", function.Name)
		if len(function.Inputs) > 0 {
			functionDetails += "[yellow]Inputs:\n"
			for _, input := range function.Inputs {
				functionDetails += fmt.Sprintf("  - %s: %s\n", input.Name, input.Type)
			}
		}

		functionDetails += fmt.Sprintf("[green]State mutability: [white]%s\n", function.StateMutability)

		// Display outputs if available
		if len(function.Outputs) > 0 {
			functionDetails += "[yellow]Outputs:\n"
			for _, output := range function.Outputs {
				functionDetails += fmt.Sprintf("  - %s: %s\n", output.Name, output.Type)
			}
		}
		// Display called functions if available
		if calls, ok := selectedContract.Calls[function.Name]; ok && len(calls) > 0 {
			functionDetails += "[yellow]Calls:\n"
			for _, call := range calls {
				functionDetails += fmt.Sprintf("  - %s\n", call)
			}
		} else {
			functionDetails += "[yellow]Calls: [white]None\n"
		}

		detailsText.SetText(functionDetails)
	})

	// Handle left arrow to go back to contracts list and hide functionsList and detailsText
	functionsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			// Remove functionsList and detailsText from the layout
			flex.RemoveItem(functionsList)
			flex.RemoveItem(detailsText)
			selectedContract = nil // Reset selected contract
			app.SetFocus(contractsList)
			return nil
		case tcell.KeyRight:
			// No action on right arrow
			return nil
		}
		return event
	})

	// Handle contracts list navigation
	contractsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRight:
			// Call the selectContract function directly
			if contractsList.GetItemCount() == 0 {
				return nil
			}
			index := contractsList.GetCurrentItem()
			mainText, _ := contractsList.GetItemText(index)
			selectContract(index, mainText)
			return nil
		}
		return event
	})

	// Set up the application root and run it
	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// Function to parse a single ABI file
func parseABIFile(path string) ([]ABIEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal as an array first
	var abiEntries []ABIEntry
	err = json.Unmarshal(data, &abiEntries)
	if err == nil {
		// Successfully parsed as array
		return abiEntries, nil
	}

	// If that fails, try to unmarshal as an object with an 'abi' field
	var abiFile ABIFile
	err = json.Unmarshal(data, &abiFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI file %s: %w", path, err)
	}

	return abiFile.ABI, nil
}

// Function to parse all ABI files in the data folder
func parseAllABIs(dataFolder string) (map[string]Contract, error) {
	contracts := make(map[string]Contract)
	err := filepath.Walk(dataFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing file %s: %w", path, err)
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			abiEntries, err := parseABIFile(path)
			if err != nil {
				return fmt.Errorf("error parsing file %s: %w", path, err)
			}
			contractName := filepath.Base(path)
			contractName = contractName[:len(contractName)-len(filepath.Ext(contractName))] // Remove extension
			contract := Contract{
				Name:  contractName,
				Calls: make(map[string][]string),
			}
			for _, entry := range abiEntries {
				switch entry.Type {
				case "function":
					function := Function{
						Name:    entry.Name,
						Inputs:  entry.Inputs,
						Outputs: entry.Outputs,
						StateMutability: entry.StateMutability,
					}
					contract.Functions = append(contract.Functions, function)
				case "event":
					event := Event{
						Name:   entry.Name,
						Inputs: entry.Inputs,
					}
					contract.Events = append(contract.Events, event)
				// Add more cases if needed
				}
			}
			// Placeholder for inheritance: You might populate Inherits here if you have that data
			contracts[contractName] = contract
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return contracts, nil
}
