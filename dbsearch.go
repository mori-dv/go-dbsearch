package go_dbsearch

import "gorm.io/gorm"

// SearchQuery is the internal representation of a parsed search request.
type SearchQuery struct {
	Filters    []Filter
	Sorts      []SortOption
	Pagination Pagination
}

// ApplyWithOptions applies filters/sorts/pagination using per-handler Options.
//
// Phase-4: Options is required (AllowedFields must be set).
func ApplyWithOptions(db *gorm.DB, query SearchQuery, opts *Options) *gorm.DB {
	// Filters already include safe field names because parsing required options+validator.
	tx := db

	for _, filter := range query.Filters {
		tx = filter.Apply(tx)
	}

	// Sort validation (defense-in-depth)
	v, err := NewValidatorFromOptions(opts)
	if err == nil {
		for _, sort := range query.Sorts {
			norm, err := v.ValidateSortOption(sort)
			if err != nil {
				continue
			}
			tx = tx.Order(norm.Field + " " + norm.Direction)
		}
	}

	limit := query.Pagination.Limit
	if opts != nil && opts.MaxLimit > 0 && limit > opts.MaxLimit {
		limit = opts.MaxLimit
	}
	if limit > 0 {
		tx = tx.Limit(limit)
	}
	if query.Pagination.Offset > 0 {
		tx = tx.Offset(query.Pagination.Offset)
	}

	return tx
}
