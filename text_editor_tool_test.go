package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// RunTextEditorToolTests contains reusable tests for any TextEditorTool
// implementation. This allows testing the interface contract across
// different implementations.
func RunTextEditorToolTests(t *testing.T, tool TextEditorTool) {
	t.Run("View", func(t *testing.T) {
		testTextEditorToolView(t, tool)
	})

	t.Run("StringReplace", func(t *testing.T) {
		testTextEditorToolStringReplace(t, tool)
	})

	t.Run("Create", func(t *testing.T) {
		testTextEditorToolCreate(t, tool)
	})

	t.Run("Insert", func(t *testing.T) {
		testTextEditorToolInsert(t, tool)
	})

	t.Run("UndoEdit", func(t *testing.T) {
		testTextEditorToolUndoEdit(t, tool)
	})
}

func testTextEditorToolView(t *testing.T, tool TextEditorTool) {
	tempDir := t.TempDir()

	t.Run("ViewFile", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tempDir, "test.txt")
		content := "line 1\nline 2\nline 3\nline 4\nline 5"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Test viewing entire file
		result, err := tool.View(testFile, nil, nil)
		if err != nil {
			t.Errorf("View() error = %v", err)
		}
		if result != content {
			t.Errorf("View() = %q, want %q", result, content)
		}

		// Test viewing with start line
		start := 2
		result, err = tool.View(testFile, &start, nil)
		if err != nil {
			t.Errorf("View() with start error = %v", err)
		}
		expected := "line 2\nline 3\nline 4\nline 5"
		if result != expected {
			t.Errorf("View() with start = %q, want %q", result, expected)
		}

		// Test viewing with end line
		end := 3
		result, err = tool.View(testFile, nil, &end)
		if err != nil {
			t.Errorf("View() with end error = %v", err)
		}
		expected = "line 1\nline 2\nline 3"
		if result != expected {
			t.Errorf("View() with end = %q, want %q", result, expected)
		}

		// Test viewing with start and end
		start = 2
		end = 4
		result, err = tool.View(testFile, &start, &end)
		if err != nil {
			t.Errorf("View() with start and end error = %v", err)
		}
		expected = "line 2\nline 3\nline 4"
		if result != expected {
			t.Errorf("View() with start and end = %q, want %q",
				result, expected)
		}

		// Test viewing with end = -1 (until end of file)
		start = 3
		end = -1
		result, err = tool.View(testFile, &start, &end)
		if err != nil {
			t.Errorf("View() with end = -1 error = %v", err)
		}
		expected = "line 3\nline 4\nline 5"
		if result != expected {
			t.Errorf("View() with end = -1 = %q, want %q", result, expected)
		}
	})

	t.Run("ViewDirectory", func(t *testing.T) {
		// Create test directory structure
		testDir := filepath.Join(tempDir, "testdir")
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Create files and subdirectories
		err = os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte(""),
			0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		err = os.MkdirAll(filepath.Join(testDir, "subdir"), 0755)
		if err != nil {
			t.Fatalf("Failed to create test subdirectory: %v", err)
		}

		result, err := tool.View(testDir, nil, nil)
		if err != nil {
			t.Errorf("View() directory error = %v", err)
		}

		// Check that both file and directory are listed
		if !strings.Contains(result, "file1.txt") {
			t.Errorf("View() directory result missing file1.txt: %q",
				result)
		}
		if !strings.Contains(result, "subdir/") {
			t.Errorf("View() directory result missing subdir/: %q", result)
		}
	})

	t.Run("ViewErrors", func(t *testing.T) {
		// Test non-existent file
		_, err := tool.View(filepath.Join(tempDir, "nonexistent.txt"),
			nil, nil)
		if err == nil {
			t.Error("View() non-existent file should return error")
		}

		// Create test file for range tests
		testFile := filepath.Join(tempDir, "range_test.txt")
		content := "line 1\nline 2\nline 3"
		err = os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Test invalid start line
		start := 0
		_, err = tool.View(testFile, &start, nil)
		if err == nil {
			t.Error("View() with start = 0 should return error")
		}

		// Test start line beyond file length
		start = 10
		_, err = tool.View(testFile, &start, nil)
		if err == nil {
			t.Error("View() with start beyond file length should return error")
		}

		// Test invalid end line
		end := 0
		_, err = tool.View(testFile, nil, &end)
		if err == nil {
			t.Error("View() with end = 0 should return error")
		}
	})
}

