// parser.go
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)



// ABIFile represents the structure of the ABI JSON file.
type ABIFile struct {
	ContractName 				string     `json:"contractName,omitempty"`
	ABI          				[]ABIEntry `json:"abi"`
	AST          				AST        `json:"ast,omitempty"`
}

// AST represents the Abstract Syntax Tree of the contract.
type AST struct {
	Nodes 							[]ASTNode `json:"nodes"`
}

// ABIEntry represents an entry in the ABI array.
type ABIEntry struct {
	Type            		string  `json:"type"`
	Name            		string  `json:"name,omitempty"`
	Inputs          		[]Param `json:"inputs,omitempty"`
	Outputs         		[]Param `json:"outputs,omitempty"`
	StateMutability 		string  `json:"stateMutability,omitempty"`
}

// Param represents a parameter in a function or event.
type Param struct {
	Name 								string `json:"name"`
	Type 								string `json:"type"`
}

// ASTNode represents a node in the AST.
type ASTNode struct {
	ID               		int             `json:"id"`
	NodeType         		string          `json:"nodeType"`
	Name             		string          `json:"name,omitempty"`
	Nodes            		[]ASTNode       `json:"nodes,omitempty"`
	BaseContracts    		[]BaseContract  `json:"baseContracts,omitempty"`
	Literals         		[]string        `json:"literals,omitempty"`
	Parameters       		*ParameterList  `jsfon:"parameters,omitempty"`
	ReturnParameters 		*ParameterList  `json:"returnParameters,omitempty"`
	Visibility       		string          `json:"visibility,omitempty"`
	StateMutability  		string          `json:"stateMutability,omitempty"`
	Kind             		string          `json:"kind,omitempty"`
	TypeName         		*TypeName       `json:"typeName,omitempty"`
	StorageLocation  		string          `json:"storageLocation,omitempty"`
}

// ParameterList represents a list of parameters.
type ParameterList struct {
	Parameters 					[]ASTNode `json:"parameters"`
}

// BaseContract represents a base contract in inheritance.
type BaseContract struct {
	ID       						int      `json:"id"`
	NodeType 						string   `json:"nodeType"`
	Src      						string   `json:"src"`
	BaseName 						BaseName `json:"baseName"`
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
	ID               			int               `json:"id"`
	Name             			string            `json:"name"`
	NodeType         			string            `json:"nodeType"`
	Src              			string            `json:"src"`
	StateMutability  			string            `json:"stateMutability,omitempty"`
	TypeDescriptions 			*TypeDescriptions `json:"typeDescriptions,omitempty"`
}

// TypeDescriptions provides type information.
type TypeDescriptions struct {
	TypeIdentifier 				string `json:"typeIdentifier"`
	TypeString     				string `json:"typeString"`
}

// Contract represents a smart contract with its functions and events.
type Contract struct {
	Name      						string
	Functions 						[]*Function
	Events    						[]*Event
	Constructor 					*Constructor
	Inherits  						[]string
	Pragma    						string
}

type Constructor struct {
	Inputs 								[]Param
	StateMutability 			string
}

// Function represents a function in a contract.
type Function struct {
	Name            			string
	Inputs          			[]Param
	Outputs         			[]Param
	StateMutability 			string
}

// Event represents an event in a contract.
type Event struct {
	Name   								string
	Inputs 								[]Param
}

// ParseABIFile parses a single ABI file and extracts the Contract information.
func ParseABIFile(path string) (*Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal as an array first
	var abiEntries []ABIEntry
	err = json.Unmarshal(data, &abiEntries)
	if err == nil {
		// Successfully parsed as array
		// No 'ast' field, so inheritance is nil
		contractName := filepath.Base(path)
		contractName = contractName[:len(contractName)-len(filepath.Ext(contractName))]
		contract := &Contract{
			Name: contractName,
		}
		// Process abiEntries to populate functions and events
		for _, entry := range abiEntries {
			switch entry.Type {
			case "function":
				function := &Function{
					Name:            entry.Name,
					Inputs:          entry.Inputs,
					Outputs:         entry.Outputs,
					StateMutability: entry.StateMutability,
				}
				contract.Functions = append(contract.Functions, function)
			case "event":
				event := &Event{
					Name:   entry.Name,
					Inputs: entry.Inputs,
				}
				contract.Events = append(contract.Events, event)
			case "constructor":
				constructor := &Constructor{
					Inputs:          entry.Inputs,
					StateMutability: entry.StateMutability,
				}
				fmt.Println(constructor)
				contract.Constructor = constructor
			}
		}
		return contract, nil
	}

	// Try to unmarshal as ABIFile
	var abiFile ABIFile
	err = json.Unmarshal(data, &abiFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI file %s: %w", path, err)
	}

	// Now process abiFile
	contractName := abiFile.ContractName
	if contractName == "" {
		contractName = filepath.Base(path)
		contractName = contractName[:len(contractName)-len(filepath.Ext(contractName))]
	}

	contract := &Contract{
		Name: contractName,
	}
	// Process abiEntries to populate functions and events
	for _, entry := range abiFile.ABI {
		switch entry.Type {
		case "function":
			function := &Function{
				Name:            entry.Name,
				Inputs:          entry.Inputs,
				Outputs:         entry.Outputs,
				StateMutability: entry.StateMutability,
			}
			contract.Functions = append(contract.Functions, function)
		case "event":
			event := &Event{
				Name:   entry.Name,
				Inputs: entry.Inputs,
			}
			contract.Events = append(contract.Events, event)
		case "constructor":
			constructor := &Constructor{
				Inputs:          entry.Inputs,
				StateMutability: entry.StateMutability,
			}
			contract.Constructor = constructor
		}
	}

	// Extract information from AST if available
	if len(abiFile.AST.Nodes) > 0 {
		ExtractInfoFromAST(abiFile.AST, contract)
	}

	return contract, nil
}

