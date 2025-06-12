package base

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// GenerationContext provides context and utilities for schema generation.
type GenerationContext struct {
	// IndentLevel tracks the current indentation level
	IndentLevel int

	// IndentString is the string used for indentation (default: "  ")
	IndentString string

	// Options contains generator-specific options
	Options map[string]interface{}

	// Metadata contains generation metadata and state
	Metadata map[string]interface{}

	// Path tracks the current schema path for error reporting
	Path []string

	// Identifiers tracks generated identifiers to avoid conflicts
	Identifiers map[string]int
}

// NewGenerationContext creates a new GenerationContext with defaults.
func NewGenerationContext() *GenerationContext {
	return &GenerationContext{
		IndentLevel:  0,
		IndentString: "  ",
		Options:      make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		Path:         make([]string, 0),
		Identifiers:  make(map[string]int),
	}
}

// Clone creates a deep copy of the GenerationContext.
func (c *GenerationContext) Clone() *GenerationContext {
	clone := &GenerationContext{
		IndentLevel:  c.IndentLevel,
		IndentString: c.IndentString,
		Options:      make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		Path:         make([]string, len(c.Path)),
		Identifiers:  make(map[string]int),
	}

	// Deep copy options
	for k, v := range c.Options {
		clone.Options[k] = v
	}

	// Deep copy metadata
	for k, v := range c.Metadata {
		clone.Metadata[k] = v
	}

	// Copy path
	copy(clone.Path, c.Path)

	// Copy identifiers
	for k, v := range c.Identifiers {
		clone.Identifiers[k] = v
	}

	return clone
}

// Indentation methods

// Indent returns the current indentation string.
func (c *GenerationContext) Indent() string {
	if c.IndentLevel <= 0 {
		return ""
	}
	return strings.Repeat(c.IndentString, c.IndentLevel)
}

// IndentBy returns an indentation string for a specific level.
func (c *GenerationContext) IndentBy(level int) string {
	if level <= 0 {
		return ""
	}
	return strings.Repeat(c.IndentString, level)
}

// PushIndent increases the indentation level.
func (c *GenerationContext) PushIndent() {
	c.IndentLevel++
}

// PopIndent decreases the indentation level.
func (c *GenerationContext) PopIndent() {
	if c.IndentLevel > 0 {
		c.IndentLevel--
	}
}

// WithIndent executes a function with increased indentation.
func (c *GenerationContext) WithIndent(fn func()) {
	c.PushIndent()
	defer c.PopIndent()
	fn()
}

// Path methods

// PushPath adds an element to the current path.
func (c *GenerationContext) PushPath(element string) {
	c.Path = append(c.Path, element)
}

// PopPath removes the last element from the current path.
func (c *GenerationContext) PopPath() string {
	if len(c.Path) == 0 {
		return ""
	}
	lastIndex := len(c.Path) - 1
	element := c.Path[lastIndex]
	c.Path = c.Path[:lastIndex]
	return element
}

// CurrentPath returns the current path as a dot-separated string.
func (c *GenerationContext) CurrentPath() string {
	return strings.Join(c.Path, ".")
}

// PathDepth returns the current path depth.
func (c *GenerationContext) PathDepth() int {
	return len(c.Path)
}

// WithPath executes a function with a path element added.
func (c *GenerationContext) WithPath(element string, fn func()) {
	c.PushPath(element)
	defer c.PopPath()
	fn()
}

// Option methods

// SetOption sets a generation option.
func (c *GenerationContext) SetOption(key string, value interface{}) {
	c.Options[key] = value
}

// GetOption gets a generation option.
func (c *GenerationContext) GetOption(key string) (interface{}, bool) {
	value, exists := c.Options[key]
	return value, exists
}

