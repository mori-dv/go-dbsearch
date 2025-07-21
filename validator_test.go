package go_dbsearch

import "testing"

func TestIsFieldAllowed_Allowed(t *testing.T) {
	AllowedFields["foo"] = true
	if !IsFieldAllowed("foo") {
		t.Fatalf("Expected field to be allowed")
	}
}

func TestIsFieldAllowed_Disallowed(t *testing.T) {
	AllowedFields["bar"] = false
	if IsFieldAllowed("bar") {
		t.Fatalf("Expected field to be disallowed")
	}
}