func testTextEditorToolStringReplace(t *testing.T, tool TextEditorTool) {
	tempDir := t.TempDir()

	t.Run("SuccessfulReplace", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "replace_test.txt")
		content := "Hello world\nThis is a test\nHello again"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = tool.StringReplace(testFile, "This is a test", "This was a test")
		if err != nil {
			t.Errorf("StringReplace() error = %v", err)
		}

		// Verify the replacement
		result, err := tool.View(testFile, nil, nil)
		if err != nil {
			t.Fatalf("Failed to read modified file: %v", err)
		}
		expected := "Hello world\nThis was a test\nHello again"
		if result != expected {
			t.Errorf("StringReplace() result = %q, want %q", result, expected)
		}
	})

	t.Run("StringNotFound", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "not_found_test.txt")
		content := "Hello world\nThis is a test"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = tool.StringReplace(testFile, "nonexistent string", "replacement")
		if err == nil {
			t.Error("StringReplace() with non-existent string should return error")
		}
	})

	t.Run("MultipleMatches", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "multiple_test.txt")
		content := "Hello world\nHello again\nHello once more"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = tool.StringReplace(testFile, "Hello", "Hi")
		if err == nil {
			t.Error("StringReplace() with multiple matches should return error")
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		err := tool.StringReplace(filepath.Join(tempDir, "nonexistent.txt"),
			"from", "to")
		if err == nil {
			t.Error("StringReplace() on non-existent file should return error")
		}
	})
}

func testTextEditorToolCreate(t *testing.T, tool TextEditorTool) {
	tempDir := t.TempDir()

	t.Run("CreateNewFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "new_file.txt")
		content := "This is new content\nLine 2"

		err := tool.Create(testFile, content)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}

		// Verify the file was created with correct content
		result, err := tool.View(testFile, nil, nil)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}
		if result != content {
			t.Errorf("Create() content = %q, want %q", result, content)
		}
	})

	t.Run("CreateInNestedDirectory", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "nested", "dir", "file.txt")
		content := "Nested file content"

		err := tool.Create(testFile, content)
		if err != nil {
			t.Errorf("Create() in nested directory error = %v", err)
		}

		// Verify the file was created
		result, err := tool.View(testFile, nil, nil)
		if err != nil {
			t.Fatalf("Failed to read created nested file: %v", err)
		}
		if result != content {
			t.Errorf("Create() nested content = %q, want %q", result, content)
		}
	})

	t.Run("CreateExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "existing.txt")
		err := os.WriteFile(testFile, []byte("existing content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		err = tool.Create(testFile, "new content")
		if err == nil {
			t.Error("Create() on existing file should return error")
		}
	})
}

func testTextEditorToolInsert(t *testing.T, tool TextEditorTool) {
	tempDir := t.TempDir()

	t.Run("InsertAtBeginning", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "insert_begin.txt")
		content := "line 1\nline 2\nline 3"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = tool.Insert(testFile, 0, "new first line")
		if err != nil {
			t.Errorf("Insert() at beginning error = %v", err)
		}

		result, err := tool.View(testFile, nil, nil)
		if err != nil {
			t.Fatalf("Failed to read modified file: %v", err)
		}
		expected := "new first line\nline 1\nline 2\nline 3"
		if result != expected {
			t.Errorf("Insert() at beginning = %q, want %q", result, expected)
		}
	})

	t.Run("InsertInMiddle", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "insert_middle.txt")
		content := "line 1\nline 2\nline 3"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = tool.Insert(testFile, 2, "inserted line")
		if err != nil {
			t.Errorf("Insert() in middle error = %v", err)
		}

		result, err := tool.View(testFile, nil, nil)
		if err != nil {
			t.Fatalf("Failed to read modified file: %v", err)
		}
		expected := "line 1\nline 2\ninserted line\nline 3"
		if result != expected {
			t.Errorf("Insert() in middle = %q, want %q", result, expected)
		}
	})

	t.Run("InsertAtEnd", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "insert_end.txt")
		content := "line 1\nline 2\nline 3"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err = tool.Insert(testFile, 3, "new last line")
		if err != nil {
			t.Errorf("Insert() at end error = %v", err)
		}

		result, err := tool.View(testFile, nil, nil)
		if err != nil {
			t.Fatalf("Failed to read modified file: %v", err)
		}
		expected := "line 1\nline 2\nline 3\nnew last line"
		if result != expected {
			t.Errorf("Insert() at end = %q, want %q", result, expected)
		}
	})

	t.Run("InsertErrors", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "insert_errors.txt")
		content := "line 1\nline 2"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Test negative afterLine
		err = tool.Insert(testFile, -1, "text")
		if err == nil {
			t.Error("Insert() with negative afterLine should return error")
		}

		// Test afterLine beyond file length
		err = tool.Insert(testFile, 10, "text")
		if err == nil {
			t.Error("Insert() with afterLine beyond file length should return error")
		}

		// Test non-existent file
		err = tool.Insert(filepath.Join(tempDir, "nonexistent.txt"), 0, "text")
		if err == nil {
			t.Error("Insert() on non-existent file should return error")
		}
	})
}

