// Package session provides the core PDFtk session management
// This is a faithful port of TK_Session.java from the original pdftk-java
package session

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Keyword represents PDFtk operations (port of keyword enum)
type Keyword int

const (
	NoneK Keyword = iota
	// Operations
	CatK      // combine pages from input PDFs into a single output
	ShuffleK  // like cat, but interleaves pages from input ranges
	BurstK    // split a single, input PDF into individual pages
	FilterK   // apply 'filters' to a single, input PDF based on output args
	DumpDataK // no PDF output
	DumpDataUTF8K
	DumpDataFieldsK
	DumpDataFieldsUTF8K
	DumpDataAnnotsK
	GenerateFdfK
	UnpackFilesK // unpack files from input; no PDF output
	FillFormK    // read FDF file and fill PDF form fields
	AttachFileK  // attach files to output
	UpdateInfoK
	UpdateInfoUTF8K
	BackgroundK // promoted from output option to operation in pdftk 1.10
	MultiBackgroundK
	StampK
	MultiStampK
	RotateK // rotate given pages as directed
)

// String returns the string representation of a keyword
func (k Keyword) String() string {
	switch k {
	case CatK:
		return "cat"
	case ShuffleK:
		return "shuffle"
	case BurstK:
		return "burst"
	case FilterK:
		return "filter"
	case DumpDataK:
		return "dump_data"
	case DumpDataUTF8K:
		return "dump_data_utf8"
	case DumpDataFieldsK:
		return "dump_data_fields"
	case DumpDataFieldsUTF8K:
		return "dump_data_fields_utf8"
	case DumpDataAnnotsK:
		return "dump_data_annots"
	case GenerateFdfK:
		return "generate_fdf"
	case UnpackFilesK:
		return "unpack_files"
	case FillFormK:
		return "fill_form"
	case AttachFileK:
		return "attach_file"
	case UpdateInfoK:
		return "update_info"
	case UpdateInfoUTF8K:
		return "update_info_utf8"
	case BackgroundK:
		return "background"
	case MultiBackgroundK:
		return "multibackground"
	case StampK:
		return "stamp"
	case MultiStampK:
		return "multistamp"
	case RotateK:
		return "rotate"
	default:
		return "none"
	}
}

// ParseKeyword converts string to Keyword
func ParseKeyword(s string) Keyword {
	switch strings.ToLower(s) {
	case "cat":
		return CatK
	case "shuffle":
		return ShuffleK
	case "burst":
		return BurstK
	case "filter":
		return FilterK
	case "dump_data":
		return DumpDataK
	case "dump_data_utf8":
		return DumpDataUTF8K
	case "dump_data_fields":
		return DumpDataFieldsK
	case "dump_data_fields_utf8":
		return DumpDataFieldsUTF8K
	case "dump_data_annots":
		return DumpDataAnnotsK
	case "generate_fdf":
		return GenerateFdfK
	case "unpack_files":
		return UnpackFilesK
	case "fill_form":
		return FillFormK
	case "attach_file":
		return AttachFileK
	case "update_info":
		return UpdateInfoK
	case "update_info_utf8":
		return UpdateInfoUTF8K
	case "background":
		return BackgroundK
	case "multibackground":
		return MultiBackgroundK
	case "stamp":
		return StampK
	case "multistamp":
		return MultiStampK
	case "rotate":
		return RotateK
	default:
		return NoneK
	}
}

// InputPdf represents an input PDF file with metadata
type InputPdf struct {
	Filename string
	Password string
	Handle   string
	NumPages int // number of pages (set after opening PDF)
}

// PageRef represents a reference to a page with enhanced parsing support
type PageRef struct {
	InputIndex int    // index into input PDF array
	PageNum    int    // 1-based page number
	Rotation   int    // degrees: 0, 90, 180, 270
	Handle     string // A, B, C, etc.
}

// PageRange represents a parsed range like "1-5", "A2-end", "Beven"
type PageRange struct {
	Begin      int
	End        int
	EvenOnly   bool
	OddOnly    bool
	Rotation   int
	InputIndex int
	Handle     string
}

