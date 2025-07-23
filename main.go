package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rsvihladremio/threaded-top-reporter/parser"   // Import the parser package
	"github.com/rsvihladremio/threaded-top-reporter/reporter"
)

func main() {
	// Define command-line flags
	outputFile := flag.String("o", "ttop.html", "Output HTML file path")
	reportTitle := flag.String("n", "Threaded Top Report", "Report title")
	metadata := flag.String("m", "", "Additional metadata as JSON string")

	flag.Parse()

	// Validate input file
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Please provide an input file")
	}
	inputFile := args[0]

	// Read input file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	// Parse input using the new package
	parsedData, err := parser.ParseTopOutput(data)
	if err != nil {
		log.Fatalf("Error parsing top output: %v", err)
	}

	// Generate report
	if err := reporter.GenerateReport(parsedData, *outputFile, *reportTitle, *metadata); err != nil {
		log.Fatalf("Error generating report: %v", err)
	}

	fmt.Printf("report '%s' written to %s\n", *reportTitle, *outputFile)
}

