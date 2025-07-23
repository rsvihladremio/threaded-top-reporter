package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rsvihladremio/threaded-top-reporter/parser" // Import the parser package
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
	if err := generateReport(parsedData, *outputFile, *reportTitle, *metadata); err != nil {
		log.Fatalf("Error generating report: %v", err)
	}

	fmt.Printf("report '%s' written to %s\n", *reportTitle, *outputFile)
}

func generateReport(data parser.ReportData, outputPath, title, metadata string) error {
	// TODO: Implement actual HTML generation using the parsed data
	fmt.Printf("Generating report with title: %s, metadata: %s\n", title, metadata)
	// Example: Iterate through snapshots and then processes within each snapshot
	totalProcesses := 0
	for i, snapshot := range data.Snapshots {
		totalProcesses += len(snapshot.Processes)
		fmt.Printf("Snapshot %d: Number of processes: %d\n", i+1, len(snapshot.Processes))
		for _, process := range snapshot.Processes {
			fmt.Printf("  PID: %d, User: %s, CPU: %.2f, MEM: %.2f, Command: %s\n", process.PID, process.User, process.CPU, process.MEM, process.Command)
		}
	}
	fmt.Printf("Total processes across all snapshots: %d\n", totalProcesses)
	return nil
}
