package core

import (
	"io/fs"
	"time"
)

// FileValue represents a file as a validatable value
type FileValue interface {
	Value[FileInfo]

	// File-specific value methods
	Path() string
	Content() ([]byte, error)
	Extension() string
	Size() int64
	ModTime() time.Time
	Mode() fs.FileMode

	// Validation context
	RelativePath() string
	ProjectRoot() string
	Exists() bool
}

// DirectoryValue represents a directory as a validatable value
type DirectoryValue interface {
	ComplexValue[DirectoryInfo]

	// Directory-specific value methods
	Path() string
	Files() ([]FileValue, error)
	Subdirectories() ([]DirectoryValue, error)
	AllEntries() ([]FileSystemEntry, error)

	// Validation context
	RelativePath() string
	ProjectRoot() string
	Exists() bool
	Structure() (map[string]FileSystemEntry, error)
}

// NodeValue represents a project node as a validatable value
type NodeValue interface {
	ComplexValue[NodeInfo]

	// Node-specific value methods
	Path() string
	Files() ([]FileValue, error)
	Directories() ([]DirectoryValue, error)
	Children() ([]NodeValue, error)
	Parent() NodeValue

	// Validation context
	Level() int
	Metadata() map[string]any
	ProjectRoot() string
	RelativePath() string

	// Node tree navigation
	FindFile(pattern string) ([]FileValue, error)
	FindDirectory(pattern string) ([]DirectoryValue, error)
	FindNode(pattern string) ([]NodeValue, error)
}

// FileSystemEntry represents any file system entry (file or directory)
type FileSystemEntry interface {
	Name() string
	Path() string
	IsDir() bool
	IsFile() bool
	Size() int64
	ModTime() time.Time
	Mode() fs.FileMode
}

// FileInfo provides information about a file
type FileInfo interface {
	FileSystemEntry

	// File-specific information
	Extension() string
	MimeType() string
	Encoding() string
	LineCount() (int, error)
}

// DirectoryInfo provides information about a directory
type DirectoryInfo interface {
	FileSystemEntry

	// Directory-specific information
	EntryCount() (int, error)
	FileCount() (int, error)
	DirectoryCount() (int, error)
	TotalSize() (int64, error)
}

// NodeInfo provides information about a project node
type NodeInfo interface {
	FileSystemEntry

	// Node-specific information
	NodeType() string
	ConfigFiles() []string
	HasConfig(name string) bool
	GetConfig(name string) ([]byte, error)
}
