package go_dbsearch

import (
	"net/url"
	"testing"
)

func TestParseQueryWithOptions_IN_BETWEEN_Casting(t *testing.T) {
	opts := NewOptions([]string{"age", "created_at", "status"})
	opts.WithFieldTypes(map[string]FieldType{
		"age":        FieldTypeInt,
		"created_at": FieldTypeDate,
		"status":     FieldTypeString,
	})

	values := url.Values{}
	values.Set("filter[age:in]", "18,21")
	values.Set("filter[created_at:between]", "2023-01-01,2023-12-31")
	values.Set("filter[status:eq]", "active")

	q, err := ParseQueryWithOptions(values, opts)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(q.Filters) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(q.Filters))
	}

	// age IN -> []interface{}{18,21}
	var foundAge bool
	for _, f := range q.Filters {
		if f.Field == "age" && f.Op == "IN" {
			foundAge = true
			s, ok := f.Value.([]interface{})
			if !ok || len(s) != 2 {
				t.Fatalf("age IN expected []interface{} len=2, got %T %#v", f.Value, f.Value)
			}
			if s[0].(int) != 18 || s[1].(int) != 21 {
				t.Fatalf("age IN values mismatch: %#v", s)
			}
		}
	}
	if !foundAge {
		t.Fatalf("age IN filter not found")
	}

	// created_at BETWEEN -> []interface{}{time, time}
	var foundBetween bool
	for _, f := range q.Filters {
		if f.Field == "created_at" && f.Op == "BETWEEN" {
			foundBetween = true
			pair, ok := f.Value.([]interface{})
			if !ok || len(pair) != 2 {
				t.Fatalf("between expected []interface{} len=2, got %T %#v", f.Value, f.Value)
			}
		}
	}
	if !foundBetween {
		t.Fatalf("created_at BETWEEN filter not found")
	}
}
