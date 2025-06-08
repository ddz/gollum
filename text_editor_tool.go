package main

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
