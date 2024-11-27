/***********************************************************************************
Author: Caleb Princewill Nwokocha
***********************************************************************************/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Clause []int // A clause is a slice of integers representing literals
type CNF []Clause // CNF is a conjunction of clauses

type FormulaStore struct {
	Formulas    map[string]string          `json:"formulas"`
	Assignments map[string]map[string]bool `json:"assignments"`
}

func loadStore(filename string) (*FormulaStore, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return &FormulaStore{
				Formulas:    make(map[string]string),
				Assignments: make(map[string]map[string]bool),
			}, nil
		}
		return nil, err
	}
	defer file.Close()
	var store FormulaStore
	err = json.NewDecoder(file).Decode(&store)
	return &store, err
}

func saveStore(filename string, store *FormulaStore) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(store)
}

// UnitPropagation simplifies the CNF by assigning values for unit clauses
func UnitPropagation(cnf CNF, assignment map[int]bool) (CNF, bool) {
	for {
		unitFound := false
		for _, clause := range cnf {
			if len(clause) == 1 { // Found a unit clause
				unit := clause[0]
				unitFound = true
				value := unit > 0
				variable := abs(unit)
				assignment[variable] = value
				cnf = assign(cnf, variable, value)
				break
			}
		}
		if !unitFound {
			break
		}
	}
	for _, clause := range cnf {
		if len(clause) == 0 {
			return cnf, false // Conflict detected
		}
	}
	return cnf, true
}

// PureLiteralElimination simplifies CNF by assigning values for pure literals
func PureLiteralElimination(cnf CNF, assignment map[int]bool) CNF {
	literalCount := make(map[int]int)
	for _, clause := range cnf {
		for _, literal := range clause {
			literalCount[literal]++
		}
	}
	for literal, count := range literalCount {
		if count > 0 && literalCount[-literal] == 0 { // Pure literal found
			value := literal > 0
			variable := abs(literal)
			assignment[variable] = value
			cnf = assign(cnf, variable, value)
		}
	}
	return cnf
}

// Assign simplifies the CNF given a variable assignment
func assign(cnf CNF, variable int, value bool) CNF {
	newCNF := CNF{}
	for _, clause := range cnf {
		newClause := Clause{}
		skipClause := false
		for _, literal := range clause {
			if literal == variable && value || literal == -variable && !value {
				skipClause = true
				break
			} else if literal != variable && literal != -variable {
				newClause = append(newClause, literal)
			}
		}
		if !skipClause {
			newCNF = append(newCNF, newClause)
		}
	}
	return newCNF
}

// DPLL implements the main algorithm
func DPLL(cnf CNF, assignment map[int]bool) (bool, map[int]bool) {
	// Apply unit propagation
	cnf, ok := UnitPropagation(cnf, assignment)
	if !ok {
		return false, assignment // Conflict detected
	}

	// Apply pure literal elimination
	cnf = PureLiteralElimination(cnf, assignment)

	// Check if all clauses are satisfied
	if len(cnf) == 0 {
		return true, assignment // Satisfiable
	}

	// Select the next variable to assign (heuristic: first literal in the first clause)
	var variable int
	for _, clause := range cnf {
		if len(clause) > 0 {
			variable = abs(clause[0])
			break
		}
	}

	// Try assigning true
	assignment[variable] = true
	if satisfiable, result := DPLL(assign(cnf, variable, true), assignment); satisfiable {
		return true, result
	}

	// Backtrack and try assigning false
	assignment[variable] = false
	return DPLL(assign(cnf, variable, false), assignment)
}

// Helper function: absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Parse CNF string and handle stored formulas
func parseCNF(input string, varMap map[string]int, store *FormulaStore) (CNF, map[int]string) {
	reverseMap := make(map[int]string)
	counter := 1
	assignVar := func(v string) int {
		if _, ok := varMap[v]; !ok {
			varMap[v] = counter
			reverseMap[counter] = v
			counter++
		}
		return varMap[v]
	}

	var cnf CNF
	input = strings.ReplaceAll(input, " ", "")
	clauses := strings.Split(input, "/\\")
	for _, clauseStr := range clauses {
		var clause Clause
		literals := strings.Split(strings.Trim(clauseStr, "()"), "\\/")
		for _, litStr := range literals {
			isNeg := strings.HasPrefix(litStr, "!")
			litStr = strings.Trim(litStr, "!")

			if storedFormula, exists := store.Formulas[litStr]; exists {
				// Handle stored formula as a literal
				subCNF, _ := parseCNF(storedFormula, varMap, store)
				// If negation, negate the subformula's clauses
				if isNeg {
					for _, subClause := range subCNF {
						for i := range subClause {
							subClause[i] = -subClause[i]
						}
						cnf = append(cnf, subClause)
					}
				} else {
					// If not negated, add the subformula's clauses
					cnf = append(cnf, subCNF...)
				}
			} else {
				// Handle regular variable literals
				lit := assignVar(litStr)
				if isNeg {
					lit = -lit
				}
				clause = append(clause, lit)
			}
		}
		if len(clause) > 0 {
			cnf = append(cnf, clause)
		}
	}
	return cnf, reverseMap
}

func main() {
	const storeFile = "formulas.json"
	store, err := loadStore(storeFile)
	if err != nil {
		fmt.Println("Error loading store:", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Satisfier checks satisfiability of a formula in Conjuctive Norm Form (CNF).")

	for {
		fmt.Print("Enter 1 to check a new formula, 2 to view stored formulas, or 3 to exit: ")
		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		switch option {
		case "1":
			fmt.Print("Enter formula name: ")
			formulaName, _ := reader.ReadString('\n')
			formulaName = strings.TrimSpace(formulaName)

			fmt.Print("Enter CNF of " + formulaName + " e.g., (\"Rainy\" \\/ \"Sunny\") " +
				"/\\ (\"Cold\" \\/ !\"Hot\") /\\ (\"Windy\" \\/ !\"Clear\"):\n")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			varMap := make(map[string]int)
			cnf, reverseMap := parseCNF(input, varMap, store)
			assignment := make(map[int]bool)

			if satisfiable, result := DPLL(cnf, assignment); satisfiable {
				fmt.Println("SATISFIABLE")
				store.Formulas[formulaName] = input
				store.Assignments[formulaName] = map[string]bool{}
				for lit, val := range result {
					if val {
						store.Assignments[formulaName][reverseMap[abs(lit)]] = true
					} else {
						store.Assignments[formulaName][reverseMap[abs(lit)]] = false
					}
				}
				for lit, val := range store.Assignments[formulaName] {
					fmt.Printf("%s : %v\n", lit, val)
				}
			} else {
				fmt.Println("UNSATISFIABLE")
			}
			saveStore(storeFile, store)

		case "2":
			fmt.Println("Stored formulas:")
			if len(store.Formulas) == 0 {
				fmt.Println("No formula stored")
			} else {
				for name, formula := range store.Formulas {
					fmt.Printf("%s = %s\n", name, formula)
					assignment := store.Assignments[name]
					for lit, val := range assignment {
						fmt.Printf("  %s = %v\n", lit, val)
					}
				}
			}

		case "3":
			fmt.Println("Exiting.")
			return

		default:
			fmt.Println("Invalid option.")
		}
	}
}
