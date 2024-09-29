// parser.go
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Contract represents a smart contract with all its components.
type Contract struct {
    Name        string
    Pragma      string
    Imports     []Import
    Inherits    []string
		Constructor *Function
    Variables   []Variable
    Functions   []Function
    Events      []Event
    Modifiers   []Modifier
    Structs     []Struct
    Enums       []Enum
		Mappings 		[]Variable
}

// Import represents an import directive in Solidity.
type Import struct {
    AbsolutePath string
    File         string
    Alias        string
}

// Variable represents a state variable declaration.
type Variable struct {
    Name             string
    Type             string
    Visibility       string
    StateVariable    bool
    StorageLocation  string
    Constant         bool
    Mutability       string // For 'immutable' variables
    FunctionSelector string // For variables with selectors
    Value            string
}

// Function represents a function definition.
type Function struct {
    Name             string
    Visibility       string
		Kind             string // For 'constructor' functions
    StateMutability  string
    Parameters       []Parameter
    ReturnParameters []Parameter
    Modifiers        []string
    BaseFunctions    []int      // IDs of base functions
    Overrides        []string   // Names of contracts being overridden
}

// Event represents an event definition.
type Event struct {
    Name       string
    Parameters []Parameter
}

// Modifier represents a function modifier.
type Modifier struct {
    Name       string
    Parameters []Parameter
}

// Struct represents a struct definition.
type Struct struct {
    Name    string
    Members []Variable
}

// Enum represents an enum definition.
type Enum struct {
    Name   string
    Values []string
}

// Parameter represents a function or event parameter.
type Parameter struct {
    Name string
    Type string
		Indexed bool // For event parameters
}

// ABIFile represents the structure of the ABI JSON file including the AST.
type ABIFile struct {
    ContractName string      `json:"contractName,omitempty"`
    // ABI          interface{} `json:"abi,omitempty"` // We won't use ABI in this parser
    AST          AST         `json:"ast,omitempty"`
}

// AST represents the Abstract Syntax Tree of the contract.
type AST struct {
    Nodes []ASTNode `json:"nodes"`
}

// ASTNode represents a node in the AST.
type ASTNode struct {
    ID                     int               `json:"id"`
    NodeType               string            `json:"nodeType"`
    Name                   string            `json:"name,omitempty"`
    AbsolutePath           string            `json:"absolutePath,omitempty"`
    File                   string            `json:"file,omitempty"`
    BaseContracts          []BaseContract    `json:"baseContracts,omitempty"`
    Members                []ASTNode         `json:"members,omitempty"`
    Modifiers              []ModifierInvocation `json:"modifiers,omitempty"`
    Parameters             *ParameterList    `json:"parameters,omitempty"`
    ReturnParameters       *ParameterList    `json:"returnParameters,omitempty"`
    Visibility             string            `json:"visibility,omitempty"`
    StateMutability        string            `json:"stateMutability,omitempty"`
    Kind                   string            `json:"kind,omitempty"`
    OverloadedDeclarations []int             `json:"overloadedDeclarations,omitempty"`
    BaseFunctions          []int             `json:"baseFunctions,omitempty"`
    Overrides              *OverrideSpecifier `json:"overrides,omitempty"`
    FunctionSelector       string            `json:"functionSelector,omitempty"`
    StorageLocation        string            `json:"storageLocation,omitempty"`
    Constant               bool              `json:"constant,omitempty"`
    Mutability             string            `json:"mutability,omitempty"`
    StateVariable          bool              `json:"stateVariable,omitempty"`
    Value                  interface{}       `json:"value,omitempty"`
    TypeName               *TypeName         `json:"typeName,omitempty"`
    Literals               []string          `json:"literals,omitempty"`
    Nodes                  []ASTNode         `json:"nodes,omitempty"`
    Scope                  int               `json:"scope,omitempty"`
    Operator               string            `json:"operator,omitempty"`        // For UnaryOperation
    SubExpression          *ASTNode          `json:"subExpression,omitempty"`   // For UnaryOperation
    Expression             *ASTNode          `json:"expression,omitempty"`      // For FunctionCall
		Arguments 							[]interface{} `json:"arguments,omitempty"` 					// For FunctionCall
    HexValue               string            `json:"hexValue,omitempty"`        // For Literal nodes
    IsConstant             bool              `json:"isConstant,omitempty"`      // For Literal nodes
    IsLValue               bool              `json:"isLValue,omitempty"`        // For Literal nodes
    IsPure                 bool              `json:"isPure,omitempty"`          // For Literal nodes
    LeftExpression         *ASTNode          `json:"leftExpression,omitempty"`  // For BinaryOperation
    RightExpression        *ASTNode          `json:"rightExpression,omitempty"` // For BinaryOperation
    Indexed 							 *bool 						 `json:"indexed,omitempty"`  				// Indexed parameter for events
}

