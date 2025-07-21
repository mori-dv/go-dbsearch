package go_dbsearch

import "testing"

func TestPagination_Defaults(t *testing.T) {
	p := Pagination{}
	if p.Limit != 0 || p.Offset != 0 {
		t.Fatalf("Expected default Pagination to be zero, got: %+v", p)
	}
}

func TestPagination_Boundaries(t *testing.T) {
	p := Pagination{Limit: 100, Offset: 50}
	if p.Limit != 100 || p.Offset != 50 {
		t.Fatalf("Pagination boundaries failed, got: %+v", p)
	}
}
