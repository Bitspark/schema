package core

// ValidationSchema extends the core Schema interface for file system validation
type ValidationSchema interface {
	Schema // Inherits: Validate, ToJSONSchema, Type, Metadata, etc.

	// Validation-specific introspection
	Patterns() []string
	ValidatorRefs() []ValidatorRef
	Inheritance() []string
	ConfigSchema() Schema // Schema for validator configuration
}

// ValidatorRef represents a reference to a validator with its configuration
type ValidatorRef struct {
	Name   string
	Config map[string]any
	Schema Schema // Schema for the validator's configuration
}

// FileValidationSchema validates file system files
type FileValidationSchema interface {
	ValidationSchema

	// File-specific introspection
	FilePatterns() []string
	SupportedExtensions() []string
	RequiredValidators() []string
	ContentValidation() Schema // Schema for file content validation
}

// DirectoryValidationSchema validates directory structure
type DirectoryValidationSchema interface {
	ValidationSchema

	// Directory-specific introspection
	DirectoryPatterns() []string
	RequiredFiles() map[string]FileValidationSchema
	RequiredDirectories() map[string]DirectoryValidationSchema
	OptionalFiles() map[string]FileValidationSchema
	OptionalDirectories() map[string]DirectoryValidationSchema
	AllowAdditionalFiles() bool
	AllowAdditionalDirectories() bool
}

// NodeValidationSchema validates complete project nodes
type NodeValidationSchema interface {
	ValidationSchema

	// Node-specific introspection
	FileSchemas() []FileValidationSchema
	DirectorySchemas() []DirectoryValidationSchema
	CompletenessRules() []CompletenessRule
	ConsistencyRules() []ConsistencyRule
	InheritanceRules() []InheritanceRule
}

// CompletenessRule defines what must be present in a node
type CompletenessRule struct {
	Name        string
	Description string
	Required    bool
	Pattern     string
	Schema      Schema
}

// ConsistencyRule defines consistency requirements across files/directories
type ConsistencyRule struct {
	Name        string
	Description string
	Sources     []string // File/directory patterns to check
	Rule        string   // Consistency rule expression
	Schema      Schema   // Schema for rule configuration
}

// InheritanceRule defines how validation rules are inherited
type InheritanceRule struct {
	Name        string
	Description string
	Inherits    []string // Parent schema names
	Overrides   map[string]any
	Schema      Schema // Schema for inheritance configuration
}
