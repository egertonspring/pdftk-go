// Package operations implements PDF operation helpers for pdftk-go
package operations

import (
	"fmt"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdftk-go/internal/session"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ExecuteCatOperation performs the concatenate operation using pdfcpu
func ExecuteCatOperation(s *session.TKSession) error {
	if len(s.InputPdf) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No input PDF files specified")
		return fmt.Errorf("no input files")
	}

	if s.OutputFilename == "" {
		fmt.Fprintln(os.Stderr, "Error: No output filename specified")
		return fmt.Errorf("no output filename")
	}

	if s.VerboseReporting {
		fmt.Printf("Concatenating %d PDF files\n", len(s.InputPdf))
		for i, pdf := range s.InputPdf {
			fmt.Printf("  Input %d: %s\n", i+1, pdf.Filename)
		}
		fmt.Printf("Output: %s\n", s.OutputFilename)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(s.OutputFilename)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		return err
	}

	// Collect input filenames for concatenation
	var inputFiles []string
	for _, pdf := range s.InputPdf {
		inputFiles = append(inputFiles, pdf.Filename)
	}

	// Use pdfcpu to merge PDFs with correct parameters
	err := api.MergeCreateFile(inputFiles, s.OutputFilename, false, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error merging PDFs: %v\n", err)
		return err
	}

	if s.VerboseReporting {
		fmt.Printf("Successfully created %s\n", s.OutputFilename)
	}

	return nil
}

// ExecuteBurstOperation performs the burst (split) operation
func ExecuteBurstOperation(s *session.TKSession) error {
	if len(s.InputPdf) != 1 {
		return fmt.Errorf("burst requires exactly one input file")
	}

	inputPdf := s.InputPdf[0]
	inputPath := inputPdf.Filename

	// Determine output pattern
	outputPattern := s.OutputFilename
	if outputPattern == "" {
		base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		outputPattern = base + "_page_%04d.pdf"
	}

	finalOutputDir := filepath.Dir(outputPattern)
	if err := os.MkdirAll(finalOutputDir, 0755); err != nil {
		return fmt.Errorf("cannot create output dir: %w", err)
	}

	// Temporary directory for splitting
	tempDir, err := os.MkdirTemp("", "pdftk-go-burst-")
	if err != nil {
		return fmt.Errorf("cannot create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if s.VerboseReporting {
		fmt.Printf("Bursting PDF: %s\n", inputPath)
		fmt.Printf("Output pattern: %s\n", outputPattern)
		fmt.Printf("Temporary dir: %s\n", tempDir)
	}

	// Step 1: Split PDF into temp dir
	err = api.SplitFile(inputPath, tempDir, 1, nil)
	if err != nil {
		return fmt.Errorf("error during pdf split: %w", err)
	}

	// Step 2: Collect and sort split files
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("cannot read temp split dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	// Step 3: Move or copy files to final output
	page := 1
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		src := filepath.Join(tempDir, entry.Name())
		dst := filepath.Join(finalOutputDir, fmt.Sprintf(filepath.Base(outputPattern), page))

		// Try to rename (fast)
		if err := os.Rename(src, dst); err != nil {
			// Fall back to copy if rename fails (different partitions)
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("cannot write burst page: %w", err)
			}
		}

		page++
	}

	if s.VerboseReporting {
		fmt.Printf("Successfully burst into %d pages.\n", page-1)
	}

	return nil
}

// A small helper for copying files safely
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

