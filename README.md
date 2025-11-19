# pdftk-go

A Go port of the popular [PDFtk](https://www.pdflabs.com/tools/pdftk-server/) toolkit for manipulating PDF documents.

## Overview

pdftk-go aims to provide the same functionality as the original PDFtk but implemented in Go for better performance, easier deployment, and modern language features. This project is a faithful port that maintains compatibility with the original PDFtk command-line interface.

## Current Status

🚧 **Work in Progress** 🚧

This is an early-stage port with basic project structure in place. Currently implemented:

- [x] Basic project structure
- [x] CLI framework with traditional PDFtk syntax support
- [x] Core session and document handling
- [x] Placeholder implementations for core operations
- [ ] Full PDF processing capabilities (requires PDF library integration)
- [ ] All PDFtk operations (cat, burst, dump_data, etc.)
- [ ] Complete test suite
- [ ] Documentation and examples

## Planned Features

The goal is to support all original PDFtk operations:

### Core Operations
- **cat** - Concatenate PDF pages into a single output
- **shuffle** - Interleave pages from multiple input PDFs
- **burst** - Split a PDF into individual page files
- **dump_data** - Extract PDF metadata and bookmarks
- **dump_data_fields** - Extract PDF form field information
- **generate_fdf** - Generate FDF files from PDF forms
- **fill_form** - Fill PDF forms using FDF data

### Advanced Features
- **attach_files** - Attach files to PDF documents
- **unpack_files** - Extract attached files from PDFs
- **update_info** - Update PDF metadata
- **background/stamp** - Apply background or stamp to pages
- **rotate** - Rotate PDF pages
- **encrypt/decrypt** - Password protection and encryption

## Installation

### From Source

```bash
git clone https://github.com/egertonspring/pdftk-go
cd pdftk-go
go build ./cmd/pdftk-go
```

### Usage (Planned)

The command-line interface will be fully compatible with original PDFtk:

```bash
# Concatenate PDFs
pdftk-go file1.pdf file2.pdf cat output combined.pdf

# Split PDF into pages
pdftk-go input.pdf burst output page_%04d.pdf

# Extract metadata
pdftk-go input.pdf dump_data output metadata.txt

# Fill forms
pdftk-go form.pdf fill_form data.fdf output filled.pdf

# Encrypt PDF
pdftk-go input.pdf output encrypted.pdf owner_pw password

# Rotate pages
pdftk-go input.pdf cat 1east 2-endsouth output rotated.pdf
```

## Project Structure

```
pdftk-go/
├── cmd/
│   └── pdftk-go/           # Main application entry point
├── internal/
│   ├── cli/                # Command-line interface
│   ├── pdf/                # Core PDF handling and session management
│   └── operations/         # PDF operation implementations
├── pkg/                    # Public API (future)
├── test/                   # Test files and test data
└── docs/                   # Documentation
```

## Architecture

The project follows Go best practices with a clean separation of concerns:

1. **CLI Layer** (`internal/cli`) - Handles command-line parsing and user interaction
2. **Session Management** (`internal/pdf`) - Manages PDF documents, operations, and configuration
3. **Operations** (`internal/operations`) - Implements specific PDF manipulations
4. **PDF Library Integration** - Will use established Go PDF libraries for low-level operations

## Development Roadmap

### Phase 1: Foundation (Current)
- [x] Project structure and build system
- [x] CLI framework with PDFtk-compatible syntax
- [x] Core session and document management
- [x] Basic operation framework

### Phase 2: PDF Library Integration
- [ ] Evaluate and integrate Go PDF library (candidates: pdfcpu, gofpdf, others)
- [ ] Implement basic PDF reading and writing
- [ ] Add comprehensive error handling

### Phase 3: Core Operations
- [ ] Implement `cat` (concatenation)
- [ ] Implement `burst` (page splitting)  
- [ ] Implement `dump_data` (metadata extraction)
- [ ] Add basic tests for core operations

### Phase 4: Advanced Operations
- [ ] Form handling (`fill_form`, `generate_fdf`)
- [ ] File attachments (`attach_files`, `unpack_files`)
- [ ] Encryption and security features
- [ ] Page manipulation (rotation, stamping)

### Phase 5: Compatibility and Polish
- [ ] Full PDFtk command-line compatibility
- [ ] Comprehensive test suite
- [ ] Performance optimization
- [ ] Documentation and examples

## Contributing

Contributions are welcome! This project is in early development, so there are many opportunities to help:

1. **PDF Library Evaluation** - Help evaluate Go PDF libraries for best fit
2. **Operation Implementation** - Port specific PDFtk operations to Go
3. **Testing** - Create test cases and validate against original PDFtk
4. **Documentation** - Improve docs and examples

## Comparison with Original PDFtk

| Feature | Original PDFtk | pdftk-go | Status |
|---------|----------------|----------|--------|
| cat | ✅ | 🚧 | In Progress |
| burst | ✅ | 🚧 | In Progress |
| dump_data | ✅ | 🚧 | In Progress |
| fill_form | ✅ | ❌ | Planned |
| encrypt | ✅ | ❌ | Planned |
| rotate | ✅ | ❌ | Planned |
| stamp | ✅ | ❌ | Planned |
| attach_files | ✅ | ❌ | Planned |

## License

This project will be licensed under the same GPL license as the original PDFtk to maintain compatibility and honor the original work.

## Acknowledgments

- Original PDFtk by Sid Steward and PDFLabs
- The Go community for excellent PDF processing libraries
- All contributors to this port

---

**Note**: This is an independent port and is not affiliated with the original PDFtk or PDFLabs.
