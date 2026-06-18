package handlers

// strPtr returns a pointer to the string, or nil if the string is empty.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// safeStr returns the value of a string pointer, or empty string if nil.
func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
