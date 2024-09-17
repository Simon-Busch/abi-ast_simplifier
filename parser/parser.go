package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ABIEntry represents an entry in the ABI array.
type ABIEntry struct {
	Type            string   `json:"type"`
	Name            string   `json:"name,omitempty"`
	Inputs          []Param  `json:"inputs,omitempty"`
	Outputs         []Param  `json:"outputs,omitempty"`
	StateMutability string   `json:"stateMutability,omitempty"`
}

// Param represents a parameter in a function or event.
type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ABIFile represents the structure of the ABI JSON file.
type ABIFile struct {
	ContractName string     `json:"contractName,omitempty"`
	ABI          []ABIEntry `json:"abi"`
	AST          AST        `json:"ast,omitempty"`
}

// AST represents the Abstract Syntax Tree of the contract.
type AST struct {
	Nodes []ASTNode `json:"nodes"`
}

// ASTNode represents a node in the AST.
type ASTNode struct {
	ID               int            `json:"id"`
	NodeType         string         `json:"nodeType"`
	Name             string         `json:"name,omitempty"`
	Nodes            []ASTNode      `json:"nodes,omitempty"`
	BaseContracts    []BaseContract `json:"baseContracts,omitempty"`
	Parameters       *ParameterList `json:"parameters,omitempty"`
	ReturnParameters *ParameterList `json:"returnParameters,omitempty"`
	Visibility       string         `json:"visibility,omitempty"`
	StateMutability  string         `json:"stateMutability,omitempty"`
	Kind             string         `json:"kind,omitempty"`
	TypeName         *TypeName      `json:"typeName,omitempty"`
	StorageLocation  string         `json:"storageLocation,omitempty"`
}

type ParameterList struct {
	Parameters []ASTNode `json:"parameters"`
}

// BaseContract represents a base contract in inheritance.
type BaseContract struct {
	ID       int      `json:"id"`
	NodeType string   `json:"nodeType"`
	Src      string   `json:"src"`
	BaseName BaseName `json:"baseName"`
}

// BaseName represents the name of a base contract.
type BaseName struct {
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	NodeType              string `json:"nodeType"`
	ReferencedDeclaration int    `json:"referencedDeclaration"`
	Src                   string `json:"src"`
}

// TypeName represents the type of a variable or parameter.
type TypeName struct {
	ID               int              `json:"id"`
	Name             string           `json:"name"`
	NodeType         string           `json:"nodeType"`
	Src              string           `json:"src"`
	StateMutability  string           `json:"stateMutability,omitempty"`
	TypeDescriptions *TypeDescriptions `json:"typeDescriptions,omitempty"`
}

// TypeDescriptions provides type information.
type TypeDescriptions struct {
	TypeIdentifier string `json:"typeIdentifier"`
	TypeString     string `json:"typeString"`
}

// Contract represents a smart contract with its functions and events.
type Contract struct {
	Name      string
	Functions []Function
	Events    []Event
	Inherits  []string
	Calls     map[string][]string
}

// Function represents a function in a contract.
type Function struct {
	Name            string
	Inputs          []Param
	Outputs         []Param
	StateMutability string
}

// Event represents an event in a contract.
type Event struct {
	Name   string
	Inputs []Param
}


// ParseABIFile parses a single ABI file and extracts the ABI entries and inheritance.
func ParseABIFile(path string) ([]ABIEntry, []string, error) {
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
			inherits = ExtractInheritanceFromAST(abiFile.AST)
	}

	return abiFile.ABI, inherits, nil
}

// ExtractInheritanceFromAST extracts inheritance information from the AST.
func ExtractInheritanceFromAST(ast AST) []string {
	var inherits []string
	for _, node := range ast.Nodes {
			if node.NodeType == "ContractDefinition" {
					inherits = append(inherits, ExtractInheritanceFromNode(node)...)
					break
			}
	}
	return inherits
}

// ExtractInheritanceFromNode extracts inheritance from a ContractDefinition node.
func ExtractInheritanceFromNode(node ASTNode) []string {
	var inherits []string
	if node.NodeType == "ContractDefinition" {
			for _, baseContract := range node.BaseContracts {
					inherits = append(inherits, baseContract.BaseName.Name)
			}
	}
	return inherits
}

// ParseAllABIs parses all ABI files in the data folder.
func ParseAllABIs(dataFolder string) (map[string]Contract, error) {
	contracts := make(map[string]Contract)
	err := filepath.Walk(dataFolder, func(path string, info os.FileInfo, err error) error {
			if err != nil {
					return fmt.Errorf("error accessing file %s: %w", path, err)
			}
			if !info.IsDir() && filepath.Ext(path) == ".json" {
					abiEntries, inherits, err := ParseABIFile(path)
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
