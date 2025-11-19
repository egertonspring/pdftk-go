// Package pdf provides core PDF manipulation functionality
package pdf

import (
	"fmt"
	"io"
	"os"
)

// Document represents a PDF document
type Document struct {
	path     string
	password string
	handle   string // A, B, C, etc. for referencing in operations
}

// NewDocument creates a new PDF document reference
func NewDocument(path, password, handle string) *Document {
	return &Document{
		path:     path,
		password: password,
		handle:   handle,
	}
}

// Path returns the document path
func (d *Document) Path() string {
	return d.path
}

// Password returns the document password
func (d *Document) Password() string {
	return d.password
}

// Handle returns the document handle
func (d *Document) Handle() string {
	return d.handle
}

// PageRange represents a range of pages to process
type PageRange struct {
	Start    int
	End      int
	Rotation int  // 0, 90, 180, 270 degrees
	Reverse  bool // for end-1 syntax
}

// Session represents a pdftk operation session
type Session struct {
	InputDocuments []*Document
	OutputFile     string
	Operation      string
	PageRanges     []PageRange
	Options        map[string]interface{}
	Verbose        bool
	DontAsk        bool
	OutputUTF8     bool
}

// NewSession creates a new operation session
func NewSession() *Session {
	return &Session{
		InputDocuments: make([]*Document, 0),
		Options:        make(map[string]interface{}),
	}
}

// AddDocument adds a document to the session
func (s *Session) AddDocument(doc *Document) {
	s.InputDocuments = append(s.InputDocuments, doc)
}

// Validate checks if the session is valid for the operation
func (s *Session) Validate() error {
	if len(s.InputDocuments) == 0 {
		return fmt.Errorf("no input documents specified")
	}

	switch s.Operation {
	case "burst", "dump_data", "dump_data_utf8", "dump_data_fields", "dump_data_annots",
		"generate_fdf", "unpack_files":
		if len(s.InputDocuments) != 1 {
			return fmt.Errorf("operation %s requires exactly one input document", s.Operation)
		}
	case "cat", "shuffle":
		if len(s.InputDocuments) == 0 {
			return fmt.Errorf("operation %s requires at least one input document", s.Operation)
		}
	}

	return nil
}

// Processor defines the interface for PDF operations
type Processor interface {
	Process(session *Session) error
}

// ProcessorRegistry holds all available processors
var ProcessorRegistry = map[string]Processor{
	// Will be populated as we implement each operation
}

// Execute runs the operation defined in the session
func (s *Session) Execute() error {
	if err := s.Validate(); err != nil {
		return err
	}

	processor, exists := ProcessorRegistry[s.Operation]
	if !exists {
		return fmt.Errorf("operation '%s' is not implemented", s.Operation)
	}

	return processor.Process(s)
}

// PromptForPassword prompts the user for a password
func PromptForPassword() (string, error) {
	fmt.Print("Please enter a password for this PDF (32 char max): ")

	var password string
	if _, err := fmt.Scanln(&password); err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	if len(password) > 32 {
		fmt.Printf("The password you entered was over 32 characters long,\n")
		fmt.Printf("   so I am dropping: \"%s\"\n", password[32:])
		password = password[:32]
	}

	return password, nil
}

// PromptForFilename prompts the user for a filename
func PromptForFilename(message string) (string, error) {
	fmt.Print(message)

	var filename string
	if _, err := fmt.Scanln(&filename); err != nil {
		return "", fmt.Errorf("failed to read filename: %w", err)
	}

	return filename, nil
}

// ConfirmOverwrite asks the user if they want to overwrite a file
func ConfirmOverwrite(filename string) (bool, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return true, nil // File doesn't exist, no need to confirm
	}

	fmt.Printf("The file \"%s\" already exists.\n", filename)
	fmt.Print("Do you want to overwrite it? (y/N): ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	return response == "y" || response == "Y" || response == "yes" || response == "YES", nil
}

// GetPrintWriter returns a writer for the given filename, handling stdout
func GetPrintWriter(filename string) (io.WriteCloser, error) {
	if filename == "" || filename == "-" {
		return os.Stdout, nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", filename, err)
	}

	return file, nil
}