// BaseContract represents a base contract in inheritance.
type BaseContract struct {
    BaseName BaseName `json:"baseName"`
}

// BaseName represents the name of a base contract.
type BaseName struct {
    Name string `json:"name"`
}

// ParameterList represents a list of parameters.
type ParameterList struct {
    Parameters []ASTNode `json:"parameters"`
}

// ModifierInvocation represents a modifier applied to a function.
type ModifierInvocation struct {
    ID           int       `json:"id"`
    NodeType     string    `json:"nodeType"`
    ModifierName ASTNode   `json:"modifierName"`
    Arguments    []ASTNode `json:"arguments,omitempty"`
    Kind         string    `json:"kind,omitempty"`
    Src          string    `json:"src"`
}

// OverrideSpecifier represents function overrides.
type OverrideSpecifier struct {
    ID        int       `json:"id"`
    NodeType  string    `json:"nodeType"`
    Overrides []ASTNode `json:"overrides"`
    Src       string    `json:"src"`
}

// TypeName represents the type of a variable or parameter.
type TypeName struct {
		NodeType         string            `json:"nodeType"`
		Name             string            `json:"name,omitempty"`
		Path             string            `json:"path,omitempty"`
		BaseType         *TypeName         `json:"baseType,omitempty"`    // For ArrayTypeName
		Length           interface{}       `json:"length,omitempty"`      // For ArrayTypeName
		KeyType          *TypeName         `json:"keyType,omitempty"`     // For Mapping
		ValueType        *TypeName         `json:"valueType,omitempty"`   // For Mapping
		TypeDescriptions *TypeDescriptions `json:"typeDescriptions,omitempty"`
		PathNode         *IdentifierPath   `json:"pathNode,omitempty"`    // Updated to use IdentifierPath struct
}

type IdentifierPath struct {
		ID                 int      `json:"id"`
		Name               string   `json:"name"`
		NameLocations      []string `json:"nameLocations,omitempty"`
		NodeType           string   `json:"nodeType"`
		ReferencedDeclaration int   `json:"referencedDeclaration,omitempty"`
		Src                string   `json:"src"`
}

// TypeDescriptions provides type information.
type TypeDescriptions struct {
    TypeIdentifier string `json:"typeIdentifier"`
    TypeString     string `json:"typeString"`
}

// ParseAllContracts parses all contracts in the specified data folder.
func ParseAllContracts(dataFolder string) (map[string]*Contract, error) {
    contracts := make(map[string]*Contract)
    err := filepath.Walk(dataFolder, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return fmt.Errorf("error accessing file %s: %w", path, err)
        }
        if !info.IsDir() && filepath.Ext(path) == ".json" {
            contract, err := ParseContractFile(path)
            if err != nil {
                return fmt.Errorf("error parsing file %s: %w", path, err)
            }
						if contract.Name != "" {
							contracts[contract.Name] = contract
						}
        }
        return nil
    })
    if err != nil {
        return nil, err
    }
    return contracts, nil
}

// ParseContractFile parses a single contract file and extracts the contract information.
func ParseContractFile(path string) (*Contract, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var abiFile ABIFile
    err = json.Unmarshal(data, &abiFile)
    if err != nil {
        return nil, fmt.Errorf("failed to parse contract file %s: %w", path, err)
    }

    contract := &Contract{
        Name: abiFile.ContractName,
    }

    // Process the AST
    if len(abiFile.AST.Nodes) > 0 {
        err := ExtractContractInfoFromAST(abiFile.AST, contract)
        if err != nil {
            return nil, err
        }
    } else {
        return nil, fmt.Errorf("no AST found in file %s", path)
    }

    return contract, nil
}

// ExtractContractInfoFromAST extracts information from the AST and populates the contract struct.
func ExtractContractInfoFromAST(ast AST, contract *Contract) error {
    for _, node := range ast.Nodes {
        switch node.NodeType {
        case "PragmaDirective":
            contract.Pragma = ExtractPragmaDirective(node)
        case "ImportDirective":
            imp := ExtractImportDirective(node)
            contract.Imports = append(contract.Imports, imp)
        case "ContractDefinition":
            // Process the contract definition
            if contract.Name == "" {
                contract.Name = node.Name
            }
            ExtractContractDefinition(node, contract)
        }
    }
    return nil
}

