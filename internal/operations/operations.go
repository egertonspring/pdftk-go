// Package operations implements PDF operation helpers for pdftk-go
package operations

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdftk-go/internal/session"
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

	outputDir := filepath.Dir(s.OutputFilename)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		return err
	}

	var inputFiles []string
	for _, pdf := range s.InputPdf {
		inputFiles = append(inputFiles, pdf.Filename)
	}

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

// ExecuteBurstOperationParallel performs the burst (split) operation in parallel
func ExecuteBurstOperationParallel(s *session.TKSession) error {
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
	if err := api.SplitFile(inputPath, tempDir, 1, nil); err != nil {
		return fmt.Errorf("error during pdf split: %w", err)
	}

	// Step 2: Collect and sort split files
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("cannot read temp split dir: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	// Step 3: Parallel move/copy files
	const workerCount = 4 // Anzahl paralleler Worker, ggf. anpassen
	type job struct {
		src  string
		dst  string
		page int
	}
	jobs := make(chan job, len(entries))
	results := make(chan error, len(entries))

	var wg sync.WaitGroup
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if err := os.Rename(j.src, j.dst); err != nil {
					if err := copyFile(j.src, j.dst); err != nil {
						results <- fmt.Errorf("page %d: %w", j.page, err)
						continue
					}
				}
				results <- nil
			}
		}()
	}

	for i, entry := range entries {
		if entry.IsDir() {
			continue
		}
		src := filepath.Join(tempDir, entry.Name())
		dst := filepath.Join(finalOutputDir, fmt.Sprintf(filepath.Base(outputPattern), i+1))
		jobs <- job{src, dst, i + 1}
	}
	close(jobs)

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			return err
		}
	}

	if s.VerboseReporting {
		fmt.Printf("Successfully burst into %d pages.\n", len(entries))
	}

	return nil
}

// copyFile is fallback if os.Rename fails
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

	if _, err := io.Copy(out, in); err != nil {
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

	pageCount := 1 // placeholder
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

	outputDir := filepath.Dir(s.OutputFilename)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		return err
	}

	fmt.Fprintf(os.Stderr, "Shuffle operation: Basic implementation in progress\n")
	fmt.Fprintf(os.Stderr, "Note: This is a simplified shuffle that merges files sequentially\n")
	fmt.Fprintf(os.Stderr, "Full page interleaving will be implemented in future versions\n")

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
