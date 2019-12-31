package main

// UseString is safe way to use a string given a pointer to a string
func UseString(s *string) string {
	if s == nil {
		temp := "" // *string cannot be initialized
		s = &temp  // in one statement
	}
	return *s // safe to dereference the *string
}

// UsePointer allows for construction of pointer in one line
func UsePointer(s string) *string {
	return &s
}