func testTextEditorToolUndoEdit(t *testing.T, tool TextEditorTool) {
	tempDir := t.TempDir()

	t.Run("UndoStringReplace", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "undo_replace.txt")
		originalContent := "Hello world\nThis is a test"
		err := os.WriteFile(testFile, []byte(originalContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Make a change
		err = tool.StringReplace(testFile, "Hello", "Hi")
		if err != nil {
			t.Fatalf("StringReplace() failed: %v", err)
		}

		// Undo the change
		err = tool.UndoEdit(testFile)
		if err != nil {
			t.Errorf("UndoEdit() error = %v", err)
		}

		// Verify content is restored
		result, err := tool.View(testFile, nil, nil)
		if err != nil {
			t.Fatalf("Failed to read file after undo: %v", err)
		}
		if result != originalContent {
			t.Errorf("UndoEdit() result = %q, want %q", result, originalContent)
		}
	})

	t.Run("UndoInsert", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "undo_insert.txt")
		originalContent := "line 1\nline 2"
		err := os.WriteFile(testFile, []byte(originalContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Make a change
		err = tool.Insert(testFile, 1, "inserted line")
		if err != nil {
			t.Fatalf("Insert() failed: %v", err)
		}

		// Undo the change
		err = tool.UndoEdit(testFile)
		if err != nil {
			t.Errorf("UndoEdit() error = %v", err)
		}

		// Verify content is restored
		result, err := tool.View(testFile, nil, nil)
		if err != nil {
			t.Fatalf("Failed to read file after undo: %v", err)
		}
		if result != originalContent {
			t.Errorf("UndoEdit() result = %q, want %q", result, originalContent)
		}
	})

	t.Run("UndoNoHistory", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "no_history.txt")
		content := "test content"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Try to undo without any edit history
		err = tool.UndoEdit(testFile)
		if err == nil {
			t.Error("UndoEdit() without history should return error")
		}
	})

	t.Run("UndoTwice", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "undo_twice.txt")
		originalContent := "original content"
		err := os.WriteFile(testFile, []byte(originalContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Make a change
		err = tool.StringReplace(testFile, "original", "modified")
		if err != nil {
			t.Fatalf("StringReplace() failed: %v", err)
		}

		// First undo should succeed
		err = tool.UndoEdit(testFile)
		if err != nil {
			t.Errorf("First UndoEdit() error = %v", err)
		}

		// Second undo should fail (no history)
		err = tool.UndoEdit(testFile)
		if err == nil {
			t.Error("Second UndoEdit() should return error")
		}
	})
}

// TestSimpleTextEditorTool tests the SimpleTextEditorTool implementation
// using the reusable interface tests.
func TestSimpleTextEditorTool(t *testing.T) {
	tool := NewSimpleTextEditorTool()
	RunTextEditorToolTests(t, tool)
}

// TestSimpleTextEditorToolSpecific tests SimpleTextEditorTool-specific
// functionality that may not be part of the interface contract.
func TestSimpleTextEditorToolSpecific(t *testing.T) {
	tool := NewSimpleTextEditorTool()
	tempDir := t.TempDir()

	t.Run("UndoHistoryManagement", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "history_test.txt")
		content := "original content"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Verify undo history is initially empty
		if len(tool.undoHistory) != 0 {
			t.Errorf("Initial undo history should be empty, got %d entries",
				len(tool.undoHistory))
		}

		// Make a change
		err = tool.StringReplace(testFile, "original", "modified")
		if err != nil {
			t.Fatalf("StringReplace() failed: %v", err)
		}

		// Verify undo history has one entry
		if len(tool.undoHistory) != 1 {
			t.Errorf("Undo history should have 1 entry, got %d",
				len(tool.undoHistory))
		}

		// Undo the change
		err = tool.UndoEdit(testFile)
		if err != nil {
			t.Errorf("UndoEdit() error = %v", err)
		}

		// Verify undo history is cleared after undo
		if len(tool.undoHistory) != 0 {
			t.Errorf("Undo history should be empty after undo, got %d entries",
				len(tool.undoHistory))
		}
	})

	t.Run("FileWithTrailingNewline", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "trailing_newline.txt")
		content := "line 1\nline 2\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// View should preserve the file structure
		result, err := tool.View(testFile, nil, nil)
		if err != nil {
			t.Errorf("View() error = %v", err)
		}
		expected := "line 1\nline 2"
		if result != expected {
			t.Errorf("View() = %q, want %q", result, expected)
		}

		// Insert should preserve trailing newline behavior
		err = tool.Insert(testFile, 1, "inserted")
		if err != nil {
			t.Errorf("Insert() error = %v", err)
		}

		// Read file directly to check newline preservation
		modifiedContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read modified file: %v", err)
		}
		expectedContent := "line 1\ninserted\nline 2\n"
		if string(modifiedContent) != expectedContent {
			t.Errorf("Insert() preserved content = %q, want %q",
				string(modifiedContent), expectedContent)
		}
	})
}
