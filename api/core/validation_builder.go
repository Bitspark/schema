package core

// ValidationSchemaBuilder provides a fluent interface for building validation schemas
type ValidationSchemaBuilder interface {
	// Basic properties
	Name(name string) ValidationSchemaBuilder
	Description(description string) ValidationSchemaBuilder
	Version(version string) ValidationSchemaBuilder

	// Pattern matching
	Pattern(pattern string) ValidationSchemaBuilder
	Patterns(patterns ...string) ValidationSchemaBuilder

	// Validator configuration
	Validator(name string, config Schema) ValidationSchemaBuilder
	RequiredValidator(name string, config Schema) ValidationSchemaBuilder
	OptionalValidator(name string, config Schema) ValidationSchemaBuilder

	// Inheritance
	Extends(parent string) ValidationSchemaBuilder
	Inherits(parents ...string) ValidationSchemaBuilder

	// Configuration schema
	ConfigSchema(schema Schema) ValidationSchemaBuilder

	// Metadata
	Tag(tag string) ValidationSchemaBuilder
	Tags(tags ...string) ValidationSchemaBuilder
	Property(key, value string) ValidationSchemaBuilder

	// Build the schema
	Build() ValidationSchema
}

// FileValidationSchemaBuilder builds file validation schemas
type FileValidationSchemaBuilder interface {
	// Basic properties
	Name(name string) FileValidationSchemaBuilder
	Description(description string) FileValidationSchemaBuilder
	Version(version string) FileValidationSchemaBuilder

	// Pattern matching
	Pattern(pattern string) FileValidationSchemaBuilder
	Patterns(patterns ...string) FileValidationSchemaBuilder

	// Validator configuration
	Validator(name string, config Schema) FileValidationSchemaBuilder
	RequiredValidator(name string, config Schema) FileValidationSchemaBuilder
	OptionalValidator(name string, config Schema) FileValidationSchemaBuilder

	// Inheritance
	Extends(parent string) FileValidationSchemaBuilder
	Inherits(parents ...string) FileValidationSchemaBuilder

	// Configuration schema
	ConfigSchema(schema Schema) FileValidationSchemaBuilder

	// Metadata
	Tag(tag string) FileValidationSchemaBuilder
	Tags(tags ...string) FileValidationSchemaBuilder
	Property(key, value string) FileValidationSchemaBuilder

	// File-specific configuration
	Extension(ext string) FileValidationSchemaBuilder
	Extensions(exts ...string) FileValidationSchemaBuilder
	MimeType(mimeType string) FileValidationSchemaBuilder
	MimeTypes(mimeTypes ...string) FileValidationSchemaBuilder

	// Content validation
	ContentSchema(schema Schema) FileValidationSchemaBuilder
	MaxSize(bytes int64) FileValidationSchemaBuilder
	MinSize(bytes int64) FileValidationSchemaBuilder

	// File-specific validators
	FormatValidator(name string, config Schema) FileValidationSchemaBuilder
	ContentValidator(name string, config Schema) FileValidationSchemaBuilder

	// Build file validation schema
	Build() FileValidationSchema
}

// DirectoryValidationSchemaBuilder builds directory validation schemas
type DirectoryValidationSchemaBuilder interface {
	ValidationSchemaBuilder

	// Required structure
	RequiredFile(name string, schema FileValidationSchema) DirectoryValidationSchemaBuilder
	RequiredDirectory(name string, schema DirectoryValidationSchema) DirectoryValidationSchemaBuilder

	// Optional structure
	OptionalFile(name string, schema FileValidationSchema) DirectoryValidationSchemaBuilder
	OptionalDirectory(name string, schema DirectoryValidationSchema) DirectoryValidationSchemaBuilder

	// Additional entries
	AllowAdditionalFiles(allow bool) DirectoryValidationSchemaBuilder
	AllowAdditionalDirectories(allow bool) DirectoryValidationSchemaBuilder
	AdditionalFileSchema(schema FileValidationSchema) DirectoryValidationSchemaBuilder
	AdditionalDirectorySchema(schema DirectoryValidationSchema) DirectoryValidationSchemaBuilder

	// Structure validators
	StructureValidator(name string, config Schema) DirectoryValidationSchemaBuilder
	OrganizationValidator(name string, config Schema) DirectoryValidationSchemaBuilder

	// Build directory validation schema
	Build() DirectoryValidationSchema
}

// NodeValidationSchemaBuilder builds node validation schemas
type NodeValidationSchemaBuilder interface {
	ValidationSchemaBuilder

	// File and directory schemas
	FileSchema(schema FileValidationSchema) NodeValidationSchemaBuilder
	DirectorySchema(schema DirectoryValidationSchema) NodeValidationSchemaBuilder

	// Completeness rules
	CompletenessRule(rule CompletenessRule) NodeValidationSchemaBuilder
	RequireFile(pattern string, schema FileValidationSchema) NodeValidationSchemaBuilder
	RequireDirectory(pattern string, schema DirectoryValidationSchema) NodeValidationSchemaBuilder

	// Consistency rules
	ConsistencyRule(rule ConsistencyRule) NodeValidationSchemaBuilder
	ConsistentNaming(patterns []string, rule string) NodeValidationSchemaBuilder
	ConsistentContent(patterns []string, schema Schema) NodeValidationSchemaBuilder

	// Inheritance rules
	InheritanceRule(rule InheritanceRule) NodeValidationSchemaBuilder
	InheritsFrom(parent string, overrides map[string]any) NodeValidationSchemaBuilder

	// Node-specific validators
	CompletenessValidator(name string, config Schema) NodeValidationSchemaBuilder
	ConsistencyValidator(name string, config Schema) NodeValidationSchemaBuilder

	// Build node validation schema
	Build() NodeValidationSchema
}
