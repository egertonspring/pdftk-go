#!/bin/bash

# Test script for pdftk-go
echo "Building pdftk-go..."
go build ./cmd/pdftk-go

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

# Create a simple test PDF using echo (not a real PDF, but for testing the pipeline)
echo "%PDF-1.4 fake PDF content for testing" > test.pdf

echo "Testing burst operation..."
echo "Running: ./pdftk-go test.pdf burst"
./pdftk-go test.pdf burst

echo ""
echo "Testing with explicit output pattern..."
echo "Running: ./pdftk-go test.pdf burst output page_%04d.pdf"
./pdftk-go test.pdf burst output page_%04d.pdf

# Cleanup
rm -f test.pdf page_*.pdf test_page_*.pdf