// ParsePageRange parses complex page range specifications
// Examples: "A1-5", "B", "A1-3even", "B2-endwest", "A5-1odd", "Beven"
func (s *TKSession) ParsePageRange(arg string) ([]PageRef, error) {
	if arg == "" {
		return nil, nil
	}

	// Parse handle more carefully - handle should be exactly one letter followed by optional content
	// Pattern: Single letter handle + optional range + optional modifiers
	var handle, rangepart, modifiers string

	if len(arg) > 0 && ((arg[0] >= 'A' && arg[0] <= 'Z') || (arg[0] >= 'a' && arg[0] <= 'z')) {
		// First character is a letter - could be a handle
		possibleHandle := strings.ToUpper(string(arg[0]))
		remaining := arg[1:]

		// Check if this handle exists in our input files
		if _, exists := s.InputPdfIndex[possibleHandle]; exists {
			handle = possibleHandle

			// Now parse the remaining part for range and modifiers
			// Look for patterns like "1-5even", "even", "2-end", etc.
			rangeRegex := regexp.MustCompile(`^(\d+(?:-(?:end|\d+))?)?(.*)$`)
			matches := rangeRegex.FindStringSubmatch(remaining)
			if len(matches) == 3 {
				rangepart = matches[1]
				modifiers = matches[2]
			} else {
				modifiers = remaining
			}
		} else {
			// Not a valid handle, treat as pure range/modifier
			rangepart = arg
		}
	} else {
		// No handle, just range/modifiers
		rangepart = arg
	} // Determine input file index
	inputIndex := 0 // default to first file
	if handle != "" {
		if idx, exists := s.InputPdfIndex[handle]; exists {
			inputIndex = idx
		} else {
			return nil, fmt.Errorf("unknown handle '%s' in page range: %s", handle, arg)
		}
	}

	// Get number of pages for this input file (simplified - would need PDF reading)
	maxPages := s.InputPdf[inputIndex].NumPages
	if maxPages == 0 {
		maxPages = 999 // placeholder - would be set after opening PDF
	}

	// Parse the range and modifiers
	pageRange, err := s.parseRangeAndModifiers(rangepart+modifiers, maxPages, arg, inputIndex, handle)
	if err != nil {
		return nil, err
	}

	// Generate page references
	return s.generatePageRefs(pageRange)
} // parseRangeAndModifiers parses range like "1-5evenwest"
func (s *TKSession) parseRangeAndModifiers(remaining string, maxPages int, original string, inputIndex int, handle string) (*PageRange, error) {
	pr := &PageRange{
		InputIndex: inputIndex,
		Handle:     handle,
	}

	// Simple regex for basic ranges: number, number-number, number-end
	rangeRegex := regexp.MustCompile(`^(\d*|end)?(-(\d+|end))?(.*)$`)
	matches := rangeRegex.FindStringSubmatch(remaining)
	if len(matches) != 5 {
		return nil, fmt.Errorf("invalid range format: %s", original)
	}

	startStr := matches[1]
	hyphen := matches[2]
	endStr := matches[3]
	modifiers := matches[4]

	// Parse start page
	if startStr == "" {
		pr.Begin = 1
		pr.End = maxPages
	} else {
		begin, err := s.parseBound(startStr, maxPages)
		if err != nil {
			return nil, err
		}
		pr.Begin = begin
		pr.End = begin // default single page
	}

	// Parse end page if range given
	if hyphen != "" && endStr != "" {
		end, err := s.parseBound(endStr, maxPages)
		if err != nil {
			return nil, err
		}
		pr.End = end
	}

	// Parse modifiers (even, odd, rotations)
	err := s.parseModifiers(modifiers, pr)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// parseBound converts "end" or number string to page number
func (s *TKSession) parseBound(bound string, maxPages int) (int, error) {
	if bound == "end" {
		return maxPages, nil
	}
	return strconv.Atoi(bound)
}

// parseModifiers handles even, odd, rotations like "evenwest"
func (s *TKSession) parseModifiers(modifiers string, pr *PageRange) error {
	remaining := modifiers

	for remaining != "" {
		if strings.HasPrefix(remaining, "even") {
			pr.EvenOnly = true
			remaining = remaining[4:]
		} else if strings.HasPrefix(remaining, "odd") {
			pr.OddOnly = true
			remaining = remaining[3:]
		} else if strings.HasPrefix(remaining, "north") {
			pr.Rotation = 0
			remaining = remaining[5:]
		} else if strings.HasPrefix(remaining, "east") {
			pr.Rotation = 90
			remaining = remaining[4:]
		} else if strings.HasPrefix(remaining, "south") {
			pr.Rotation = 180
			remaining = remaining[5:]
		} else if strings.HasPrefix(remaining, "west") {
			pr.Rotation = 270
			remaining = remaining[4:]
		} else if strings.HasPrefix(remaining, "right") {
			pr.Rotation = 90
			remaining = remaining[5:]
		} else if strings.HasPrefix(remaining, "left") {
			pr.Rotation = 270
			remaining = remaining[4:]
		} else if strings.HasPrefix(remaining, "down") {
			pr.Rotation = 180
			remaining = remaining[4:]
		} else {
			return fmt.Errorf("unknown modifier: %s", remaining)
		}
	}

	return nil
}

// generatePageRefs creates the actual page references
func (s *TKSession) generatePageRefs(pr *PageRange) ([]PageRef, error) {
	var pageRefs []PageRef

	begin := pr.Begin
	end := pr.End

	// Handle reverse ranges
	reverse := end < begin
	if reverse {
		begin, end = end, begin
	}

	// Generate pages in range
	for pageNum := begin; pageNum <= end; pageNum++ {
		// Apply even/odd filter
		if pr.EvenOnly && pageNum%2 != 0 {
			continue
		}
		if pr.OddOnly && pageNum%2 == 0 {
			continue
		}

		pageRefs = append(pageRefs, PageRef{
			InputIndex: pr.InputIndex,
			PageNum:    pageNum,
			Rotation:   pr.Rotation,
			Handle:     pr.Handle,
		})
	}

	// Reverse if needed
	if reverse {
		for i := 0; i < len(pageRefs)/2; i++ {
			j := len(pageRefs) - 1 - i
			pageRefs[i], pageRefs[j] = pageRefs[j], pageRefs[i]
		}
	}

	return pageRefs, nil
}

// PageRef is now imported from pagerange package
// Original PageRef struct replaced with pagerange.PageRef

// TKSession is the main session class (faithful port of TK_Session)
type TKSession struct {
	// Core session state
	Valid                 bool
	Authorized            bool
	InputPdfReadersOpened bool
	VerboseReporting      bool
	AskAboutWarnings      bool

	// Creator string
	Creator string

	// Input PDFs
	InputPdf      []InputPdf
	InputPdfIndex map[string]int

	// File attachments
	InputAttachFileFilename []string
	InputAttachFilePagenum  int
	InputAttachFileRelation string

	// Update info
	UpdateInfoFilename string
	UpdateInfoUTF8     bool
	UpdateXmpFilename  string

	// Operation
	Operation Keyword

	// Page sequences for operations
	PageSeq [][]PageRef

	// Output file settings
	OutputFilename    string
	OutputOwnerPW     string
	OutputUserPW      string
	OutputUTF8        bool
	OutputKeepFirstID bool
	OutputKeepFinalID bool

	// Encryption settings
	OutputEncryption128 bool
	OutputEncryption40  bool
	OutputEncryptionAES bool
	OutputEncryption256 bool

	// Permissions
	OutputUserPerms uint32

	// Compression
	OutputCompress   bool
	OutputUncompress bool

	// Background/stamp
	BackgroundFilename      string
	MultiBackgroundFilename string
	StampFilename           string
	MultiStampFilename      string

	// Form data
	FormDataFilename string
	FormDataUTF8     bool

	// Replacement strings
	ReplacementFont string

	// Drop XFA
	DropXfa bool
}

// NewTKSession creates a new session with default values
func NewTKSession() *TKSession {
	return &TKSession{
		Valid:                   false,
		Authorized:              true,
		InputPdfReadersOpened:   false,
		VerboseReporting:        false,
		AskAboutWarnings:        true, // default from original
		Creator:                 "pdftk-go 1.0.0",
		InputPdf:                make([]InputPdf, 0),
		InputPdfIndex:           make(map[string]int),
		InputAttachFileFilename: make([]string, 0),
		InputAttachFilePagenum:  0,
		InputAttachFileRelation: "Unspecified",
		Operation:               NoneK,
		PageSeq:                 make([][]PageRef, 0),
		OutputUserPerms:         0xFFFFFFFF, // all permissions by default
		OutputCompress:          true,       // default compression on
	}
}

// Parse parses command line arguments (port of TK_Session.parse())
func (s *TKSession) Parse(args []string) error {
	if len(args) == 0 {
		return errors.New("no arguments provided")
	}

	argIndex := 0

	// Parse input files and handles
	for argIndex < len(args) {
		arg := args[argIndex]

		// Check if this is an operation keyword
		if op := ParseKeyword(arg); op != NoneK {
			s.Operation = op
			argIndex++
			break
		}

		// Check for handle assignment (A=file.pdf)
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				handle := parts[0]
				filename := parts[1]

				// Validate handle (should be single letter)
				if len(handle) != 1 || !isValidHandle(handle) {
					return fmt.Errorf("invalid handle: %s", handle)
				}

				inputPdf := InputPdf{
					Filename: filename,
					Handle:   handle,
					NumPages: 0,
				}

				s.InputPdf = append(s.InputPdf, inputPdf)
				s.InputPdfIndex[handle] = len(s.InputPdf) - 1
			}
		} else {
			// Simple filename
			inputPdf := InputPdf{
				Filename: arg,
				Handle:   generateHandle(len(s.InputPdf)),
				NumPages: 0,
			}

			s.InputPdf = append(s.InputPdf, inputPdf)
			s.InputPdfIndex[inputPdf.Handle] = len(s.InputPdf) - 1
		}

		argIndex++
	}

	// Parse operation-specific arguments
	if s.Operation == NoneK {
		return errors.New("no operation specified")
	}

	// Parse remaining arguments based on operation
	switch s.Operation {
	case CatK, ShuffleK:
		err := s.parseCatArguments(args[argIndex:])
		if err != nil {
			return err
		}
	case BurstK:
		err := s.parseBurstArguments(args[argIndex:])
		if err != nil {
			return err
		}
	case DumpDataK, DumpDataUTF8K, DumpDataFieldsK, DumpDataAnnotsK:
		err := s.parseDumpDataArguments(args[argIndex:])
		if err != nil {
			return err
		}
	default:
		// For now, just parse output filename
		err := s.parseOutputArguments(args[argIndex:])
		if err != nil {
			return err
		}
	}

	s.Valid = true
	return nil
}

