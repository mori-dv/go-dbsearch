package go_dbsearch

import (
	"net/url"
	"testing"
)

func TestParseQuery_SingleFilter(t *testing.T) {
	v := url.Values{}
	v.Set("filter[name:=]", "Alice")
	q := ParseQuery(v)
	if len(q.Filters) != 1 || q.Filters[0].Field != "name" || q.Filters[0].Op != "=" || q.Filters[0].Value != "Alice" {
		t.Fatalf("Single filter parse failed: %+v", q.Filters)
	}
}

func TestParseQuery_MultipleFilters(t *testing.T) {
	v := url.Values{}
	v.Set("filter[name:=]", "Alice")
	v.Set("filter[age:>]=", "25")
	q := ParseQuery(v)
	if len(q.Filters) != 2 {
		t.Fatalf("Multiple filters parse failed: %+v", q.Filters)
	}
}

func TestParseQuery_SortAscDesc(t *testing.T) {
	v := url.Values{}
	v.Set("sort", "-age,name")
	q := ParseQuery(v)
	if len(q.Sorts) != 2 || q.Sorts[0].Field != "age" || q.Sorts[0].Direction != "DESC" || q.Sorts[1].Field != "name" || q.Sorts[1].Direction != "ASC" {
		t.Fatalf("Sort parse failed: %+v", q.Sorts)
	}
}

func TestParseQuery_Pagination(t *testing.T) {
	v := url.Values{}
	v.Set("limit", "10")
	v.Set("offset", "5")
	q := ParseQuery(v)
	if q.Pagination.Limit != 10 || q.Pagination.Offset != 5 {
		t.Fatalf("Pagination parse failed: %+v", q.Pagination)
	}
}

func TestParseQuery_InvalidInput(t *testing.T) {
	v := url.Values{}
	v.Set("filter[bad]", "x")
	q := ParseQuery(v)
	if len(q.Filters) != 0 {
		t.Fatalf("Invalid filter should be ignored: %+v", q.Filters)
	}
}
