package utils

import "defs.dev/schema/core"

// GetAnnotation returns the first annotation with the given name.
func GetAnnotation(s core.Schema, name string) (core.Annotation, bool) {
	if s == nil {
		return nil, false
	}
	for _, ann := range s.Annotations() {
		if ann.Name() == name {
			return ann, true
		}
	}
	return nil, false
}

// HasAnnotation checks if a schema has an annotation.
func HasAnnotation(s core.Schema, name string) bool {
	_, ok := GetAnnotation(s, name)
	return ok
}

// AnnotationsByName returns all annotations with the specified name.
func AnnotationsByName(s core.Schema, name string) []core.Annotation {
	var result []core.Annotation
	if s == nil {
		return result
	}
	for _, ann := range s.Annotations() {
		if ann.Name() == name {
			result = append(result, ann)
		}
	}
	return result
}