// parseCatArguments parses arguments for cat/shuffle operations
func (s *TKSession) parseCatArguments(args []string) error {
	// Look for page ranges and output specification
	outputIndex := -1

	// Find "output" keyword
	for i, arg := range args {
		if strings.ToLower(arg) == "output" {
			outputIndex = i
			break
		}
	}

	if outputIndex == -1 {
		return errors.New("output filename not specified")
	}

	if outputIndex+1 >= len(args) {
		return errors.New("output filename missing after 'output' keyword")
	}

	s.OutputFilename = args[outputIndex+1]

	// Parse page ranges with enhanced support
	pageRangeArgs := args[:outputIndex]
	if len(pageRangeArgs) == 0 {
		// No page ranges specified, use all pages from all inputs
		s.generateDefaultPageSequence()
	} else {
		// Parse each page range argument
		var allPageSeqs [][]PageRef

		for _, rangeArg := range pageRangeArgs {
			pageRefs, err := s.ParsePageRange(rangeArg)
			if err != nil {
				return fmt.Errorf("error parsing page range '%s': %v", rangeArg, err)
			}
			if len(pageRefs) > 0 {
				allPageSeqs = append(allPageSeqs, pageRefs)
			}
		}

		if len(allPageSeqs) > 0 {
			s.PageSeq = allPageSeqs
		} else {
			// Fallback to default
			s.generateDefaultPageSequence()
		}
	}

	return nil
}

