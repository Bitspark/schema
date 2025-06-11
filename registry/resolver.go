package registry

import (
	"defs.dev/schema"
)

// resolveParameters walks a schema tree and replaces ParameterRef nodes with concrete schemas
func resolveParameters(s schema.Schema, params map[string]schema.Schema, visiting map[string]bool) (schema.Schema, error) {
	if s == nil {
		return nil, nil
	}

	switch typed := s.(type) {
	case *ParameterRef:
		// Replace parameter with concrete schema
		concrete, exists := params[typed.Name()]
		if !exists {
			return nil, &ParameterError{
				Parameter: typed.Name(),
				Expected:  "concrete schema",
				Actual:    "missing",
			}
		}
		// Recursively resolve the concrete schema in case it has parameters too
		return resolveParameters(concrete, params, visiting)

	case *SchemaRef:
		// For SchemaRef, we need to resolve it first, then resolve any parameters in the result
		resolved, err := typed.Resolve()
		if err != nil {
			return nil, err
		}
		return resolveParameters(resolved, params, visiting)

	default:
		// For other schema types, we need to recursively resolve their children
		return resolveSchemaChildren(s, params, visiting)
	}
}

// resolveSchemaChildren handles resolving parameters in child schemas for different schema types
func resolveSchemaChildren(s schema.Schema, params map[string]schema.Schema, visiting map[string]bool) (schema.Schema, error) {
	switch typed := s.(type) {
	case *schema.ObjectSchema:
		return resolveObjectSchema(typed, params, visiting)
	case *schema.ArraySchema:
		return resolveArraySchema(typed, params, visiting)
	default:
		// For basic types (string, number, boolean, etc.) that don't have children,
		// just return them as-is since they can't contain parameters
		return s, nil
	}
}

// resolveObjectSchema resolves parameters in object schema properties
func resolveObjectSchema(obj *schema.ObjectSchema, params map[string]schema.Schema, visiting map[string]bool) (schema.Schema, error) {
	// We need to access the internal structure, but since ObjectSchema is from the schema package,
	// we can't directly access its fields. We'll need to work with the public interface.
	// For now, let's return the object as-is and add proper resolution later.
	// This is a limitation that would need to be addressed in a real implementation.

	// TODO: This requires either:
	// 1. Adding resolution methods to the schema package, or
	// 2. Making schema fields accessible, or
	// 3. Adding a visitor pattern to the Schema interface

	return obj, nil
}

// resolveArraySchema resolves parameters in array item schemas
func resolveArraySchema(arr *schema.ArraySchema, params map[string]schema.Schema, visiting map[string]bool) (schema.Schema, error) {
	// Same limitation as ObjectSchema - we can't access the internal itemSchema field
	// This would need to be addressed in the schema package design

	return arr, nil
}

// detectCircularReference checks for circular references in schema resolution
func detectCircularReference(name string, visiting map[string]bool) error {
	if visiting[name] {
		// Build the path for the error
		path := make([]string, 0, len(visiting)+1)
		for n := range visiting {
			path = append(path, n)
		}
		path = append(path, name)
		return NewCircularRefError(path)
	}
	return nil
}

// validateParameters checks that all required parameters are provided and no extra ones are given
func validateParameters(schemaName string, requiredParams []string, providedParams map[string]schema.Schema) error {
	var missing []string
	var extra []string

	// Check for missing parameters
	for _, required := range requiredParams {
		if _, exists := providedParams[required]; !exists {
			missing = append(missing, required)
		}
	}

	// Check for extra parameters
	for provided := range providedParams {
		found := false
		for _, required := range requiredParams {
			if provided == required {
				found = true
				break
			}
		}
		if !found {
			extra = append(extra, provided)
		}
	}

	if len(missing) > 0 || len(extra) > 0 {
		return NewInvalidParamsError(schemaName, missing, extra)
	}

	return nil
}
