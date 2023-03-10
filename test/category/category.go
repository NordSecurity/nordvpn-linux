package category

import (
	"flag"
	"log"
	"strings"
	"testing"

	"golang.org/x/exp/slices"
)

// Category defines a test category
type Category string

const (
	// Unit is a test category for unit tests
	Unit Category = "unit"
	// Integration is a test category for integration tests
	Integration Category = "integration"
	// Firewall is a test category for firewall tests
	Firewall Category = "firewall"
	// File is a test category for tests that manipulate files
	File Category = "file"
	// Link is a test category for tests that manipulate network interfaces
	Link Category = "link"
	// Route is a test category for tests that manipulate network routes
	Route Category = "route"
	// Root is a test category for tests that require root privileges
	Root Category = "root"
)

// Define known categories list in order to determine if
var known = []Category{
	Unit,
	Integration,
	Firewall,
	File,
	Link,
	Route,
	Root,
}

// this must be done in a global scope before any executions because these are custom flags
var (
	selected    []Category
	and         bool
	excluded    []Category
	excludedAnd bool
)

func init() {
	testing.Init()
	// Get flags input
	// Get list of categories for execution
	fCategories := flag.String("categories", "", "comma separated list of categories.")
	fExcluded := flag.String("exclude", "", "comma separated list of excluded categories.")
	// OR - if test hs at least one of specified categories, it will be executed
	// AND - if test has ALL of specified categories, it will be executed
	flag.BoolVar(&and, "and", false,
		"if this is true, AND operator will be used in order to select tests.")
	flag.BoolVar(&excludedAnd, "exand", false,
		"if this is true, AND operator will be used to exclude tests.")
	flag.Parse()

	listSelected := strings.Split(*fCategories, ",")
	listExcluded := strings.Split(*fExcluded, ",")
	// Null checks are not necessary here because there are default values
	if *fCategories != "" {
		checkCategoryList(listSelected)
		for _, c := range listSelected {
			selected = append(selected, Category(c))
		}
	}
	if *fExcluded != "" {
		checkCategoryList(listExcluded)
		for _, c := range listExcluded {
			excluded = append(excluded, Category(c))
		}
	}
}

// Set sets a set of categories for a test. If execution flags match the categories, test will be executed
func Set(t *testing.T, categories ...Category) {
	if len(selected) != 0 && !containsList(categories, selected, and) ||
		len(excluded) != 0 && containsList(categories, excluded, excludedAnd) {
		t.Skip()
	}
}

func containsList(l1 []Category, l2 []Category, all bool) bool {
	for _, c := range l2 {
		if slices.Contains(l1, c) != all {
			return !all
		}
	}
	return all
}

func checkCategoryList(list []string) {
	for _, c := range list {
		if !isKnownCategory(c) {
			log.Fatalf("unknown category %s. Known categories: %+v", c, known)
		}
	}
}

func isKnownCategory(category string) bool {
	for _, c := range known {
		if string(c) == category {
			return true
		}
	}
	return false
}
