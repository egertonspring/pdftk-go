// Package pagerange implements page range parsing for pdftk-go
// This is a faithful port of the Java PageRange class functionality
package pagerange

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PageRange represents a parsed page range (e.g., "1-5", "A2-end", "Beven")
type PageRange struct {
	Begin     int
	End       int
	NumPages  int
	Original  string
	InputFile int // index of input file (for handle-based ranges)
}

// PageFilter represents filtering options for page ranges
type PageFilter struct {
	EvenOnly bool
	OddOnly  bool
	Exclude  map[int]bool
}

// Rotation represents page rotation
type Rotation int

const (
	RotationNone  Rotation = 0   // North/no rotation
	RotationRight Rotation = 90  // East/clockwise 90°
	RotationDown  Rotation = 180 // South/upside down
	RotationLeft  Rotation = 270 // West/counter-clockwise 90°
)

// PageRef represents a reference to a specific page with rotation
type PageRef struct {
	InputFile int      // index in input files array
	PageNum   int      // 1-based page number
	Rotation  Rotation // page rotation
	Handle    string   // original handle (A, B, etc.)
}

// ParsePageRanges parses complex page range specifications
// Examples: "A1-5", "B", "A1-3 Beven", "A2-endevenwest", "B5-1odd"
func ParsePageRanges(args []string, inputFiles []string, handleMap map[string]int, numPages map[int]int) ([][]PageRef, error) {
	var pageSequences [][]PageRef

	for _, arg := range args {
		if isOutputKeyword(arg) {
			break // stop parsing page ranges at "output"
		}

		pageRefs, err := parsePageRangeArg(arg, inputFiles, handleMap, numPages)
		if err != nil {
			return nil, err
		}

		if len(pageRefs) > 0 {
			pageSequences = append(pageSequences, pageRefs)
		}
	}

	return pageSequences, nil
}

// parsePageRangeArg parses a single page range argument
func parsePageRangeArg(arg string, inputFiles []string, handleMap map[string]int, numPages map[int]int) ([]PageRef, error) {
	if arg == "" {
		return nil, nil
	}

	// Parse handle (A, B, AB, etc.)
	handleRegex := regexp.MustCompile(`^([A-Za-z]*)(.*)$`)
	matches := handleRegex.FindStringSubmatch(arg)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid page range format: %s", arg)
	}

	handle := strings.ToUpper(matches[1])
	remaining := matches[2]

	// Determine input file index
	inputIndex := 0 // default to first file
	if handle != "" {
		if idx, exists := handleMap[handle]; exists {
			inputIndex = idx
		} else {
			return nil, fmt.Errorf("unknown handle '%s' in page range: %s", handle, arg)
		}
	}

	// Get number of pages for this input file
	maxPages, exists := numPages[inputIndex]
	if !exists {
		return nil, fmt.Errorf("no page count available for input file %d", inputIndex)
	}

	// Parse page range and modifiers
	pageRange, filter, rotation, err := parseRangeAndModifiers(remaining, maxPages, arg)
	if err != nil {
		return nil, err
	}

	// Generate page references
	return generatePageRefs(pageRange, filter, rotation, inputIndex, handle, maxPages)
}

// parseRangeAndModifiers parses the range part and trailing modifiers
func parseRangeAndModifiers(remaining string, maxPages int, original string) (*PageRange, *PageFilter, Rotation, error) {
	// Parse page range (e.g., "1-5", "3", "end-1", "r5-r1")
	rangeRegex := regexp.MustCompile(`^(r?)(end|\d*)(-?(r?)(end|\d*))?(.*)$`)
	matches := rangeRegex.FindStringSubmatch(remaining)
	if len(matches) != 7 {
		return nil, nil, RotationNone, fmt.Errorf("invalid page range format: %s", original)
	}

	preReverse := matches[1]
	preRange := matches[2]
	hyphenPart := matches[3]
	postReverse := matches[4]
	postRange := matches[5]
	modifiers := matches[6]

	// Parse begin page
	begin, err := parseBound(preRange, maxPages)
	if err != nil {
		return nil, nil, RotationNone, fmt.Errorf("invalid start page in range %s: %v", original, err)
	}

	// Apply reverse for begin page
	if preReverse == "r" && begin > 0 {
		begin = maxPages - begin + 1
	}

	// Parse end page
	end := begin // default: single page
	if strings.HasPrefix(hyphenPart, "-") {
		if postRange != "" {
			end, err = parseBound(postRange, maxPages)
			if err != nil {
				return nil, nil, RotationNone, fmt.Errorf("invalid end page in range %s: %v", original, err)
			}

			// Apply reverse for end page
			if postReverse == "r" && end > 0 {
				end = maxPages - end + 1
			}
		}
	}

	// Handle case where no explicit range given (use all pages)
	if preRange == "" && hyphenPart == "" {
		begin = 1
		end = maxPages
	}

	// Validate range
	if begin < 1 || begin > maxPages {
		return nil, nil, RotationNone, fmt.Errorf("start page %d out of range [1-%d] in: %s", begin, maxPages, original)
	}
	if end < 1 || end > maxPages {
		return nil, nil, RotationNone, fmt.Errorf("end page %d out of range [1-%d] in: %s", end, maxPages, original)
	}

	pageRange := &PageRange{
		Begin:    begin,
		End:      end,
		NumPages: maxPages,
		Original: original,
	}

	// Parse modifiers (even, odd, rotations, exclusions)
	filter, rotation, err := parseModifiers(modifiers, original)
	if err != nil {
		return nil, nil, RotationNone, err
	}

	return pageRange, filter, rotation, nil
}

