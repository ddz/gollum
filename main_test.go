package main

import (
	"strings"
	"testing"
)

func TestAddLineNumbers(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		startLine *int
		expected  string
	}{
		{
			name:      "empty string",
			input:     "",
			startLine: nil,
			expected:  "",
		},
		{
			name:      "single line",
			input:     "hello world",
			startLine: nil,
			expected:  "1: hello world",
		},
		{
			name:      "multiple lines",
			input:     "line one\nline two\nline three",
			startLine: nil,
			expected:  "1: line one\n2: line two\n3: line three",
		},
		{
			name:      "with custom start line",
			input:     "first\nsecond",
			startLine: &[]int{10}[0],
			expected:  "10: first\n11: second",
		},
		{
			name:      "text ending with newline",
			input:     "line one\nline two\n",
			startLine: nil,
			expected:  "1: line one\n2: line two",
		},
		{
			name:      "single line ending with newline",
			input:     "single line\n",
			startLine: nil,
			expected:  "1: single line",
		},
		{
			name:      "empty lines in between",
			input:     "line one\n\nline three",
			startLine: nil,
			expected:  "1: line one\n2: \n3: line three",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addLineNumbers(tt.input, tt.startLine)
			if result != tt.expected {
				t.Errorf("addLineNumbers() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSystemPromptEmbedded(t *testing.T) {
	// Test that the systemPrompt is properly embedded from prompt.txt
	if systemPrompt == "" {
		t.Error("systemPrompt should not be empty when embedded from prompt.txt")
	}

	// Check that it contains expected content from prompt.txt
	expectedContent := []string{
		"bash commands",
		"text editing capabilities",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(systemPrompt, expected) {
			t.Errorf("systemPrompt should contain %q, but it doesn't. Content: %q", expected, systemPrompt)
		}
	}
}
