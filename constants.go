package schema

import "reflect"

// Schema Generation Constants
// These constants are used by the schema generators for realistic data generation

// Common property names for realistic schema generation
var CommonPropertyNames = []string{
	"id", "name", "email", "age", "title", "description", "status", "type",
	"createdAt", "updatedAt", "deletedAt", "userId", "username", "password",
	"firstName", "lastName", "fullName", "address", "city", "country",
	"phone", "website", "avatar", "bio", "role", "permissions", "settings",
	"metadata", "tags", "category", "priority", "score", "rating", "count",
	"amount", "price", "currency", "discount", "tax", "total", "balance",
	"startDate", "endDate", "duration", "deadline", "reminder", "notes",
	"content", "body", "summary", "keywords", "author", "editor", "reviewer",
	"version", "revision", "branch", "commit", "hash", "checksum",
	"enabled", "active", "visible", "public", "private", "archived",
	"config", "options", "parameters", "attributes", "properties", "data",
}

// Common string formats for schema generation
var CommonStringFormats = []string{
	"email", "uuid", "uri", "url", "date", "time", "date-time",
	"password", "ipv4", "ipv6", "hostname", "json-pointer",
}

// Common regex patterns for schema validation
var CommonPatterns = []string{
	"^[a-zA-Z0-9]+$",               // Alphanumeric
	"^[0-9]{3}-[0-9]{3}-[0-9]{4}$", // Phone number
	"^[A-Z]{2}[0-9]{4}$",           // Code format
	"^#[0-9A-Fa-f]{6}$",            // Hex color
	"^[a-z_]+$",                    // Snake case
	"^[A-Z][a-z]+$",                // Title case
}

// Sample descriptions for schema metadata
var SampleDescriptions = []string{
	"A unique identifier for the resource",
	"The name or title of the item",
	"A brief description of the content",
	"The current status of the operation",
	"Timestamp when the record was created",
	"User-provided configuration settings",
	"Additional metadata for the object",
	"The priority level (1-10 scale)",
	"Email address for notifications",
	"Indicates whether the item is active",
}

// Sample titles for schema metadata
var SampleTitles = []string{
	"Identifier", "Name", "Title", "Description", "Status", "Type",
	"Created Date", "Updated Date", "User ID", "Email Address",
	"Configuration", "Settings", "Metadata", "Properties", "Data",
	"Count", "Amount", "Price", "Rating", "Score", "Priority",
}

// Reflection Constants
// These constants are used by the reflection utilities

// ErrorInterface represents the error interface type for reflection
var ErrorInterface = reflect.TypeOf((*error)(nil)).Elem()

// Default Values
// These are default values used throughout the package

const (
	// Default generator configuration values
	DefaultMaxDepth     = 5
	DefaultMaxItems     = 10
	DefaultMinItems     = 1
	DefaultStringMinLen = 3
	DefaultStringMaxLen = 20
	DefaultNumberMin    = 0.0
	DefaultNumberMax    = 1000.0
	DefaultIntegerMin   = int64(0)
	DefaultIntegerMax   = int64(1000)
	DefaultOptionalProb = 0.7
	DefaultGenerateProb = 0.8
)

// Character sets for random generation
const (
	AlphaNumericChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	AlphaChars        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	NumericChars      = "0123456789"
	HexChars          = "0123456789abcdef"
)

// Validation Error Messages
const (
	ErrMsgRequired      = "field is required"
	ErrMsgInvalidType   = "invalid type"
	ErrMsgInvalidFormat = "invalid format"
	ErrMsgTooShort      = "value too short"
	ErrMsgTooLong       = "value too long"
	ErrMsgTooSmall      = "value too small"
	ErrMsgTooLarge      = "value too large"
	ErrMsgInvalidEnum   = "value not in allowed enum"
	ErrMsgPatternFailed = "value does not match pattern"
)