// ExtractPragmaDirective extracts the pragma directive.
func ExtractPragmaDirective(node ASTNode) string {
    return fmt.Sprintf("pragma %s;", strings.Join(node.Literals, ""))
}

// ExtractImportDirective extracts the import directive.
func ExtractImportDirective(node ASTNode) Import {
    return Import{
        AbsolutePath: node.AbsolutePath,
        File:         node.File,
        Alias:        node.Name, // Adjust if aliasing is handled differently
    }
}

// ExtractContractDefinition processes the ContractDefinition node.
func ExtractContractDefinition(node ASTNode, contract *Contract) {
    // Inheritance
    for _, baseContract := range node.BaseContracts {
        contract.Inherits = append(contract.Inherits, baseContract.BaseName.Name)
    }
    // Process members
    for _, member := range node.Nodes {
        switch member.NodeType {
        case "VariableDeclaration":
					variable := ExtractVariable(member)
					if member.TypeName != nil && member.TypeName.NodeType == "Mapping" {
						contract.Mappings = append(contract.Mappings, variable)
					} else {
							contract.Variables = append(contract.Variables, variable)
					}
        case "FunctionDefinition":
					function := ExtractFunction(member)
					if function.Kind == "constructor" {
						contract.Constructor = &function
					} else {
							contract.Functions = append(contract.Functions, function)
					}
				case "EventDefinition":
            event := ExtractEvent(member)
            contract.Events = append(contract.Events, event)
        case "ModifierDefinition":
            modifier := ExtractModifier(member)
            contract.Modifiers = append(contract.Modifiers, modifier)
        case "StructDefinition":
            strct := ExtractStruct(member)
            contract.Structs = append(contract.Structs, strct)
        case "EnumDefinition":
            enum := ExtractEnum(member)
            contract.Enums = append(contract.Enums, enum)
        }
    }
}

// ExtractVariable extracts a variable declaration.
func ExtractVariable(node ASTNode) Variable {
	variable := Variable{
			Name:             node.Name,
			Type:             extractTypeName(node.TypeName),
			Visibility:       node.Visibility,
			StateVariable:    node.StateVariable,
			StorageLocation:  node.StorageLocation,
			Constant:         node.Constant,
			Mutability:       node.Mutability, // For 'immutable' variables
			FunctionSelector: node.FunctionSelector,
	}
	// Extract initial value if available
	if node.Value != nil {
			variable.Value = extractValue(node.Value)
	}
	return variable
}

// ExtractFunction extracts a function definition.
func ExtractFunction(node ASTNode) Function {
    function := Function{
        Name:            node.Name,
				Kind:            node.Kind,
        Visibility:      node.Visibility,
        StateMutability: node.StateMutability,
        Modifiers:       ExtractModifiers(node),
        BaseFunctions:   node.BaseFunctions,
    }
    // Handle overrides
    if node.Overrides != nil {
        for _, override := range node.Overrides.Overrides {
            function.Overrides = append(function.Overrides, override.Name)
        }
    }
    // Parameters
    if node.Parameters != nil {
        for _, paramNode := range node.Parameters.Parameters {
            param := ExtractParameter(paramNode)
            function.Parameters = append(function.Parameters, param)
        }
    }
    // Return Parameters
    if node.ReturnParameters != nil {
        for _, paramNode := range node.ReturnParameters.Parameters {
            param := ExtractParameter(paramNode)
            function.ReturnParameters = append(function.ReturnParameters, param)
        }
    }
    return function
}

func extractValueFromNode(node *ASTNode) string {
	if node == nil {
			return ""
	}
	switch node.NodeType {
	case "Literal":
			if node.Value != nil {
					return fmt.Sprintf("%v", node.Value)
			}
			if node.HexValue != "" {
					return node.HexValue
			}
	case "FunctionCall":
			return extractFunctionCall(node)
	case "Identifier":
			return node.Name
	case "UnaryOperation":
			operand := extractValue(node.SubExpression)
			return fmt.Sprintf("%s%s", node.Operator, operand)
	case "BinaryOperation":
			left := extractValue(node.LeftExpression)
			right := extractValue(node.RightExpression)
			return fmt.Sprintf("(%s %s %s)", left, node.Operator, right)
	default:
			fmt.Printf("Unhandled node type in value extraction: %s\n", node.NodeType)
			return ""
	}
	return ""
}