// parseBound parses a page boundary ("end" or number)
func parseBound(bound string, maxPages int) (int, error) {
	if bound == "" {
		return 0, nil
	}
	if bound == "end" {
		return maxPages, nil
	}
	return strconv.Atoi(bound)
}

// parseModifiers parses trailing modifiers like "even", "odd", "west", etc.
func parseModifiers(modifiers, original string) (*PageFilter, Rotation, error) {
	filter := &PageFilter{Exclude: make(map[int]bool)}
	rotation := RotationNone

	remaining := modifiers
	for remaining != "" {
		var consumed string

		// Check for even/odd
		if strings.HasPrefix(remaining, "even") {
			filter.EvenOnly = true
			consumed = "even"
		} else if strings.HasPrefix(remaining, "odd") {
			filter.OddOnly = true
			consumed = "odd"
		} else if strings.HasPrefix(remaining, "north") {
			rotation = RotationNone
			consumed = "north"
		} else if strings.HasPrefix(remaining, "east") {
			rotation = RotationRight
			consumed = "east"
		} else if strings.HasPrefix(remaining, "south") {
			rotation = RotationDown
			consumed = "south"
		} else if strings.HasPrefix(remaining, "west") {
			rotation = RotationLeft
			consumed = "west"
		} else if strings.HasPrefix(remaining, "right") {
			rotation = RotationRight
			consumed = "right"
		} else if strings.HasPrefix(remaining, "left") {
			rotation = RotationLeft
			consumed = "left"
		} else if strings.HasPrefix(remaining, "down") {
			rotation = RotationDown
			consumed = "down"
		} else {
			return nil, RotationNone, fmt.Errorf("unknown modifier in page range: %s (in %s)", remaining, original)
		}

		remaining = remaining[len(consumed):]
	}

	return filter, rotation, nil
}

// generatePageRefs generates the actual page references from parsed data
func generatePageRefs(pageRange *PageRange, filter *PageFilter, rotation Rotation, inputIndex int, handle string, maxPages int) ([]PageRef, error) {
	var pageRefs []PageRef

	begin := pageRange.Begin
	end := pageRange.End

	// Determine if we need to reverse the sequence
	reverse := end < begin
	if reverse {
		begin, end = end, begin
	}

	// Generate pages in range
	for pageNum := begin; pageNum <= end; pageNum++ {
		// Apply even/odd filter
		if filter.EvenOnly && pageNum%2 != 0 {
			continue
		}
		if filter.OddOnly && pageNum%2 == 0 {
			continue
		}

		// Apply exclusions
		if filter.Exclude[pageNum] {
			continue
		}

		// Check page exists
		if pageNum > maxPages {
			return nil, fmt.Errorf("page %d does not exist in file (has %d pages)", pageNum, maxPages)
		}

		pageRefs = append(pageRefs, PageRef{
			InputFile: inputIndex,
			PageNum:   pageNum,
			Rotation:  rotation,
			Handle:    handle,
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

// isOutputKeyword checks if argument is an output-related keyword
func isOutputKeyword(arg string) bool {
	switch strings.ToLower(arg) {
	case "output", "verbose", "dont_ask", "do_ask":
		return true
	default:
		return false
	}
}

// String returns a human-readable representation of PageRef
func (pr PageRef) String() string {
	rotStr := ""
	switch pr.Rotation {
	case RotationRight:
		rotStr = " (90°)"
	case RotationDown:
		rotStr = " (180°)"
	case RotationLeft:
		rotStr = " (270°)"
	}
	return fmt.Sprintf("%s%d%s", pr.Handle, pr.PageNum, rotStr)
}
