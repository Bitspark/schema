package engine

import (
	"fmt"
	"regexp"
)

// SimpleReference is the default implementation of SchemaReference
type SimpleReference struct {
	name      string
	namespace string
	version   string
}

// NewReference creates a simple schema reference with just a name
func NewReference(name string) SchemaReference {
	return &SimpleReference{
		name:      name,
		namespace: "",
		version:   "",
	}
}

// NewNamespacedReference creates a schema reference with namespace and name
func NewNamespacedReference(namespace, name string) SchemaReference {
	return &SimpleReference{
		name:      name,
		namespace: namespace,
		version:   "",
	}
}

// NewVersionedReference creates a schema reference with namespace, name, and version
func NewVersionedReference(namespace, name, version string) SchemaReference {
	return &SimpleReference{
		name:      name,
		namespace: namespace,
		version:   version,
	}
}

// ParseReference parses a reference string into a SchemaReference
// Supported formats:
//   - "name"                    -> name only
//   - "namespace:name"          -> namespaced
//   - "name@version"            -> versioned
//   - "namespace:name@version"  -> fully qualified
func ParseReference(ref string) (SchemaReference, error) {
	if ref == "" {
		return nil, fmt.Errorf("reference string cannot be empty")
	}

	// Pattern: (namespace:)?(name)(@version)?
	pattern := `^(?:([a-zA-Z0-9_-]+):)?([a-zA-Z0-9_-]+)(?:@([a-zA-Z0-9._-]+))?$`
	regex := regexp.MustCompile(pattern)

	matches := regex.FindStringSubmatch(ref)
	if matches == nil {
		return nil, fmt.Errorf("invalid reference format: %s", ref)
	}

	namespace := matches[1] // Can be empty
	name := matches[2]      // Required
	version := matches[3]   // Can be empty

	if name == "" {
		return nil, fmt.Errorf("reference must have a name: %s", ref)
	}

	return &SimpleReference{
		name:      name,
		namespace: namespace,
		version:   version,
	}, nil
}

// Implementation of SchemaReference interface

func (r *SimpleReference) Name() string {
	return r.name
}

func (r *SimpleReference) Namespace() string {
	return r.namespace
}

func (r *SimpleReference) Version() string {
	return r.version
}

func (r *SimpleReference) FullName() string {
	parts := []string{}

	if r.namespace != "" {
		parts = append(parts, r.namespace+":"+r.name)
	} else {
		parts = append(parts, r.name)
	}

	if r.version != "" {
		parts[0] = parts[0] + "@" + r.version
	}

	return parts[0]
}

func (r *SimpleReference) IsVersioned() bool {
	return r.version != ""
}

func (r *SimpleReference) IsNamespaced() bool {
	return r.namespace != ""
}

func (r *SimpleReference) Validate() error {
	if r.name == "" {
		return fmt.Errorf("reference name cannot be empty")
	}

	// Validate name format
	if !isValidIdentifier(r.name) {
		return fmt.Errorf("invalid name format: %s", r.name)
	}

	// Validate namespace format if present
	if r.namespace != "" && !isValidIdentifier(r.namespace) {
		return fmt.Errorf("invalid namespace format: %s", r.namespace)
	}

	// Validate version format if present
	if r.version != "" && !isValidVersion(r.version) {
		return fmt.Errorf("invalid version format: %s", r.version)
	}

	return nil
}

// String returns the full reference string representation
func (r *SimpleReference) String() string {
	return r.FullName()
}

// Equals compares two references for equality
func (r *SimpleReference) Equals(other SchemaReference) bool {
	if other == nil {
		return false
	}

	return r.name == other.Name() &&
		r.namespace == other.Namespace() &&
		r.version == other.Version()
}

// Convenience functions for creating references

// Ref creates a simple reference with just a name
func Ref(name string) SchemaReference {
	return NewReference(name)
}

// RefNS creates a namespaced reference
func RefNS(namespace, name string) SchemaReference {
	return NewNamespacedReference(namespace, name)
}

// RefVer creates a versioned reference (can be namespaced or not)
func RefVer(namespace, name, version string) SchemaReference {
	return NewVersionedReference(namespace, name, version)
}

// Helper functions for validation

func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// Allow alphanumeric, underscore, and hyphen
	pattern := `^[a-zA-Z0-9_-]+$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

func isValidVersion(s string) bool {
	if s == "" {
		return false
	}

	// Allow semantic versioning plus additional characters
	pattern := `^[a-zA-Z0-9._-]+$`
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// ReferenceSet provides utilities for working with collections of references
type ReferenceSet struct {
	refs map[string]SchemaReference
}

// NewReferenceSet creates a new reference set
func NewReferenceSet() *ReferenceSet {
	return &ReferenceSet{
		refs: make(map[string]SchemaReference),
	}
}

// Add adds a reference to the set
func (rs *ReferenceSet) Add(ref SchemaReference) {
	rs.refs[ref.FullName()] = ref
}

// Contains checks if the set contains a reference
func (rs *ReferenceSet) Contains(ref SchemaReference) bool {
	_, exists := rs.refs[ref.FullName()]
	return exists
}

// Remove removes a reference from the set
func (rs *ReferenceSet) Remove(ref SchemaReference) {
	delete(rs.refs, ref.FullName())
}

// List returns all references in the set
func (rs *ReferenceSet) List() []SchemaReference {
	refs := make([]SchemaReference, 0, len(rs.refs))
	for _, ref := range rs.refs {
		refs = append(refs, ref)
	}
	return refs
}

// Size returns the number of references in the set
func (rs *ReferenceSet) Size() int {
	return len(rs.refs)
}

// Clear removes all references from the set
func (rs *ReferenceSet) Clear() {
	rs.refs = make(map[string]SchemaReference)
}

// FilterByNamespace returns references matching the given namespace
func (rs *ReferenceSet) FilterByNamespace(namespace string) []SchemaReference {
	var filtered []SchemaReference
	for _, ref := range rs.refs {
		if ref.Namespace() == namespace {
			filtered = append(filtered, ref)
		}
	}
	return filtered
}

// FilterByVersion returns references matching the given version
func (rs *ReferenceSet) FilterByVersion(version string) []SchemaReference {
	var filtered []SchemaReference
	for _, ref := range rs.refs {
		if ref.Version() == version {
			filtered = append(filtered, ref)
		}
	}
	return filtered
}