// ExtractEvent extracts an event definition.
func ExtractEvent(node ASTNode) Event {
	event := Event{
			Name: node.Name,
	}
	// Parameters
	if node.Parameters != nil {
			for _, paramNode := range node.Parameters.Parameters {
					param := ExtractParameter(paramNode)
					event.Parameters = append(event.Parameters, param)
			}
	}
	return event
}

// ExtractModifier extracts a function modifier.
func ExtractModifier(node ASTNode) Modifier {
    modifier := Modifier{
        Name: node.Name,
    }
    // Parameters
    if node.Parameters != nil {
        for _, paramNode := range node.Parameters.Parameters {
            param := ExtractParameter(paramNode)
            modifier.Parameters = append(modifier.Parameters, param)
        }
    }
    return modifier
}

// ExtractStruct extracts a struct definition.
func ExtractStruct(node ASTNode) Struct {
    s := Struct{
        Name: node.Name,
    }
    for _, member := range node.Members {
        variable := ExtractVariable(member)
        s.Members = append(s.Members, variable)
    }
    return s
}

// ExtractEnum extracts an enum definition.
func ExtractEnum(node ASTNode) Enum {
    enum := Enum{
        Name: node.Name,
    }
    for _, member := range node.Members {
        if member.NodeType == "EnumValue" {
            enum.Values = append(enum.Values, member.Name)
        }
    }
    return enum
}

// ExtractParameter extracts a parameter from a VariableDeclaration node.
func ExtractParameter(node ASTNode) Parameter {
	param := Parameter{
			Name: node.Name,
			Type: extractTypeName(node.TypeName),
	}

	// Check if 'Indexed' is set (only relevant for event parameters)
	if node.Indexed != nil {
			param.Indexed = *node.Indexed
	} else {
			// Default value, assuming 'indexed' is false if not specified
			param.Indexed = false
	}

	return param
}

// ExtractModifiers extracts modifiers applied to a function.
func ExtractModifiers(node ASTNode) []string {
    var modifiers []string
    for _, mod := range node.Modifiers {
        modifiers = append(modifiers, mod.ModifierName.Name)
    }
    return modifiers
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
			if typeName.TypeDescriptions != nil && typeName.TypeDescriptions.TypeString != "" {
					return typeName.TypeDescriptions.TypeString
			} else if typeName.Name != "" {
					return typeName.Name
			} else if typeName.PathNode != nil && typeName.PathNode.Name != "" {
					return typeName.PathNode.Name
			}
			return ""
	case "Mapping":
			// Handle mapping types
			keyType := extractTypeName(typeName.KeyType)
			valueType := extractTypeName(typeName.ValueType)
			return fmt.Sprintf("mapping(%s => %s)", keyType, valueType)
	case "ArrayTypeName":
			// Handle array types
			baseType := extractTypeName(typeName.BaseType)
			if typeName.Length != nil {
					return fmt.Sprintf("%s[%v]", baseType, typeName.Length)
			}
			return fmt.Sprintf("%s[]", baseType)
	default:
			return ""
	}
}

// extractValue extracts the value from an ASTNode representing a value.
func extractValue(value interface{}) string {
	if value == nil {
			return ""
	}

	switch v := value.(type) {
	case string:
			return v
	case float64, int, bool:
			return fmt.Sprintf("%v", v)
	case map[string]interface{}:
			// This is likely an ASTNode represented as a map
			nodeData, err := json.Marshal(v)
			if err != nil {
					fmt.Printf("Error marshaling node: %v\n", err)
					return ""
			}
			node := &ASTNode{}
			err = json.Unmarshal(nodeData, node)
			if err != nil {
					fmt.Printf("Error unmarshaling node: %v\n", err)
					return ""
			}
			return extractValueFromNode(node)
	default:
			fmt.Printf("Unhandled value type: %T\n", v)
			return ""
	}
}


// extractFunctionCall extracts information from a FunctionCall node used as a value.
func extractFunctionCall(node *ASTNode) string {
	if node == nil || node.Expression == nil {
			return ""
	}
	functionName := ""
	if node.Expression.NodeType == "Identifier" {
			functionName = node.Expression.Name
	} else {
			functionName = extractValue(node.Expression)
	}
	args := []string{}
	for _, arg := range node.Arguments {
			argValue := extractValue(arg)
			args = append(args, argValue)
	}
	return fmt.Sprintf("%s(%s)", functionName, strings.Join(args, ", "))
}
