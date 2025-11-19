package pdf

import (
	"testing"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument("/path/to/test.pdf", "password", "A")

	if doc.Path() != "/path/to/test.pdf" {
		t.Errorf("Expected path '/path/to/test.pdf', got '%s'", doc.Path())
	}

	if doc.Password() != "password" {
		t.Errorf("Expected password 'password', got '%s'", doc.Password())
	}

	if doc.Handle() != "A" {
		t.Errorf("Expected handle 'A', got '%s'", doc.Handle())
	}
}

func TestNewSession(t *testing.T) {
	session := NewSession()

	if session == nil {
		t.Error("NewSession() returned nil")
	}

	if len(session.InputDocuments) != 0 {
		t.Errorf("Expected 0 input documents, got %d", len(session.InputDocuments))
	}

	if session.Options == nil {
		t.Error("Session options map is nil")
	}
}

func TestSessionAddDocument(t *testing.T) {
	session := NewSession()
	doc := NewDocument("/test.pdf", "", "")

	session.AddDocument(doc)

	if len(session.InputDocuments) != 1 {
		t.Errorf("Expected 1 input document, got %d", len(session.InputDocuments))
	}

	if session.InputDocuments[0] != doc {
		t.Error("Added document doesn't match")
	}
}

func TestSessionValidate(t *testing.T) {
	session := NewSession()

	// Test empty session
	err := session.Validate()
	if err == nil {
		t.Error("Expected error for empty session, got nil")
	}

	// Test with document
	doc := NewDocument("/test.pdf", "", "")
	session.AddDocument(doc)

	// Test burst operation with one document
	session.Operation = "burst"
	err = session.Validate()
	if err != nil {
		t.Errorf("Expected no error for burst with one document, got: %v", err)
	}

	// Test burst operation with multiple documents
	session.AddDocument(NewDocument("/test2.pdf", "", ""))
	err = session.Validate()
	if err == nil {
		t.Error("Expected error for burst with multiple documents, got nil")
	}

	// Test cat operation with multiple documents
	session.Operation = "cat"
	err = session.Validate()
	if err != nil {
		t.Errorf("Expected no error for cat with multiple documents, got: %v", err)
	}
}

func TestPromptForPassword(t *testing.T) {
	// Note: This test can't easily test interactive input
	// In a real implementation, you might want to make the input source configurable
	t.Skip("Interactive test - skipping for automated testing")
}

func TestGetPrintWriter(t *testing.T) {
	// Test stdout case
	writer, err := GetPrintWriter("")
	if err != nil {
		t.Errorf("Expected no error for empty filename, got: %v", err)
	}
	if writer == nil {
		t.Error("Expected non-nil writer")
	}

	writer, err = GetPrintWriter("-")
	if err != nil {
		t.Errorf("Expected no error for '-' filename, got: %v", err)
	}
	if writer == nil {
		t.Error("Expected non-nil writer")
	}
}