// GetOptionString gets a string option with a default value.
func (c *GenerationContext) GetOptionString(key, defaultValue string) string {
	if value, exists := c.Options[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetOptionBool gets a boolean option with a default value.
func (c *GenerationContext) GetOptionBool(key string, defaultValue bool) bool {
	if value, exists := c.Options[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// GetOptionInt gets an integer option with a default value.
func (c *GenerationContext) GetOptionInt(key string, defaultValue int) int {
	if value, exists := c.Options[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return defaultValue
}

// Metadata methods

// SetMetadata sets generation metadata.
func (c *GenerationContext) SetMetadata(key string, value interface{}) {
	c.Metadata[key] = value
}

// GetMetadata gets generation metadata.
func (c *GenerationContext) GetMetadata(key string) (interface{}, bool) {
	value, exists := c.Metadata[key]
	return value, exists
}

// Identifier methods

// UniqueIdentifier generates a unique identifier based on a base name.
func (c *GenerationContext) UniqueIdentifier(baseName string) string {
	// Sanitize the base name first
	sanitized := SanitizeIdentifier(baseName)

	// Check if we've seen this identifier before
	if count, exists := c.Identifiers[sanitized]; exists {
		// Increment the count and return a numbered version
		c.Identifiers[sanitized] = count + 1
		return fmt.Sprintf("%s%d", sanitized, count+1)
	}

	// First time seeing this identifier
	c.Identifiers[sanitized] = 0
	return sanitized
}

// ResetIdentifiers clears the identifier tracking.
func (c *GenerationContext) ResetIdentifiers() {
	c.Identifiers = make(map[string]int)
}

// Utility functions

// SanitizeIdentifier sanitizes a string to be a valid identifier.
// This is a generic implementation that can be overridden for specific languages.
func SanitizeIdentifier(s string) string {
	if s == "" {
		return "unnamed"
	}

	var result strings.Builder

	// Ensure first character is valid (letter or underscore)
	firstChar := rune(s[0])
	if unicode.IsLetter(firstChar) || firstChar == '_' {
		result.WriteRune(firstChar)
	} else {
		result.WriteRune('_')
	}

	// Process remaining characters
	for _, r := range s[1:] {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}

	identifier := result.String()

	// Ensure we don't return just underscores
	if strings.Trim(identifier, "_") == "" {
		return "unnamed"
	}

	return identifier
}

// EscapeString escapes a string for safe inclusion in generated code.
// This is a generic implementation that escapes common characters.
func EscapeString(s string) string {
	// Replace common escape sequences
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// EscapeJSONString escapes a string for JSON inclusion.
func EscapeJSONString(s string) string {
	// JSON-specific escaping
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\b", "\\b")
	s = strings.ReplaceAll(s, "\f", "\\f")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// ToCamelCase converts a string to camelCase.
func ToCamelCase(s string) string {
	if s == "" {
		return ""
	}

	// Split on common separators
	words := regexp.MustCompile(`[^a-zA-Z0-9]+`).Split(s, -1)

	var result strings.Builder

	for i, word := range words {
		if word == "" {
			continue
		}

		if i == 0 {
			// First word is lowercase
			result.WriteString(strings.ToLower(word))
		} else {
			// Subsequent words are capitalized
			result.WriteString(strings.Title(strings.ToLower(word)))
		}
	}

	return result.String()
}

// ToPascalCase converts a string to PascalCase.
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	// If the string is already in PascalCase (starts with uppercase and contains no separators), return as-is
	if unicode.IsUpper(rune(s[0])) && !regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(s) {
		return s
	}

	// Split on common separators
	words := regexp.MustCompile(`[^a-zA-Z0-9]+`).Split(s, -1)

	var result strings.Builder

	for _, word := range words {
		if word == "" {
			continue
		}

		// Capitalize first letter, keep rest as lowercase
		if len(word) > 0 {
			result.WriteRune(unicode.ToUpper(rune(word[0])))
			if len(word) > 1 {
				result.WriteString(strings.ToLower(word[1:]))
			}
		}
	}

	return result.String()
}

// ToSnakeCase converts a string to snake_case.
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	// Split on common separators and camelCase boundaries
	var words []string

	// First split on explicit separators
	parts := regexp.MustCompile(`[^a-zA-Z0-9]+`).Split(s, -1)

	// Then split camelCase/PascalCase words
	for _, part := range parts {
		if part == "" {
			continue
		}

		// Split on camelCase boundaries
		camelWords := regexp.MustCompile(`([a-z])([A-Z])`).ReplaceAllString(part, "${1}_${2}")
		words = append(words, strings.ToLower(camelWords))
	}

	return strings.Join(words, "_")
}

// ToKebabCase converts a string to kebab-case.
func ToKebabCase(s string) string {
	return strings.ReplaceAll(ToSnakeCase(s), "_", "-")
}

// Pluralize attempts to pluralize an English word.
// This is a simple implementation and may not handle all cases correctly.
func Pluralize(word string) string {
	if word == "" {
		return ""
	}

	lower := strings.ToLower(word)

	// Special cases
	irregulars := map[string]string{
		"child":  "children",
		"person": "people",
		"man":    "men",
		"woman":  "women",
		"tooth":  "teeth",
		"foot":   "feet",
		"mouse":  "mice",
		"goose":  "geese",
	}

	if plural, exists := irregulars[lower]; exists {
		return plural
	}

	// Common patterns
	if strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "sh") ||
		strings.HasSuffix(lower, "ch") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") {
		return word + "es"
	}

	if strings.HasSuffix(lower, "y") && len(word) > 1 {
		beforeY := word[len(word)-2]
		if !isVowel(rune(beforeY)) {
			return word[:len(word)-1] + "ies"
		}
	}

	if strings.HasSuffix(lower, "f") {
		return word[:len(word)-1] + "ves"
	}

	if strings.HasSuffix(lower, "fe") {
		return word[:len(word)-2] + "ves"
	}

	// Default: just add 's'
	return word + "s"
}

// isVowel checks if a character is a vowel.
func isVowel(r rune) bool {
	vowels := "aeiouAEIOU"
	return strings.ContainsRune(vowels, r)
}

// FormatComment formats a comment string for code generation.
func FormatComment(comment string, prefix string, width int) []string {
	if comment == "" {
		return nil
	}

	if width <= 0 {
		width = 80
	}

	lines := strings.Split(comment, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			result = append(result, prefix)
			continue
		}

		// Wrap long lines
		maxWidth := width - len(prefix) - 1
		if len(line) <= maxWidth {
			result = append(result, prefix+" "+line)
		} else {
			words := strings.Fields(line)
			var currentLine strings.Builder

			for _, word := range words {
				if currentLine.Len() == 0 {
					currentLine.WriteString(word)
				} else if currentLine.Len()+1+len(word) <= maxWidth {
					currentLine.WriteString(" " + word)
				} else {
					result = append(result, prefix+" "+currentLine.String())
					currentLine.Reset()
					currentLine.WriteString(word)
				}
			}

			if currentLine.Len() > 0 {
				result = append(result, prefix+" "+currentLine.String())
			}
		}
	}

	return result
}
