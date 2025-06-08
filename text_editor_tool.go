package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TextEditorTool is an interface for what Claude expects an agent's
// provided text editor tool to be able to do.
type TextEditorTool interface {
	// View examines the contents of a file or lists the contents
	// of a directory. For files, it returns the file content,
	// optionally within a specific line range. For directories,
	// it returns the list of directory contents.  Both start and
	// end are optional and 1-indexed. If non-nil, start specifies
	// the first line to include. If non-nil, end specifies the
	// last line to include with -1 indicating until the end of
	// file. Returns an error if the file or directory could not
	// be read or if the requested range could not be satisfied.
	View(path string, start *int, end *int) (contents string, err error)

	// StringReplace replaces a specific string in a file with a
	// new string The string in from must match exactly, including
	// whitespace and indentation. If there is exactly one match
	// for from, this method replaces it with the string in to and
	// returns a nil error. If there are zero or more than one
	// matches of from, this method returns a non-nil error.
	StringReplace(path, from, to string) error

	// Create creates a new file with the specified contents at
	// the given path. Returns an error if the file could not be
	// created.
	Create(path, contents string) error

	// Insert inserts text at a specific location in a file.
	// afterLine specifies the line number after which to insert
	// the text (0 to insert at the beginning of the
	// file). Returns an error if the file could not be modified.
	Insert(path string, afterLine int, text string) error

	// UndoEdit reverts the last edit made to a file. Returns an
	// error if the file could not be reverted.
	UndoEdit(path string) error
}

// SimpleTextEditorTool is a basic implementation of the TextEditorTool
// interface that operates on the filesystem.
type SimpleTextEditorTool struct {
	// undoHistory maps file paths to their previous content for undo
	// operations
	undoHistory map[string]string
}

// NewSimpleTextEditorTool creates a new instance of SimpleTextEditorTool.
func NewSimpleTextEditorTool() *SimpleTextEditorTool {
	return &SimpleTextEditorTool{
		undoHistory: make(map[string]string),
	}
}

// View examines the contents of a file or lists the contents of a directory.
func (s *SimpleTextEditorTool) View(path string, start *int, end *int) (
	string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if info.IsDir() {
		return s.viewDirectory(path)
	}

	return s.viewFile(path, start, end)
}

// viewDirectory lists the contents of a directory.
func (s *SimpleTextEditorTool) viewDirectory(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	var result []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		result = append(result, name)
	}

	return strings.Join(result, "\n"), nil
}

// viewFile reads and returns the contents of a file, optionally within a
// specific line range.
func (s *SimpleTextEditorTool) viewFile(path string, start *int, end *int) (
	string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	lines := strings.Split(string(content), "\n")

	// If the file ends with a newline, remove the empty last element
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Apply line range filtering if specified
	if start != nil || end != nil {
		startIdx := 0
		endIdx := len(lines)

		if start != nil {
			if *start < 1 {
				return "", fmt.Errorf("start line must be >= 1, got %d",
					*start)
			}
			startIdx = *start - 1 // Convert to 0-indexed
			if startIdx >= len(lines) {
				return "", fmt.Errorf(
					"start line %d exceeds file length %d",
					*start, len(lines))
			}
		}

		if end != nil {
			if *end == -1 {
				endIdx = len(lines)
			} else {
				if *end < 1 {
					return "", fmt.Errorf("end line must be >= 1 or -1, "+
						"got %d", *end)
				}
				endIdx = *end // Convert to exclusive upper bound
				if endIdx > len(lines) {
					return "", fmt.Errorf(
						"end line %d exceeds file length %d",
						*end, len(lines))
				}
			}
		}

		if startIdx >= endIdx {
			return "", fmt.Errorf(
				"start line %d must be less than end line %d",
				startIdx+1, endIdx)
		}

		lines = lines[startIdx:endIdx]
	}

	return strings.Join(lines, "\n"), nil
}

// StringReplace replaces a specific string in a file with a new string.
func (s *SimpleTextEditorTool) StringReplace(path, from, to string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	originalContent := string(content)
	count := strings.Count(originalContent, from)

	if count == 0 {
		return fmt.Errorf("string %q not found in file %s", from, path)
	}

	if count > 1 {
		return fmt.Errorf("string %q appears %d times in file %s, "+
			"expected exactly 1", from, count, path)
	}

	// Store original content for undo
	s.undoHistory[path] = originalContent

	newContent := strings.Replace(originalContent, from, to, 1)
	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		// Remove from undo history on failure
		delete(s.undoHistory, path)
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// Create creates a new file with the specified contents at the given path.
func (s *SimpleTextEditorTool) Create(path, contents string) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file %s already exists", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if file %s exists: %w", path, err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	err := os.WriteFile(path, []byte(contents), 0644)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}

	return nil
}

// Insert inserts text at a specific location in a file.
func (s *SimpleTextEditorTool) Insert(path string, afterLine int,
	text string) error {
	if afterLine < 0 {
		return fmt.Errorf("afterLine must be >= 0, got %d", afterLine)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	originalContent := string(content)
	lines := strings.Split(originalContent, "\n")

	// If the file ends with a newline, remove the empty last element
	hadFinalNewline := strings.HasSuffix(originalContent, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" && hadFinalNewline {
		lines = lines[:len(lines)-1]
	}

	if afterLine > len(lines) {
		return fmt.Errorf("afterLine %d exceeds file length %d",
			afterLine, len(lines))
	}

	// Store original content for undo
	s.undoHistory[path] = originalContent

	// Insert the text
	var newLines []string
	newLines = append(newLines, lines[:afterLine]...)
	newLines = append(newLines, text)
	newLines = append(newLines, lines[afterLine:]...)

	newContent := strings.Join(newLines, "\n")
	if hadFinalNewline || len(lines) == 0 {
		newContent += "\n"
	}

	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		// Remove from undo history on failure
		delete(s.undoHistory, path)
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// UndoEdit reverts the last edit made to a file.
func (s *SimpleTextEditorTool) UndoEdit(path string) error {
	originalContent, exists := s.undoHistory[path]
	if !exists {
		return fmt.Errorf("no undo history available for file %s", path)
	}

	err := os.WriteFile(path, []byte(originalContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to undo edit for file %s: %w", path, err)
	}

	// Remove from undo history after successful undo
	delete(s.undoHistory, path)

	return nil
}
