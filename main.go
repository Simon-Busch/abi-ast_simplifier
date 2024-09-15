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
	AST 					AST 			`json:"ast, omitempty"`
}

type AST struct {
	Nodes []ASTNode `json:"nodes"`
}

type ASTNode struct {
	ID            int            `json:"id"`
	NodeType      string         `json:"nodeType"`
	Name          string         `json:"name,omitempty"`
	BaseContracts []BaseContract `json:"baseContracts,omitempty"`
	Nodes         []ASTNode      `json:"nodes,omitempty"` // Child nodes
}
type BaseContract struct {
	ID       int      `json:"id"`
	NodeType string   `json:"nodeType"`
	Src      string   `json:"src"`
	BaseName BaseName `json:"baseName"`
}
type BaseName struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	NodeType             string `json:"nodeType"`
	ReferencedDeclaration int    `json:"referencedDeclaration"`
	Src                  string `json:"src"`
}
type LibraryName struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	NodeType             string `json:"nodeType"`
	ReferencedDeclaration int    `json:"referencedDeclaration"`
	Src                  string `json:"src"`
}


type TypeName struct {
	ID               int              `json:"id"`
	Name             string           `json:"name"`
	NodeType         string           `json:"nodeType"`
	Src              string           `json:"src"`
	StateMutability  string           `json:"stateMutability,omitempty"`
	TypeDescriptions *TypeDescriptions `json:"typeDescriptions,omitempty"`
}

type TypeDescriptions struct {
	TypeIdentifier string `json:"typeIdentifier"`
	TypeString     string `json:"typeString"`
}

type Contract struct {
	Name      string
	Functions []Function
	Events    []Event
	Inherits  []string
	Calls     map[string][]string
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
	dataFolder := "data"

	contracts, err := parseAllABIs(dataFolder)
	if err != nil {
		fmt.Println("Error parsing ABI files:", err)
		return
	}

	app := tview.NewApplication()

	contractsList := tview.NewList().ShowSecondaryText(false)
	functionsList := tview.NewList().ShowSecondaryText(false)
	detailsText := tview.NewTextView().SetDynamicColors(true)

	var selectedContract *Contract

	for name := range contracts {
		contractsList.AddItem(name, "", 0, nil)
	}

	flex := tview.NewFlex().AddItem(contractsList, 0, 1, true)

	selectContract := func(index int, mainText string) {
		contract := contracts[mainText]
		selectedContract = &contract

		inheritsText := ""
		if len(contract.Inherits) > 0 {
			inheritsText = fmt.Sprintf("Inherits: %v\n", contract.Inherits)
		} else {
			inheritsText = "No inheritance\n"
		}
		detailsText.SetText("[yellow]" + inheritsText + "[white]")

		functionsList.Clear()
		for _, function := range contract.Functions {
			functionsList.AddItem(function.Name, "", 0, nil)
		}

		flex.RemoveItem(functionsList)
		flex.RemoveItem(detailsText)
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

	functionsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			flex.RemoveItem(functionsList)
			flex.RemoveItem(detailsText)
			selectedContract = nil // Reset selected contract
			app.SetFocus(contractsList)
			return nil
		case tcell.KeyRight:
			return nil
		}
		return event
	})

	// Handle contracts list navigation
	contractsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRight:
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

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func parseABIFile(path string) ([]ABIEntry, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	// Try to unmarshal as an array first
	var abiEntries []ABIEntry
	err = json.Unmarshal(data, &abiEntries)
	if err == nil {
		// Successfully parsed as array
		// No 'ast' field, so inheritance is nil
		return abiEntries, nil, nil
	}

	var abiFile ABIFile
	err = json.Unmarshal(data, &abiFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse ABI file %s: %w", path, err)
	}

	var inherits []string
	if abiFile.AST.Nodes != nil {
		// Extract inheritance information
		inherits = extractInheritanceFromAST(abiFile.AST)
	}

	return abiFile.ABI, inherits, nil
}

func extractInheritanceFromAST(ast AST) []string {
	var inherits []string
	for _, node := range ast.Nodes {
			if node.NodeType == "ContractDefinition" {
					inherits = append(inherits, extractInheritanceFromNode(node)...)
					break
			}
	}
	return inherits
}

func extractInheritanceFromNode(node ASTNode) []string {
	var inherits []string
	if node.NodeType == "ContractDefinition" {
			fmt.Printf("Processing InheritanceSpecifier: %s\n", node.Name)
			for _, baseContract := range node.BaseContracts {
					inherits = append(inherits, baseContract.BaseName.Name)
			}
	}
	return inherits
}

func parseAllABIs(dataFolder string) (map[string]Contract, error) {
	contracts := make(map[string]Contract)
	err := filepath.Walk(dataFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing file %s: %w", path, err)
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			abiEntries, inherits, err := parseABIFile(path)
			if err != nil {
				return fmt.Errorf("error parsing file %s: %w", path, err)
			}
			contractName := filepath.Base(path)
			contractName = contractName[:len(contractName)-len(filepath.Ext(contractName))]
			contract := Contract{
				Name:     contractName,
				Calls:    make(map[string][]string),
				Inherits: inherits,
			}
			for _, entry := range abiEntries {
				switch entry.Type {
				case "function":
					function := Function{
						Name:            entry.Name,
						Inputs:          entry.Inputs,
						Outputs:         entry.Outputs,
						StateMutability: entry.StateMutability,
					}
					contract.Functions = append(contract.Functions, function)
				case "event":
					event := Event{
						Name:   entry.Name,
						Inputs: entry.Inputs,
					}
					contract.Events = append(contract.Events, event)
				}
			}
			contracts[contractName] = contract
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return contracts, nil
}