// ExtractInfoFromAST extracts inheritance and pragma information from the AST and populates the contract.
func ExtractInfoFromAST(ast AST, contract *Contract) {
	for _, node := range ast.Nodes {
		switch node.NodeType {
		case "ContractDefinition":
			ExtractContractDefinition(node, contract)
		case "PragmaDirective":
			contract.Pragma = ExtractPragmaDirective(node)
		}
	}
}

// ExtractContractDefinition extracts inheritance and functions from a ContractDefinition node.
func ExtractContractDefinition(node ASTNode, contract *Contract) {
	// Extract inheritance
	for _, baseContract := range node.BaseContracts {
		contract.Inherits = append(contract.Inherits, baseContract.BaseName.Name)
	}
	// Process child nodes for functions, events, etc.
	for _, childNode := range node.Nodes {
		switch childNode.NodeType {
		case "FunctionDefinition":
			function := ExtractFunction(childNode)
			contract.Functions = append(contract.Functions, function)
		case "EventDefinition":
			event := ExtractEvent(childNode)
			contract.Events = append(contract.Events, event)
		}
	}
}

// ExtractFunction extracts a Function from a FunctionDefinition node.
func ExtractFunction(node ASTNode) *Function {
	function := &Function{
		Name:            node.Name,
		StateMutability: node.StateMutability,
	}
	// Extract inputs
	if node.Parameters != nil {
		for _, paramNode := range node.Parameters.Parameters {
			if paramNode.NodeType == "VariableDeclaration" {
				param := Param{
					Name: paramNode.Name,
					Type: extractTypeName(paramNode.TypeName),
				}
				function.Inputs = append(function.Inputs, param)
			}
		}
	}
	// Extract outputs
	if node.ReturnParameters != nil {
		for _, paramNode := range node.ReturnParameters.Parameters {
			if paramNode.NodeType == "VariableDeclaration" {
				param := Param{
					Name: paramNode.Name,
					Type: extractTypeName(paramNode.TypeName),
				}
				function.Outputs = append(function.Outputs, param)
			}
		}
	}
	return function
}

// ExtractEvent extracts an Event from an EventDefinition node.
func ExtractEvent(node ASTNode) *Event {
	event := &Event{
		Name: node.Name,
	}
	// Extract inputs
	if node.Parameters != nil {
		for _, paramNode := range node.Parameters.Parameters {
			if paramNode.NodeType == "VariableDeclaration" {
				param := Param{
					Name: paramNode.Name,
					Type: extractTypeName(paramNode.TypeName),
				}
				event.Inputs = append(event.Inputs, param)
			}
		}
	}
	return event
}

// ExtractPragmaDirective extracts the pragma directive from a PragmaDirective node.
func ExtractPragmaDirective(node ASTNode) string {
	pragma := ""
	for _, literal := range node.Literals {
		pragma += literal + ""
	}
	return pragma
}

// extractTypeName extracts the type name from a TypeName node.
func extractTypeName(typeName *TypeName) string {
	if typeName == nil {
		return ""
	}
	switch typeName.NodeType {
	case "ElementaryTypeName":
		return typeName.Name
	case "UserDefinedTypeName":
		return typeName.Name
	default:
		return ""
	}
}

// ParseAllABIs parses all ABI files in the data folder.
func ParseAllABIs(dataFolder string) (map[string]*Contract, error) {
	contracts := make(map[string]*Contract)
	err := filepath.Walk(dataFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing file %s: %w", path, err)
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			contract, err := ParseABIFile(path)
			if err != nil {
				return fmt.Errorf("error parsing file %s: %w", path, err)
			}
			contracts[contract.Name] = contract
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return contracts, nil
}