// parseBurstArguments parses arguments for burst operation
func (s *TKSession) parseBurstArguments(args []string) error {
	// Look for output pattern
	for i, arg := range args {
		if strings.ToLower(arg) == "output" && i+1 < len(args) {
			s.OutputFilename = args[i+1]
			break
		}
	}

	// If no output pattern specified, use default
	if s.OutputFilename == "" {
		if len(s.InputPdf) > 0 {
			base := strings.TrimSuffix(s.InputPdf[0].Filename, ".pdf")
			s.OutputFilename = base + "_page_%04d.pdf"
		}
	}

	return nil
}

// parseDumpDataArguments parses arguments for dump_data operations
func (s *TKSession) parseDumpDataArguments(args []string) error {
	// Look for output filename
	for i, arg := range args {
		if strings.ToLower(arg) == "output" && i+1 < len(args) {
			s.OutputFilename = args[i+1]
			break
		}
	}

	// If no output specified, will output to stdout
	return nil
}

// parseOutputArguments parses generic output arguments
func (s *TKSession) parseOutputArguments(args []string) error {
	for i, arg := range args {
		switch strings.ToLower(arg) {
		case "output":
			if i+1 < len(args) {
				s.OutputFilename = args[i+1]
			}
		case "verbose":
			s.VerboseReporting = true
		case "dont_ask":
			s.AskAboutWarnings = false
		case "do_ask":
			s.AskAboutWarnings = true
		}
	}
	return nil
}

