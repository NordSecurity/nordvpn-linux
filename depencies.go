package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Function to execute a shell command and get the output
func execCommand(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)
	output, err := command.Output()
	return string(output), err
}

// Function to build the dependency graph using `go list`
func buildDependencyGraph(pkg string) (map[string][]string, error) {
	graph := make(map[string][]string)
	output, err := execCommand("go", "list", "-json", pkg)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	var currentPackage string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Path:") {
			currentPackage = strings.TrimPrefix(line, "Path: ")
		}
		if strings.HasPrefix(line, "Deps:") {
			deps := strings.TrimPrefix(line, "Deps: [")
			deps = strings.TrimSuffix(deps, "]")
			if currentPackage != "" {
				graph[currentPackage] = strings.Fields(deps)
			}
		}
	}
	return graph, nil
}

// Function to detect cycles using DFS
func detectCycles(graph map[string][]string) [][]string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycles [][]string

	var dfs func(string, []string) bool
	dfs = func(node string, path []string) bool {
		if recStack[node] {
			cycles = append(cycles, append(path, node))
			return true
		}
		if visited[node] {
			return false
		}
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range graph[node] {
			if dfs(neighbor, append(path, node)) {
				break
			}
		}
		recStack[node] = false
		return false
	}

	for node := range graph {
		if !visited[node] {
			dfs(node, []string{})
		}
	}
	return cycles
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <package>")
		return
	}
	pkg := os.Args[1]

	graph, err := buildDependencyGraph(pkg)
	if err != nil {
		fmt.Printf("Error building dependency graph: %v\n", err)
		return
	}

	cycles := detectCycles(graph)
	if len(cycles) == 0 {
		fmt.Println("No cycles found.")
	} else {
		fmt.Println("Cycles found:")
		for _, cycle := range cycles {
			fmt.Println(strings.Join(cycle, " -> "))
		}
	}
}
