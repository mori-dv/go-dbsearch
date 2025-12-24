package go_dbsearch

import (
	"net/url"
	"strconv"
	"strings"
)

// ParseQueryWithOptions parses URL query parameters into a SearchQuery.
//
// Phase-4:
//   - Options is REQUIRED (to provide AllowedFields).
//   - Invalid filters/sorts are ignored (GET stays permissive).
func ParseQueryWithOptions(values url.Values, opts *Options) (SearchQuery, error) {
	v, err := NewValidatorFromOptions(opts)
	if err != nil {
		return SearchQuery{}, err
	}
	caster := NewValueCaster(opts)

	var filters []Filter
	var sorts []SortOption

	for key, vals := range values {
		if !strings.HasPrefix(key, "filter[") {
			continue
		}
		inner := strings.TrimSuffix(strings.TrimPrefix(key, "filter["), "]")
		if inner == "" {
			continue
		}

		parts := strings.SplitN(inner, ":", 2)
		field := strings.TrimSpace(parts[0])
		op := "="
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			op = strings.TrimSpace(parts[1])
		}

		if err := v.ValidateField(field); err != nil {
			continue
		}
		normOp, ok := NormalizeOperator(op)
		if !ok {
			continue
		}

		raw := ""
		if len(vals) > 0 {
			raw = vals[0]
		}

		value, ok := parseAndCastValue(field, normOp, raw, caster)
		if !ok {
			continue
		}

		filters = append(filters, Filter{Field: field, Op: normOp, Value: value})
	}

	sortStr := strings.TrimSpace(values.Get("sort"))
	if sortStr != "" {
		for _, part := range strings.Split(sortStr, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			dir := "ASC"
			field := part
			if strings.HasPrefix(part, "-") {
				dir = "DESC"
				field = strings.TrimPrefix(part, "-")
			}
			if err := v.ValidateField(field); err != nil {
				continue
			}
			sorts = append(sorts, SortOption{Field: field, Direction: dir})
		}
	}

	limit, _ := strconv.Atoi(values.Get("limit"))
	offset, _ := strconv.Atoi(values.Get("offset"))

	if opts != nil && opts.MaxLimit > 0 && limit > opts.MaxLimit {
		limit = opts.MaxLimit
	}

	return SearchQuery{
		Filters: filters,
		Sorts:   sorts,
		Pagination: Pagination{
			Limit:  limit,
			Offset: offset,
		},
	}, nil
}

func parseAndCastValue(field, op, raw string, caster *ValueCaster) (interface{}, bool) {
	switch op {
	case "IN":
		parts := splitCSV(raw)
		out := make([]interface{}, 0, len(parts))
		for _, p := range parts {
			cv, err := caster.CastFromString(field, p)
			if err != nil {
				return nil, false
			}
			out = append(out, cv)
		}
		return out, true
	case "BETWEEN":
		parts := splitCSV(raw)
		if len(parts) != 2 {
			return nil, false
		}
		lo, err := caster.CastFromString(field, parts[0])
		if err != nil {
			return nil, false
		}
		hi, err := caster.CastFromString(field, parts[1])
		if err != nil {
			return nil, false
		}
		return []interface{}{lo, hi}, true
	default:
		cv, err := caster.CastFromString(field, raw)
		if err != nil {
			return nil, false
		}
		return cv, true
	}
}
