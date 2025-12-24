package go_dbsearch

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Filter represents a single filter condition.
//
// Field should be validated by a Validator (whitelist + safe identifier).
// Op can be canonical ("=", "LIKE", "IN", ...) or an alias ("eq", "like", ...).
// Value is always passed to GORM as a parameter (never interpolated directly into SQL).
type Filter struct {
	Field string      `json:"field"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

// Apply applies the filter to a GORM query and returns the updated query.
//
// This method performs minimal defense-in-depth (identifier character check + operator normalization),
// but callers should validate fields/operators strictly in handlers (especially for JSON search).
func (f Filter) Apply(db *gorm.DB) *gorm.DB {
	// Defense in depth: reject obviously unsafe identifiers.
	if strings.TrimSpace(f.Field) == "" || !safeFieldRe.MatchString(f.Field) {
		return db
	}

	op, ok := NormalizeOperator(f.Op)
	if !ok {
		return db
	}

	switch op {
	case "=":
		return db.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
	case "LIKE":
		return db.Where(fmt.Sprintf("%s LIKE ?", f.Field), fmt.Sprintf("%%%v%%", f.Value))
	case ">":
		return db.Where(fmt.Sprintf("%s > ?", f.Field), f.Value)
	case "<":
		return db.Where(fmt.Sprintf("%s < ?", f.Field), f.Value)
	case ">=":
		return db.Where(fmt.Sprintf("%s >= ?", f.Field), f.Value)
	case "<=":
		return db.Where(fmt.Sprintf("%s <= ?", f.Field), f.Value)
	case "IN":
		return db.Where(fmt.Sprintf("%s IN ?", f.Field), normalizeINValue(f.Value))
	case "BETWEEN":
		lo, hi, ok := normalizeBetweenValue(f.Value)
		if !ok {
			return db
		}
		return db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", f.Field), lo, hi)
	default:
		return db
	}
}

func normalizeINValue(v interface{}) interface{} {
	switch vv := v.(type) {
	case string:
		parts := splitCSV(vv)
		out := make([]interface{}, 0, len(parts))
		for _, p := range parts {
			out = append(out, p)
		}
		return out
	default:
		return v
	}
}

func normalizeBetweenValue(v interface{}) (interface{}, interface{}, bool) {
	switch vv := v.(type) {
	case string:
		parts := splitCSV(vv)
		if len(parts) != 2 {
			return nil, nil, false
		}
		return parts[0], parts[1], true
	case []interface{}:
		if len(vv) != 2 {
			return nil, nil, false
		}
		return vv[0], vv[1], true
	case []string:
		if len(vv) != 2 {
			return nil, nil, false
		}
		return vv[0], vv[1], true
	default:
		return nil, nil, false
	}
}

func splitCSV(s string) []string {
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
