package go_dbsearch

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FieldType describes the expected type of a searchable field.
// It is used to cast query-string values and to normalize JSON values.
type FieldType string

const (
	// FieldTypeString treats values as strings.
	FieldTypeString FieldType = "string"
	// FieldTypeInt casts values to int.
	FieldTypeInt FieldType = "int"
	// FieldTypeInt64 casts values to int64.
	FieldTypeInt64 FieldType = "int64"
	// FieldTypeFloat64 casts values to float64.
	FieldTypeFloat64 FieldType = "float64"
	// FieldTypeBool casts values to bool.
	FieldTypeBool FieldType = "bool"
	// FieldTypeDate parses dates in "2006-01-02" format (UTC, midnight).
	FieldTypeDate FieldType = "date"
	// FieldTypeTime parses timestamps in RFC3339 (e.g. "2023-01-02T15:04:05Z") or "2006-01-02 15:04:05".
	FieldTypeTime FieldType = "time"
)

// ValueCaster casts and normalizes values based on Options.FieldTypes.
type ValueCaster struct {
	fieldTypes map[string]FieldType
}

// NewValueCaster creates a caster from options. If opts is nil, it defaults to string casting.
func NewValueCaster(opts *Options) *ValueCaster {
	ft := map[string]FieldType{}
	if opts != nil && opts.FieldTypes != nil {
		ft = opts.FieldTypes
	}
	return &ValueCaster{fieldTypes: ft}
}

// CastFromString casts a raw query-string value for a given field into the configured type.
// If no type is configured for the field, the value is returned as-is (string).
func (c *ValueCaster) CastFromString(field string, raw string) (interface{}, error) {
	t, ok := c.fieldTypes[field]
	if !ok || t == "" || t == FieldTypeString {
		return raw, nil
	}

	switch t {
	case FieldTypeInt:
		v, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("invalid int for %s: %q", field, raw)
		}
		return v, nil
	case FieldTypeInt64:
		v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int64 for %s: %q", field, raw)
		}
		return v, nil
	case FieldTypeFloat64:
		v, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float64 for %s: %q", field, raw)
		}
		return v, nil
	case FieldTypeBool:
		v, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("invalid bool for %s: %q", field, raw)
		}
		return v, nil
	case FieldTypeDate:
		tt, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("invalid date for %s: %q", field, raw)
		}
		return time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, time.UTC), nil
	case FieldTypeTime:
		s := strings.TrimSpace(raw)
		if tt, err := time.Parse(time.RFC3339, s); err == nil {
			return tt, nil
		}
		if tt, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
			return tt, nil
		}
		return nil, fmt.Errorf("invalid time for %s: %q", field, raw)
	default:
		return raw, nil
	}
}

// NormalizeJSONValue normalizes JSON values for a given field according to its configured FieldType.
// It accepts common JSON types (string/number/bool) and converts them to the expected Go type.
// If no type is configured, the value is returned unchanged.
func (c *ValueCaster) NormalizeJSONValue(field string, v interface{}) (interface{}, error) {
	t, ok := c.fieldTypes[field]
	if !ok || t == "" || t == FieldTypeString {
		return v, nil
	}

	switch t {
	case FieldTypeInt:
		return normalizeToInt(field, v)
	case FieldTypeInt64:
		return normalizeToInt64(field, v)
	case FieldTypeFloat64:
		return normalizeToFloat64(field, v)
	case FieldTypeBool:
		return normalizeToBool(field, v)
	case FieldTypeDate:
		switch vv := v.(type) {
		case string:
			return c.CastFromString(field, vv)
		case time.Time:
			return time.Date(vv.Year(), vv.Month(), vv.Day(), 0, 0, 0, 0, time.UTC), nil
		default:
			return nil, fmt.Errorf("invalid date value for %s: %T", field, v)
		}
	case FieldTypeTime:
		switch vv := v.(type) {
		case string:
			return c.CastFromString(field, vv)
		case time.Time:
			return vv, nil
		default:
			return nil, fmt.Errorf("invalid time value for %s: %T", field, v)
		}
	default:
		return v, nil
	}
}

func normalizeToInt(field string, v interface{}) (interface{}, error) {
	switch vv := v.(type) {
	case float64:
		return int(vv), nil
	case int:
		return vv, nil
	case int64:
		return int(vv), nil
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(vv))
		if err != nil {
			return nil, fmt.Errorf("invalid int for %s: %q", field, vv)
		}
		return i, nil
	default:
		return nil, fmt.Errorf("invalid int value for %s: %T", field, v)
	}
}

func normalizeToInt64(field string, v interface{}) (interface{}, error) {
	switch vv := v.(type) {
	case float64:
		return int64(vv), nil
	case int:
		return int64(vv), nil
	case int64:
		return vv, nil
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(vv), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid int64 for %s: %q", field, vv)
		}
		return i, nil
	default:
		return nil, fmt.Errorf("invalid int64 value for %s: %T", field, v)
	}
}

func normalizeToFloat64(field string, v interface{}) (interface{}, error) {
	switch vv := v.(type) {
	case float64:
		return vv, nil
	case int:
		return float64(vv), nil
	case int64:
		return float64(vv), nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(vv), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float64 for %s: %q", field, vv)
		}
		return f, nil
	default:
		return nil, fmt.Errorf("invalid float64 value for %s: %T", field, v)
	}
}

func normalizeToBool(field string, v interface{}) (interface{}, error) {
	switch vv := v.(type) {
	case bool:
		return vv, nil
	case string:
		b, err := strconv.ParseBool(strings.TrimSpace(vv))
		if err != nil {
			return nil, fmt.Errorf("invalid bool for %s: %q", field, vv)
		}
		return b, nil
	default:
		return nil, fmt.Errorf("invalid bool value for %s: %T", field, v)
	}
}
