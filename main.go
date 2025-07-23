package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rsvihladremio/threaded-top-reporter/parser" // Import the parser package
	"github.com/rsvihladremio/threaded-top-reporter/reporter"
)

var (
	outputFile  string
	reportTitle string
	metadata    string
	Version     string = "dev" // overridden via -ldflags "-X main.Version=…"
)

func init() {
	flag.StringVar(&outputFile, "o", "ttop.html", "Output HTML file path")
	flag.StringVar(&reportTitle, "n", "Threaded Top Report", "Report title")
	flag.StringVar(&metadata, "m", "", "Additional metadata as JSON string")
	showVersion := flag.Bool("version", false, "show version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Println(Version)
		os.Exit(0)
	}
}

func main() {

	// Validate input file
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Please provide an input file")
	}
	inputFile := args[0]

	// Sanitize and read input file
	cleanInput := filepath.Clean(inputFile)
	if strings.Contains(cleanInput, "..") {
		log.Fatalf("invalid input path: %s", inputFile)
	}
	data, err := os.ReadFile(cleanInput)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	// Parse input using the new package
	parsedData, err := parser.ParseTopOutput(data)
	if err != nil {
		log.Fatalf("Error parsing top output: %v", err)
	}

	// Generate report
	fileName := filepath.Base(cleanInput)
	fileHash := fmt.Sprintf("%x", sha256.Sum256(data))
	if err := reporter.GenerateReport(parsedData, outputFile, reportTitle, metadata, fileName, fileHash); err != nil {
		log.Fatalf("Error generating report: %v", err)
	}

	fmt.Printf("report '%s' written to %s\n", reportTitle, outputFile)
}
