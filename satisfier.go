package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Clause []int
type CNF []Clause

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

func UnitPropagation(cnf CNF, assignment map[int]bool) (CNF, bool) {
	for {
		unitFound := false
		for _, clause := range cnf {
			if len(clause) == 1 {
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
			return cnf, false
		}
	}
	return cnf, true
}

func PureLiteralElimination(cnf CNF, assignment map[int]bool) CNF {
	literalCount := make(map[int]int)
	for _, clause := range cnf {
		for _, literal := range clause {
			literalCount[literal]++
		}
	}
	for literal, count := range literalCount {
		if count > 0 && literalCount[-literal] == 0 {
			value := literal > 0
			variable := abs(literal)
			assignment[variable] = value
			cnf = assign(cnf, variable, value)
		}
	}
	return cnf
}

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

func iterativeDPLL(cnf CNF, assignment map[int]bool) (bool, map[int]bool) {
	stack := []struct {
		cnf        CNF
		assignment map[int]bool
	}{
		{cnf, copyAssignment(assignment)},
	}

	for len(stack) > 0 {
		state := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		var ok bool
		state.cnf, ok = UnitPropagation(state.cnf, state.assignment)
		if !ok {
			continue
		}

		state.cnf = PureLiteralElimination(state.cnf, state.assignment)

		if len(state.cnf) == 0 {
			return true, state.assignment
		}

		var variable int
		for _, clause := range state.cnf {
			if len(clause) > 0 {
				variable = abs(clause[0])
				break
			}
		}

		stack = append(stack, struct {
			cnf        CNF
			assignment map[int]bool
		}{
			assign(state.cnf, variable, false),
			copyAssignment(state.assignment),
		})

		stack = append(stack, struct {
			cnf        CNF
			assignment map[int]bool
		}{
			assign(state.cnf, variable, true),
			copyAssignment(state.assignment),
		})
	}

	return false, assignment
}

func copyAssignment(original map[int]bool) map[int]bool {
	copy := make(map[int]bool)
	for key, value := range original {
		copy[key] = value
	}
	return copy
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
	stack := []string{}
	currentClause := Clause{}
	input = strings.ReplaceAll(input, " ", "") // Remove spaces

	for i := 0; i < len(input); {
		switch input[i] {
		case '(':
			// Push current clause to stack and start a new one
			stack = append(stack, "(")
			i++
		case ')':
			// Process completed clause
			if len(currentClause) > 0 {
				cnf = append(cnf, currentClause)
				currentClause = Clause{}
			}
			stack = stack[:len(stack)-1] // Pop stack
			i++
		case '/':
			if i+1 < len(input) && input[i+1] == '\\' { // Detect "/\"
				if len(currentClause) > 0 {
					cnf = append(cnf, currentClause)
					currentClause = Clause{}
				}
				i += 2
			} else {
				i++
			}
		case '\\':
			if i+1 < len(input) && input[i+1] == '/' { // Detect "\/"
				i += 2
			} else {
				i++
			}
		case '!':
			// Negate the next literal
			isNeg := true
			i++
			start := i
			for i < len(input) && (input[i] != ')' && input[i] != '\\' && input[i] != '/') {
				i++
			}
			literal := input[start:i]
			if storedFormula, exists := store.Formulas[literal]; exists {
				subCNF, _ := parseCNF(storedFormula, varMap, store)
				for _, subClause := range subCNF {
					for j := range subClause {
						subClause[j] = -subClause[j]
					}
					cnf = append(cnf, subClause)
				}
			} else {
				lit := assignVar(literal)
				if isNeg {
					lit = -lit
				}
				currentClause = append(currentClause, lit)
			}
		default:
			start := i
			for i < len(input) && (input[i] != ')' && input[i] != '\\' && input[i] != '/') {
				i++
			}
			literal := input[start:i]
			if storedFormula, exists := store.Formulas[literal]; exists {
				subCNF, _ := parseCNF(storedFormula, varMap, store)
				cnf = append(cnf, subCNF...)
			} else {
				lit := assignVar(literal)
				currentClause = append(currentClause, lit)
			}
		}
	}

	// Append the last processed clause
	if len(currentClause) > 0 {
		cnf = append(cnf, currentClause)
	}
	return cnf, reverseMap
}

// Apply pre-assigned values to the CNF before DPLL execution
func applyPreAssignments(cnf CNF, assignment map[int]bool) CNF {
	for variable, value := range assignment {
		cnf = assign(cnf, variable, value)
	}
	return cnf
}

// Helper function: check if a variable exists in the CNF formula
func containsVariable(cnf CNF, varName string, reverseMap map[int]string) bool {
	for _, clause := range cnf {
		for _, lit := range clause {
			if reverseMap[abs(lit)] == varName {
				return true
			}
		}
	}
	return false
}

// Main function
func main() {
	const storeFile = "formulas.json"
	store, err := loadStore(storeFile)
	if err != nil {
		fmt.Println("Error loading store:", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Satisfier checks satisfiability of a formula in Conjunctive Normal Form (CNF).")

	for {
		fmt.Print("Enter 1 to check a new formula, 2 to view stored formulas, or 3 to exit: ")
		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		switch option {
		case "1":
			fmt.Print("Enter formula name: ")
			formulaName, _ := reader.ReadString('\n')
			formulaName = strings.TrimSpace(formulaName)

			fmt.Printf("Enter CNF of \"%s\" e.g., (\"R\" \\/ \"S\") /\\ (\"C\" \\/ !\"H\") "+
				"/\\ (\"W\" \\/ !\"C\"):\n", formulaName)

			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			fmt.Print("Enter initial assignments (e.g., \"R\" := true, \"S\" := false) or leave blank: ")
			preAssignmentsInput, _ := reader.ReadString('\n')
			preAssignmentsInput = strings.TrimSpace(preAssignmentsInput)

			varMap := make(map[string]int) // Initialize varMap for variable mapping
			cnf, reverseMap := parseCNF(input, varMap, store)

			preAssignments := make(map[int]bool)
			if preAssignmentsInput != "" {
				assignments := strings.Split(preAssignmentsInput, ",")
				for _, assign := range assignments {
					parts := strings.Split(assign, ":=")
					if len(parts) != 2 {
						fmt.Printf("Invalid assignment format: %s. Expected 'Var := Value'.\n", assign)
						continue
					}
					varName := strings.TrimSpace(parts[0])
					varValue := strings.TrimSpace(parts[1]) == "true"

					// Here we check if the variable is part of the CNF
					if varIndex, exists := varMap[varName]; exists {
						preAssignments[varIndex] = varValue
					} else {
						// Check if the variable is part of any subformula
						if containsVariable(cnf, varName, reverseMap) {
							preAssignments[varMap[varName]] = varValue
						} else {
							fmt.Printf("Warning: Variable %s not found in \"%s\"\n", varName, formulaName)
						}
					}
				}
			}

			// Apply user-provided assignments
			cnf = applyPreAssignments(cnf, preAssignments)
			assignment := preAssignments

			if satisfiable, result := iterativeDPLL(cnf, assignment); satisfiable {
				fmt.Printf("\"%s\" is SATISFIABLE.\n", formulaName)
				store.Formulas[formulaName] = input
				store.Assignments[formulaName] = map[string]bool{}
				for lit, val := range result {
					store.Assignments[formulaName][reverseMap[abs(lit)]] = val
				}
				fmt.Printf("Assignments for \"%s\":\n", formulaName)
				for lit, val := range store.Assignments[formulaName] {
					fmt.Printf("%s : %v\n", lit, val)
				}
			} else {
				fmt.Printf("\"%s\" is UNSATISFIABLE.\n", formulaName)
			}

			err := saveStore(storeFile, store)
			if err != nil {
				fmt.Println("Error saving store:", err)
			}

		case "2":
			fmt.Println("Stored formulas:")
			if len(store.Formulas) == 0 {
				fmt.Println("No formula stored.")
			} else {
				for name, formula := range store.Formulas {
					fmt.Printf("\"%s\" := %s\n", name, formula)
					assignment := store.Assignments[name]
					for lit, val := range assignment {
						fmt.Printf("%s := %v\n", lit, val)
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
