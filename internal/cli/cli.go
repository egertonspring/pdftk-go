// Package cli provides the command line interface for pdftk-go
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdftk-go/internal/pdf"
)

var (
	verbose    bool
	dontAsk    bool
	doAsk      bool
	outputUTF8 bool
)

// Execute runs the root command
func Execute(version, commit, date string) error {
	// Get command line arguments
	args := os.Args[1:] // Skip program name

	// If no arguments, show help
	if len(args) == 0 {
		showHelp(version, commit, date)
		return nil
	}

	// Parse traditional pdftk-style command line
	return parseTraditionalSyntax(args)
}

func showHelp(version, commit, date string) {
	fmt.Printf("pdftk-go version %s (commit: %s, built: %s)\n", version, commit, date)
	fmt.Println("Basic PDF manipulation toolkit - Go port of PDFtk")
	fmt.Println()
	fmt.Println("Usage: pdftk-go [input_files...] [operation] [output file] [options...]")
	fmt.Println()
	fmt.Println("Operations:")
	fmt.Println("  cat         Concatenate PDF files")
	fmt.Println("  burst       Split PDF into individual pages")
	fmt.Println("  dump_data   Extract PDF metadata")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  pdftk-go file1.pdf file2.pdf cat output combined.pdf")
	fmt.Printf("  pdftk-go input.pdf burst output page_%%02d.pdf\n")
	fmt.Println("  pdftk-go input.pdf dump_data output metadata.txt")
}

// parseTraditionalSyntax handles the traditional pdftk command line syntax
// Example: pdftk input1.pdf input2.pdf cat output combined.pdf
func parseTraditionalSyntax(args []string) error {
	fmt.Printf("DEBUG: Parsing arguments: %v\n", args)

	if len(args) < 2 {
		return fmt.Errorf("insufficient arguments")
	}

	// Find operation keyword
	operationIndex := -1
	operations := []string{"cat", "shuffle", "burst", "dump_data", "dump_data_utf8",
		"dump_data_fields", "dump_data_annots", "generate_fdf", "unpack_files",
		"attach_files", "update_info", "update_info_utf8", "fill_form"}

	for i, arg := range args {
		for _, op := range operations {
			if strings.ToLower(arg) == op {
				operationIndex = i
				break
			}
		}
		if operationIndex != -1 {
			break
		}
	}

	if operationIndex == -1 {
		return fmt.Errorf("no valid operation found in: %v", args)
	}

	inputFiles := args[:operationIndex]
	operation := args[operationIndex]
	remainingArgs := args[operationIndex+1:]

	fmt.Printf("DEBUG: Found operation '%s' at index %d\n", operation, operationIndex)
	fmt.Printf("DEBUG: Input files: %v\n", inputFiles)
	fmt.Printf("DEBUG: Remaining args: %v\n", remainingArgs)

	// Handle specific operations
	switch strings.ToLower(operation) {
	case "cat", "shuffle":
		return handleCatOperation(inputFiles, operation, remainingArgs)
	case "burst":
		return handleBurstOperation(inputFiles, remainingArgs)
	case "dump_data", "dump_data_utf8":
		return handleDumpDataOperation(inputFiles, operation, remainingArgs)
	default:
		return fmt.Errorf("operation '%s' not yet implemented", operation)
	}
}

func handleCatOperation(inputFiles []string, operation string, args []string) error {
	session := pdf.NewSession()
	session.Operation = operation
	session.Verbose = verbose

	// Add input documents
	for i, file := range inputFiles {
		handle := string(rune('A' + i)) // A, B, C, etc.
		doc := pdf.NewDocument(file, "", handle)
		session.AddDocument(doc)
	}

	// Parse output file from args
	if len(args) >= 2 && args[0] == "output" {
		session.OutputFile = args[1]
	} else {
		return fmt.Errorf("expected 'output filename' after operation")
	}

	return session.Execute()
}

func handleBurstOperation(inputFiles []string, args []string) error {
	if len(inputFiles) != 1 {
		return fmt.Errorf("burst operation requires exactly one input file")
	}

	session := pdf.NewSession()
	session.Operation = "burst"
	session.Verbose = true // Always verbose for debugging

	// Add single input document
	doc := pdf.NewDocument(inputFiles[0], "", "A")
	session.AddDocument(doc)

	// Parse output pattern from args
	if len(args) >= 2 && args[0] == "output" {
		session.OutputFile = args[1]
	} else {
		// Default pattern
		inputFile := inputFiles[0]
		base := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
		session.OutputFile = fmt.Sprintf("%s_page_%%04d.pdf", base)
		fmt.Printf("Using default output pattern: %s\n", session.OutputFile)
	}

	fmt.Printf("Executing burst operation on: %s\n", inputFiles[0])
	return session.Execute()
}

func handleDumpDataOperation(inputFiles []string, operation string, args []string) error {
	if len(inputFiles) != 1 {
		return fmt.Errorf("dump_data operation requires exactly one input file")
	}

	session := pdf.NewSession()
	session.Operation = operation
	session.Verbose = verbose

	// Add single input document
	doc := pdf.NewDocument(inputFiles[0], "", "A")
	session.AddDocument(doc)

	// Parse output file from args (optional for dump_data)
	if len(args) >= 2 && args[0] == "output" {
		session.OutputFile = args[1]
	} else {
		session.OutputFile = "" // Will output to stdout
	}

	return session.Execute()
}