// generateDefaultPageSequence creates page sequence for all pages from all inputs
func (s *TKSession) generateDefaultPageSequence() {
	var pageSeq []PageRef

	for inputIndex := range s.InputPdf {
		// For now, assume each PDF has pages (we'd need to open PDFs to get actual count)
		// This is simplified - real implementation would read PDF to get page count
		pageRef := PageRef{
			InputIndex: inputIndex,
			PageNum:    1, // Simplified - would need actual page range
			Rotation:   0,
			Handle:     s.InputPdf[inputIndex].Handle,
		}
		pageSeq = append(pageSeq, pageRef)
	}

	s.PageSeq = [][]PageRef{pageSeq}
}

// IsValid returns whether the session is valid
func (s *TKSession) IsValid() bool {
	return s.Valid
}

// DumpSessionData prints session information (port of dump_session_data())
func (s *TKSession) DumpSessionData() {
	if s.VerboseReporting {
		fmt.Printf("Input PDF count: %d\n", len(s.InputPdf))
		for i, pdf := range s.InputPdf {
			fmt.Printf("  Input %d: %s (handle: %s)\n", i+1, pdf.Filename, pdf.Handle)
		}
		fmt.Printf("Operation: %s\n", s.Operation.String())
		if s.OutputFilename != "" {
			fmt.Printf("Output: %s\n", s.OutputFilename)
		}
	}
}

// CreateOutput creates output based on operation
func (s *TKSession) CreateOutput() error {
	switch s.Operation {
	case CatK:
		// Call external operation function
		return s.createCatOutput()
	case BurstK:
		return s.createBurstOutput()
	case DumpDataK, DumpDataUTF8K, DumpDataFieldsK, DumpDataFieldsUTF8K, DumpDataAnnotsK:
		return s.createDumpDataOutput()
	default:
		return fmt.Errorf("unsupported operation: %v", s.Operation)
	}
}

// createCatOutput implements cat operation
func (s *TKSession) createCatOutput() error {
	// This would need actual PDF library implementation
	fmt.Printf("Cat operation: merging %d input files to %s\n", len(s.InputPdf), s.OutputFilename)
	return nil
}

// createBurstOutput implements burst operation
func (s *TKSession) createBurstOutput() error {
	// This would need actual PDF library implementation
	fmt.Printf("Burst operation: splitting %s into pages\n", s.InputPdf[0].Filename)
	return nil
}

// createDumpDataOutput implements dump_data operations
func (s *TKSession) createDumpDataOutput() error {
	// This would need actual PDF library implementation
	fmt.Printf("Dump data operation for %s\n", s.InputPdf[0].Filename)
	return nil
}

// Helper functions

// isValidHandle checks if a handle is valid (A-Z, a-z)
func isValidHandle(handle string) bool {
	if len(handle) != 1 {
		return false
	}
	r := rune(handle[0])
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

// generateHandle generates a handle (A, B, C, ...) for the given index
func generateHandle(index int) string {
	if index < 26 {
		return string(rune('A' + index))
	}
	// For more than 26 files, use double letters
	first := index / 26
	second := index % 26
	return string(rune('A'+first)) + string(rune('A'+second))
}
