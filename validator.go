package go_dbsearch

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// safeFieldRe is a defense-in-depth check to prevent SQL fragments from being injected as an identifier.
// It permits identifiers like: "name", "users.email", "created_at".
var safeFieldRe = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)

// Validator validates fields, operators and sorts against an allowlist.
type Validator struct {
	allowed map[string]struct{}
}

// NewValidator creates a validator from a set of allowed fields.
// Passing nil/empty allowlist makes all fields invalid (safe default).
func NewValidator(allowedFields map[string]struct{}) *Validator {
	m := make(map[string]struct{}, len(allowedFields))
	for k := range allowedFields {
		m[k] = struct{}{}
	}
	return &Validator{allowed: m}
}

// NewValidatorFromOptions creates a validator from Options.
// Options.AllowedFields is REQUIRED in Phase-4.
func NewValidatorFromOptions(opts *Options) (*Validator, error) {
	if opts == nil {
		return nil, errors.New("options is required (phase-4): AllowedFields must be provided")
	}
	if opts.AllowedFields == nil || len(opts.AllowedFields) == 0 {
		return nil, errors.New("AllowedFields is required (phase-4): provide a non-empty allowlist")
	}
	return NewValidator(opts.AllowedFields), nil
}

// ValidateField validates that a field is safe to interpolate as a SQL identifier and is whitelisted.
func (v *Validator) ValidateField(field string) error {
	field = strings.TrimSpace(field)
	if field == "" {
		return errors.New("field is empty")
	}
	if !safeFieldRe.MatchString(field) {
		return fmt.Errorf("field contains invalid characters: %q", field)
	}
	if _, ok := v.allowed[field]; !ok {
		return fmt.Errorf("field is not allowed: %q", field)
	}
	return nil
}

// NormalizeOperator converts operator aliases into canonical SQL operators.
// Supported canonical ops: =, >, <, >=, <=, LIKE, IN, BETWEEN.
// Supported aliases: eq, gt, lt, gte, lte, like, in, between (case-insensitive).
func NormalizeOperator(op string) (string, bool) {
	op = strings.TrimSpace(op)
	if op == "" {
		return "", false
	}
	switch strings.ToLower(op) {
	case "eq", "=":
		return "=", true
	case "gt", ">":
		return ">", true
	case "lt", "<":
		return "<", true
	case "gte", ">=":
		return ">=", true
	case "lte", "<=":
		return "<=", true
	case "like":
		return "LIKE", true
	case "in":
		return "IN", true
	case "between":
		return "BETWEEN", true
	default:
		s := strings.ToUpper(op)
		switch s {
		case "=", ">", "<", ">=", "<=", "LIKE", "IN", "BETWEEN":
			return s, true
		default:
			return "", false
		}
	}
}

// ValidateOperator validates and canonicalizes an operator.
func ValidateOperator(op string) (string, error) {
	n, ok := NormalizeOperator(op)
	if !ok {
		return "", fmt.Errorf("operator is not allowed: %q", op)
	}
	return n, nil
}

// NormalizeSortDirection normalizes sort direction. It accepts "asc"/"desc" in any casing.
func NormalizeSortDirection(direction string) (string, bool) {
	s := strings.TrimSpace(direction)
	if s == "" {
		return "", false
	}
	s = strings.ToUpper(s)
	return s, s == "ASC" || s == "DESC"
}

// ValidateSortOption validates and normalizes a sort option.
func (v *Validator) ValidateSortOption(opt SortOption) (SortOption, error) {
	if err := v.ValidateField(opt.Field); err != nil {
		return SortOption{}, err
	}
	dir, ok := NormalizeSortDirection(opt.Direction)
	if !ok {
		return SortOption{}, fmt.Errorf("invalid sort direction: %q", opt.Direction)
	}
	opt.Direction = dir
	return opt, nil
}

// ValidateFilter validates a filter (field + operator) and normalizes Op in-place.
func (v *Validator) ValidateFilter(f *Filter) error {
	if f == nil {
		return nil
	}
	if err := v.ValidateField(f.Field); err != nil {
		return err
	}
	op, err := ValidateOperator(f.Op)
	if err != nil {
		return err
	}
	f.Op = op
	return nil
}

// ValidateFilterGroup validates a filter group recursively and normalizes operators in-place.
func (v *Validator) ValidateFilterGroup(g *FilterGroup) error {
	if g == nil {
		return nil
	}
	for i := range g.And {
		if err := v.validateLeaf(&g.And[i]); err != nil {
			return err
		}
	}
	for i := range g.Or {
		if err := v.validateLeaf(&g.Or[i]); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) validateLeaf(leaf *FilterGroupOrLeaf) error {
	if leaf == nil {
		return nil
	}
	if leaf.Filter != nil {
		return v.ValidateFilter(leaf.Filter)
	}
	if leaf.Group != nil {
		return v.ValidateFilterGroup(leaf.Group)
	}
	return nil
}
