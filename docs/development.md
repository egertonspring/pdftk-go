# PDFtk-Go Development Guide

## Project Overview

This document outlines the development approach for porting the Java PDFtk implementation to Go.

## Architecture Decisions

### 1. Project Structure

```
pdftk-go/
├── cmd/pdftk-go/           # Main application entry point
├── internal/
│   ├── cli/                # Command-line interface parsing
│   ├── pdf/                # Core PDF session management
│   └── operations/         # PDF operation implementations
├── pkg/                    # Public APIs (future)
├── test/                   # Test files and test data
└── docs/                   # Documentation
```

### 2. Layered Architecture

1. **CLI Layer** - Handles argument parsing and user interaction
2. **Session Layer** - Manages operation sessions, documents, and configuration
3. **Operations Layer** - Implements specific PDF manipulations
4. **PDF Library Layer** - Low-level PDF processing (to be integrated)

## Current Implementation Status

### ✅ Completed
- Basic Go project structure with modules
- CLI framework with PDFtk-compatible syntax parsing
- Core session and document management types
- Placeholder implementations for major operations (cat, burst, dump_data)
- Basic test framework
- Build system (Makefile) with cross-platform support
- Project documentation

### 🚧 In Progress
- PDF library evaluation and integration
- Full operation implementations

### ❌ Todo
- Complete PDF processing capabilities
- All PDFtk operations
- Comprehensive test suite
- Performance optimization
- Documentation

## Next Steps

### Phase 1: PDF Library Integration

**Goal**: Replace placeholder implementations with real PDF processing

**Tasks**:
1. **Evaluate Go PDF Libraries**:
   - [pdfcpu](https://github.com/pdfcpu/pdfcpu) - Pure Go, comprehensive
   - [gofpdf](https://github.com/jung-kurt/gofpdf) - PDF generation focused
   - [unidoc/unipdf](https://unidoc.io/) - Commercial, full-featured
   - Native C bindings to existing libraries

2. **Selection Criteria**:
   - Feature completeness (reading, writing, manipulation)
   - Performance and memory usage
   - License compatibility (GPL)
   - Community support and maintenance
   - Ease of integration

3. **Recommended Choice: pdfcpu**
   - Pure Go implementation
   - Comprehensive PDF processing capabilities
   - Active development and good documentation
   - Apache 2.0 license (compatible)

### Phase 2: Core Operations Implementation

**Priority Order** (based on usage frequency):

1. **cat** - PDF concatenation
   - Merge multiple PDFs into one
   - Handle page ranges and rotations
   - Preserve bookmarks and metadata

2. **burst** - PDF splitting
   - Split PDF into individual pages
   - Generate data dump files
   - Handle output naming patterns

3. **dump_data** - Metadata extraction
   - Extract PDF information
   - Bookmark structure
   - Form fields
   - Annotations

4. **fill_form** - Form handling
   - Read FDF files
   - Fill PDF forms
   - Generate completed PDFs

### Phase 3: Advanced Features

1. **Security Operations**:
   - Password protection (owner/user passwords)
   - Encryption (40-bit, 128-bit, 256-bit AES)
   - Permission settings

2. **Page Manipulation**:
   - Rotation (90, 180, 270 degrees)
   - Background and stamps
   - Page ranges and shuffling

3. **File Operations**:
   - Attach files to PDFs
   - Extract attached files
   - Update PDF metadata

## Implementation Guidelines

### Code Style
- Follow Go conventions (gofmt, golint)
- Use meaningful variable and function names
- Add comprehensive comments
- Include examples in documentation

### Error Handling
- Use Go's idiomatic error handling
- Provide meaningful error messages
- Match original PDFtk error behavior where appropriate

### Testing
- Unit tests for all packages
- Integration tests with real PDF files
- Compatibility tests against original PDFtk
- Performance benchmarks

### Compatibility
- Maintain command-line compatibility with original PDFtk
- Support all major PDFtk operations
- Preserve output format compatibility

## Development Commands

```bash
# Set up development environment
make deps

# Build the project
make build

# Run tests
make test

# Run in development mode
make dev

# Build for all platforms
make build-all

# Format and lint code
make fmt
make lint
```

## Contributing Guidelines

### Adding New Operations

1. **Define Interface**:
   ```go
   type MyOperationProcessor struct{}

   func (m *MyOperationProcessor) Process(session *pdf.Session) error {
       // Implementation
   }
   ```

2. **Register Processor**:
   ```go
   func init() {
       pdf.ProcessorRegistry["my_operation"] = &MyOperationProcessor{}
   }
   ```

3. **Add CLI Support**:
   - Update operation list in CLI parser
   - Add argument validation
   - Update help text

4. **Write Tests**:
   - Unit tests for the processor
   - Integration tests with sample PDFs
   - Error case handling

### Testing Strategy

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test complete operations with real PDFs
3. **Compatibility Tests**: Compare output with original PDFtk
4. **Performance Tests**: Benchmark against original implementation

### Documentation

- Update README.md with new features
- Add operation-specific documentation
- Include examples and common use cases
- Document any differences from original PDFtk

## PDF Library Integration Examples

### Using pdfcpu for cat operation:

```go
import "github.com/pdfcpu/pdfcpu/pkg/api"

func (c *CatProcessor) Process(session *pdf.Session) error {
    var inputFiles []string
    for _, doc := range session.InputDocuments {
        inputFiles = append(inputFiles, doc.Path())
    }
    
    return api.MergeCreateFile(inputFiles, session.OutputFile, nil)
}
```

### Using pdfcpu for burst operation:

```go
func (b *BurstProcessor) Process(session *pdf.Session) error {
    inputFile := session.InputDocuments[0].Path()
    
    // Get page count
    info, err := api.InfoFile(inputFile, nil, nil)
    if err != nil {
        return err
    }
    
    // Extract each page
    for i := 1; i <= info.PageCount; i++ {
        outputFile := fmt.Sprintf(session.OutputFile, i)
        err := api.ExtractPagesFile(inputFile, outputFile, 
            []string{strconv.Itoa(i)}, nil)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## Performance Considerations

- **Memory Usage**: Handle large PDFs efficiently
- **Concurrency**: Utilize Go's concurrency for parallel processing
- **Streaming**: Process large files without loading entirely into memory
- **Caching**: Cache frequently accessed PDF objects

## Security Considerations

- **Input Validation**: Validate all user inputs
- **File Access**: Secure file handling and permissions
- **Memory Safety**: Prevent buffer overflows and memory leaks
- **Password Handling**: Secure password input and storage

This development guide provides the roadmap for completing the PDFtk-Go port while maintaining high code quality and compatibility with the original tool.