
# Satisfier

## Overview

This project is a **Satisfiability Checker** that works with formulas in **Conjunctive Normal Form (CNF)**. It allows the user to input CNF formulas, stores them, and checks their satisfiability using the **DPLL algorithm** (Davis-Putnam-Logemann-Loveland). The formulas can include both literals and previously stored formulas, which are treated as variables within the CNF. The program supports unit propagation, pure literal elimination, and branching for DPLL.

### Features:
- **Satisfiability Checking**: Check if a CNF formula is satisfiable.
- **Formula Storage**: Store and retrieve previously input formulas.
- **Substitution of Stored Formulas**: Substitute stored formulas as literals within new formulas.
- **DPLL Algorithm**: Uses the DPLL algorithm to check satisfiability, employing unit propagation, pure literal elimination, and recursion.

## Requirements

To run this project, you need:

- Go (version 1.18 or higher) installed on your system.

## Setup

1. Clone this repository to your local machine:
   ```bash
   git clone https://github.com/yourusername/formula-satisfiability-checker.git
   cd formula-satisfiability-checker
   ```

2. Install dependencies (if any).

3. Run the project:
   ```bash
   go run main.go
   ```

## How to Use

When you run the program, it will prompt you with a few options:

1. **Check a New Formula**: You can input a new CNF formula and its name to check its satisfiability.
2. **View Stored Formulas**: View all previously stored formulas and their satisfiability assignments.
3. **Exit**: Exit the program.

### Example:

#### Input 1: Add Formula "R"
```
Enter formula name: R
Enter CNF of R e.g., ("Rainy" \/ "Sunny") /\ ("Cold" \/ !"Hot") /\ ("Windy" \/ !"Clear"):
(!"j" \/ !"y")
```

#### Input 2: Add Formula "I"
```
Enter formula name: I
Enter CNF of I e.g., ("Rainy" \/ "Sunny") /\ ("Cold" \/ !"Hot") /\ ("Windy" \/ !"Clear"):
("R" \/ "j") /\ ("j" \/ "y")
```

#### Input 3: Add Formula "G"
```
Enter formula name: G
Enter CNF of G e.g., ("Rainy" \/ "Sunny") /\ ("Cold" \/ !"Hot") /\ ("Windy" \/ !"Clear"):
("R" \/ "z") /\ ("j" \/ "y")
```

#### Input 4: View Stored Formulas
```
Stored formulas:
R = (!"j" \/ !"y")
  "j" = false
  "y" = false
I = ("R" \/ "j") /\ ("j" \/ "y")
  "R" = true
  "j" = true
G = ("R" \/ "z") /\ ("j" \/ "y")
  "R" = true
  "z" = true
  "j" = true
  "y" = true
```

## Code Explanation

### Main Logic

1. **Formula Parsing**:
   - The program parses CNF formulas as input strings, converting them into a list of clauses represented by integers. 
   - Literals are assigned integer values, and negations are represented as negative integers.

2. **Satisfiability Check (DPLL Algorithm)**:
   - The **DPLL algorithm** checks whether a given CNF formula is satisfiable by recursively assigning values to literals and simplifying the formula. 
   - It uses unit propagation and pure literal elimination to simplify the formula and backtracks if necessary.

3. **Substitution**:
   - When a formula references another stored formula (e.g., `"R"`), the program substitutes the stored formula into the CNF as a literal.

### Core Functions

1. **`loadStore()`**: Loads the stored formulas and assignments from a file.
2. **`saveStore()`**: Saves the formulas and assignments to a file.
3. **`unitPropagation()`**: Performs unit propagation to simplify the CNF.
4. **`pureLiteralElimination()`**: Performs pure literal elimination.
5. **`chooseLiteral()`**: Chooses a literal for branching in the DPLL algorithm.
6. **`DPLL()`**: The main function that checks the satisfiability of a CNF formula using the DPLL algorithm.
7. **`assign()`**: Simplifies the CNF by assigning a value to a literal.

## Unit Tests

### Test Setup

The unit tests are written in Go's testing framework. The test ensures that the substitution of `"R"` into another formula works as expected. 

### Unit Test: Substitution of `"R"`

We will test the substitution behavior for the formula `"R"` in different cases.

```go
package main

import (
	"testing"
)

func TestSubstituteR(t *testing.T) {
	// Define the stored formula "R"
	store := &FormulaStore{
		Formulas: map[string]string{
			"R": "(!"j" \/ !"y")",
		},
		Assignments: make(map[string]map[string]bool),
	}

	// Input formula that references "R"
	cnf := " ("R" \/ "j") /\ ("j" \/ "y") "

	// Parse the CNF string
	varMap := make(map[string]int)
	parsedCNF, reverseMap := parseCNF(cnf, varMap, store)

	// Check if substitution worked correctly
	expectedCNF := CNF{
		{-1, -2, 3}, // (!"j" \/ !"y" \/ "j")
		{3, 4},      // ("j" \/ "y")
	}

	if len(parsedCNF) != len(expectedCNF) {
		t.Fatalf("Expected CNF length %d, got %d", len(expectedCNF), len(parsedCNF))
	}

	for i, clause := range parsedCNF {
		for j, lit := range clause {
			if lit != expectedCNF[i][j] {
				t.Errorf("Expected literal %d, got %d at clause %d", expectedCNF[i][j], lit, i)
			}
		}
	}
}
```

### Explanation:

- **Test Case**: The test checks whether the formula `"R"` gets substituted correctly into a new formula that references it.
- **Input**: The formula `("R" \/ "j") /\ ("j" \/ "y")`.
- **Expected Output**: After substitution, the CNF should correctly reflect the literals from `"R"` (i.e., `(!"j" \/ !"y")`), resulting in a CNF like `[ {-1, -2, 3}, {3, 4} ]`.
- **Test Logic**: The test parses the CNF, performs substitution, and compares the result with the expected CNF.

### Running Tests

To run the tests, use the following command:
```bash
go test -v
```

## Conclusion

This project implements a satisfiability checker for formulas in Conjunctive Normal Form (CNF). The checker uses the DPLL algorithm to determine whether the formula is satisfiable. It supports the substitution of stored formulas as literals and includes unit tests to verify the correctness of the substitution logic.
