package go_dbsearch

import (
	"net/url"
	"strconv"
	"strings"
)

func ParseQuery(values url.Values) SearchQuery {
	var filters []Filter
	var sorts []SortOption

	// filters
	for key, val := range values {
		if strings.HasPrefix(key, "filter[") {
			fieldOp := strings.TrimSuffix(strings.TrimPrefix(key, "filter["), "]") // name:eq
			parts := strings.Split(fieldOp, ":")
			if len(parts) != 2 || !IsFieldAllowed(parts[0]) {
				continue
			}
			filters = append(filters, Filter{
				Field: parts[0],
				Op:    parts[1],
				Value: val[0],
			})
		}
	}

	// sorting
	if sortStr := values.Get("sort"); sortStr != "" {
		for _, s := range strings.Split(sortStr, ",") {
			dir := "ASC"
			field := s
			if strings.HasPrefix(s, "-") {
				dir = "DESC"
				field = s[1:]
			}
			if IsFieldAllowed(field) {
				sorts = append(sorts, SortOption{Field: field, Direction: dir})
			}
		}
	}

	// pagination
	limit, _ := strconv.Atoi(values.Get("limit"))
	offset, _ := strconv.Atoi(values.Get("offset"))

	return SearchQuery{
		Filters: filters,
		Sorts:   sorts,
		Pagination: Pagination{
			Limit:  limit,
			Offset: offset,
		},
	}
}
