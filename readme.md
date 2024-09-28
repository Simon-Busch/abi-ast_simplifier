# Solidity AST Parser and Terminal UI Viewer

Welcome to the Solidity AST Parser and Terminal UI Viewer project! This tool allows you to parse Solidity contract files, extract detailed information from their Abstract Syntax Trees (AST), and interactively explore the contracts' structure and components via a terminal-based user interface.

## Status

This project is still a WIP.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Parsing Contracts](#parsing-contracts)
  - [Navigating the Terminal UI](#navigating-the-terminal-ui)
- [Contributing](#contributing)
- [License](#license)

## Features

- **AST Parsing**: Parse Solidity contract JSON files containing ASTs to extract comprehensive information.
- **Detailed Extraction**: Extract contracts' variables, functions, constructors, events, modifiers, structs, enums, and inheritance information.
- **Interactive Terminal UI**: Navigate through contracts and their components using an intuitive terminal-based interface.
- **Supports Multiple Contracts**: Parse and explore multiple contracts within a specified directory.

## Installation

### Prerequisites

- [Go](https://golang.org/dl/) (version 1.20 or later)
- [Git](https://git-scm.com/downloads)
- Solidity compiler to generate AST JSON files (optional if you already have AST files) -- prefered `foundry`.

### Clone the Repository

```bash
git clone git@github.com:Simon-Busch/solidity-ast-parser.git
cd solidity-ast-parser
```

## Usage

### Parsing Contracts

1. Prepare Contract AST Files: Place your Solidity contract JSON files (containing the AST) into the data/ directory. These JSON files can be generated using the Solidity compiler with the appropriate flags.

Can be done with:
`forge build --ast` for a whole project.

2. Create a `data` folder and paste all desired build folder or all the json inside.

3. Build && run the Application: `make run`

### Navigating the Terminal UI

Contracts List: Upon running the application, you'll see a list of contracts parsed from the data/ directory on the left panel.

- Use the Up (↑) and Down (↓) arrow keys to navigate through the list.
- Press Right (→) to select a contract and view its details.

Details Panel: The middle panel displays the selected contract's components, such as constructor, functions, variables, events, structs, and enums.

- Navigate using the Up (↑) and Down (↓) arrow keys.
- Press Right (→) to view detailed information about a selected item in the right panel.
- Press Left (←) to go back to the contracts list or previous panel.

Information Panel: The right panel shows detailed information about the selected component, including parameters, modifiers, visibility, and state mutability.

Exit: Press q or Ctrl+C to exit the application at any time.

## Contributing

Contributions are welcome! If you'd like to improve this project.

## License

This project is licensed under the MIT License.
