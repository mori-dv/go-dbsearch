package go_dbsearch

import "gorm.io/gorm"

type SearchQuery struct {
	Filters    []Filter
	Sorts      []SortOption
	Pagination Pagination
}

func Apply(db *gorm.DB, query SearchQuery) *gorm.DB {

	tx := db

	for _, filter := range query.Filters {
		tx = filter.Apply(tx)
	}

	for _, sort := range query.Sorts {
		tx = tx.Order(sort.Field + " " + sort.Direction)
	}

	if query.Pagination.Limit > 0 {
		tx = tx.Limit(query.Pagination.Limit)
	}
	if query.Pagination.Offset > 0 {
		tx = tx.Offset(query.Pagination.Offset)
	}

	return tx
}
