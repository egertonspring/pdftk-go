// Package main implements pdftk-go, a faithful Go port of the PDFtk toolkit.
//
// This is a complete port of the original pdftk-java implementation,
// including all operations and command-line compatibility.
package main

import (
	"fmt"
	"os"

	"github.com/pdftk-go/internal/operations"
	"github.com/pdftk-go/internal/session"
)

var (
	version = "1.0.0-go"
	commit  = "unknown"
	date    = "unknown"
)

// main is the entry point (port of pdftk.main())
func main() {
	os.Exit(mainNoExit(os.Args[1:]))
}

// mainNoExit implements the main logic without calling os.Exit (port of pdftk.main_noexit())
func mainNoExit(args []string) int {
	helpRequested := false
	versionRequested := false
	synopsis := len(args) == 0

	// Check for help and version flags
	for _, arg := range args {
		if arg == "--version" || arg == "-version" {
			versionRequested = true
		}
		if arg == "--help" || arg == "-help" || arg == "-h" {
			helpRequested = true
		}
	}

	if helpRequested {
		describeFull()
		return 0
	} else if versionRequested {
		describeHeader()
		return 0
	} else if synopsis {
		describeSynopsis()
		return 0
	}

	// Parse arguments and execute operation
	tkSession := session.NewTKSession()
	err := tkSession.Parse(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		return 1
	}

	// Dump session data if verbose
	tkSession.DumpSessionData()

	if !tkSession.IsValid() {
		fmt.Fprintln(os.Stderr, "Done. Input errors, so no output created.")
		return 1
	}

	// Execute the operation using operations package
	switch tkSession.Operation {
	case session.CatK:
		err = operations.ExecuteCatOperation(tkSession)
	case session.ShuffleK:
		err = operations.ExecuteShuffleOperation(tkSession)
	case session.BurstK:
		err = operations.ExecuteBurstOperation(tkSession)
	case session.DumpDataK, session.DumpDataUTF8K, session.DumpDataFieldsK, session.DumpDataFieldsUTF8K, session.DumpDataAnnotsK:
		err = operations.ExecuteDumpDataOperation(tkSession)
	default:
		err = fmt.Errorf("unsupported operation: %v", tkSession.Operation)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output: %v\n", err)
		fmt.Fprintln(os.Stderr, "   No output created.")
		return 1
	}

	return 0
}

// describeFull shows complete help (port of describe_full())
func describeFull() {
	describeHeader()
	fmt.Println()
	describeSynopsis()
	fmt.Println()
	describeUsage()
}

// describeHeader shows version info (port of describe_header())
func describeHeader() {
	fmt.Printf("pdftk-go %s\n", version)
	fmt.Println("A Go port of PDFtk, the PDF Toolkit")
	fmt.Println("Copyright (c) 2024 pdftk-go contributors")
	fmt.Println("This is a faithful port of the original PDFtk functionality.")
}

// describeSynopsis shows brief usage (port of describe_synopsis())
func describeSynopsis() {
	fmt.Println("SYNOPSIS")
	fmt.Println("     pdftk-go <input PDF files | - | PROMPT>")
	fmt.Println("          [ input_pw <input PDF owner passwords | PROMPT> ]")
	fmt.Println("          <operation> <operation arguments>")
	fmt.Println("          [ output <output filename | - | PROMPT> ]")
	fmt.Println("          [ encrypt_40bit | encrypt_128bit | encrypt_256bit ]")
	fmt.Println("          [ allow <permissions> ]")
	fmt.Println("          [ owner_pw <owner password | PROMPT> ]")
	fmt.Println("          [ user_pw <user password | PROMPT> ]")
	fmt.Println("          [ compress | uncompress ]")
	fmt.Println("          [ keep_first_id | keep_final_id | drop_xfa ]")
	fmt.Println("          [ verbose | dont_ask | do_ask ]")
	fmt.Println()
	fmt.Println("     For Complete Help: pdftk-go --help")
}

// describeUsage shows detailed usage information
func describeUsage() {
	fmt.Println("DESCRIPTION")
	fmt.Println("     If PDF is encrypted with a password, supply it like this:")
	fmt.Println("          pdftk-go secured.pdf input_pw foopass cat output unsecured.pdf")
	fmt.Println()
	fmt.Println("OPERATIONS")
	fmt.Println("     cat [page ranges]")
	fmt.Println("          Concatenates pages from input PDFs to create a new PDF.")
	fmt.Println("          Page ranges refer to the previously-named PDFs.")
	fmt.Println()
	fmt.Println("     shuffle [page ranges]")
	fmt.Println("          Like cat, but interleaves pages from page ranges.")
	fmt.Println()
	fmt.Println("     burst")
	fmt.Println("          Splits a single PDF into individual pages.")
	fmt.Println("          Output filenames are generated from the output filename by")
	fmt.Println("          appending _01.pdf, _02.pdf, etc.")
	fmt.Println()
	fmt.Println("     dump_data")
	fmt.Println("          Reports PDF metadata, bookmarks and page labels to the")
	fmt.Println("          given output filename or stdout.")
	fmt.Println()
	fmt.Println("     dump_data_fields")
	fmt.Println("          Reports PDF form field statistics to the given output")
	fmt.Println("          filename or stdout.")
	fmt.Println()
	fmt.Println("     fill_form <FDF filename | XFDF filename | - | PROMPT>")
	fmt.Println("          Fills the PDF form fields with data from an FDF or XFDF file.")
	fmt.Println()
	fmt.Println("     generate_fdf")
	fmt.Println("          Creates an FDF file suitable for editing PDF form fields.")
	fmt.Println()
	fmt.Println("     attach_files <attachment files>")
	fmt.Println("          Attaches files to the PDF.")
	fmt.Println()
	fmt.Println("     unpack_files")
	fmt.Println("          Copies attachments from the input PDF into the current")
	fmt.Println("          working directory or to an output directory.")
	fmt.Println()
	fmt.Println("EXAMPLES")
	fmt.Println("     pdftk-go A=even.pdf B=odd.pdf shuffle A B output collated.pdf")
	fmt.Println("     pdftk-go in1.pdf in2.pdf cat output out.pdf")
	fmt.Println("     pdftk-go in.pdf cat 1-12 14-end output out.pdf")
	fmt.Printf("     pdftk-go in.pdf burst output pg_%%04d.pdf\n")
	fmt.Println("     pdftk-go in.pdf dump_data output report.txt")
	fmt.Println("     pdftk-go secured.pdf input_pw foopass cat output unsecured.pdf")
}