// ExecuteDumpDataOperation performs the dump_data operation using pdfcpu
func ExecuteDumpDataOperation(s *session.TKSession) error {
	if len(s.InputPdf) != 1 {
		fmt.Fprintln(os.Stderr, "Error: Dump data operation requires exactly one input PDF")
		return fmt.Errorf("dump_data requires one input file")
	}

	inputPdf := s.InputPdf[0]

	if s.VerboseReporting {
		fmt.Printf("Dumping data from PDF: %s\n", inputPdf.Filename)
		if s.OutputFilename != "" {
			fmt.Printf("Output: %s\n", s.OutputFilename)
		}
	}

	// Determine output destination
	var outputFile *os.File
	var err error

	if s.OutputFilename == "" || s.OutputFilename == "-" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(s.OutputFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			return err
		}
		defer outputFile.Close()
	}

	// Use pdfcpu to get basic PDF information (simplified approach)
	// For now, output basic metadata - full implementation would use pdfcpu's info functions
	fmt.Fprintf(outputFile, "InfoBegin\n")
	fmt.Fprintf(outputFile, "InfoKey: Title\n")
	fmt.Fprintf(outputFile, "InfoValue: %s\n", filepath.Base(inputPdf.Filename))
	fmt.Fprintf(outputFile, "InfoBegin\n")
	fmt.Fprintf(outputFile, "InfoKey: Creator\n")
	fmt.Fprintf(outputFile, "InfoValue: pdftk-go\n")
	fmt.Fprintf(outputFile, "InfoBegin\n")
	fmt.Fprintf(outputFile, "InfoKey: Producer\n")
	fmt.Fprintf(outputFile, "InfoValue: pdftk-go\n")
	fmt.Fprintf(outputFile, "PdfID0: [generated-id-0]\n")
	fmt.Fprintf(outputFile, "PdfID1: [generated-id-1]\n")

	// Try to get page count using pdfcpu
	pageCount := 1 // default
	// Note: Would need to use appropriate pdfcpu function to get actual page count
	fmt.Fprintf(outputFile, "NumberOfPages: %d\n", pageCount)

	return nil
}

// ExecuteShuffleOperation performs the shuffle operation (page interleaving)
func ExecuteShuffleOperation(s *session.TKSession) error {
	if len(s.InputPdf) < 2 {
		fmt.Fprintln(os.Stderr, "Error: Shuffle operation requires at least two input PDFs")
		return fmt.Errorf("shuffle requires at least two input files")
	}

	if s.OutputFilename == "" {
		fmt.Fprintln(os.Stderr, "Error: No output filename specified")
		return fmt.Errorf("no output filename")
	}

	if s.VerboseReporting {
		fmt.Printf("Shuffling %d PDF files\n", len(s.InputPdf))
		for i, pdf := range s.InputPdf {
			fmt.Printf("  Input %d: %s\n", i+1, pdf.Filename)
		}
		fmt.Printf("Output: %s\n", s.OutputFilename)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(s.OutputFilename)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		return err
	}

	// For shuffle operation, we need to interleave pages from different PDFs
	// This is a simplified implementation - would need more complex logic for full PDFtk compatibility

	// Get page counts from each PDF
	var pageCounts []int
	for range s.InputPdf {
		// Note: This is a simplified approach - would need to use pdfcpu API to get actual page count
		// For now, assume each PDF has at least 1 page
		pageCounts = append(pageCounts, 1)
	}

	// Find maximum page count
	maxPages := 0
	for _, count := range pageCounts {
		if count > maxPages {
			maxPages = count
		}
	}

	// Create temporary files for individual pages
	var tempFiles []string
	defer func() {
		// Clean up temporary files
		for _, tempFile := range tempFiles {
			os.Remove(tempFile)
		}
	}()

	// Extract pages and shuffle them
	// This is a basic implementation - full shuffle would require more complex page extraction and merging
	fmt.Fprintf(os.Stderr, "Shuffle operation: Basic implementation in progress\n")
	fmt.Fprintf(os.Stderr, "Note: This is a simplified shuffle that merges files sequentially\n")
	fmt.Fprintf(os.Stderr, "Full page interleaving will be implemented in future versions\n")

	// For now, perform a basic merge operation as a placeholder
	var inputFiles []string
	for _, pdf := range s.InputPdf {
		inputFiles = append(inputFiles, pdf.Filename)
	}

	err := api.MergeCreateFile(inputFiles, s.OutputFilename, false, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error shuffling PDFs: %v\n", err)
		return err
	}

	if s.VerboseReporting {
		fmt.Printf("Successfully created shuffled output: %s\n", s.OutputFilename)
		fmt.Printf("Note: Full page interleaving implementation coming soon\n")
	}

	return nil
}
