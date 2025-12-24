package go_dbsearch

import "fmt"

// NormalizeFilterGroupValues casts/normalizes filter values in a FilterGroup in-place.
//
// Operator-specific behavior:
//   - IN:       value may be "a,b" or an array; normalized into []interface{}.
//   - BETWEEN:  value may be "a,b" or an array length 2; normalized into []interface{}{lo, hi}.
//   - LIKE:     value is converted to string.
//   - Others:   value is normalized to the configured type for the field.
func NormalizeFilterGroupValues(g *FilterGroup, caster *ValueCaster) error {
	if g == nil {
		return nil
	}
	for i := range g.And {
		if err := normalizeLeaf(&g.And[i], caster); err != nil {
			return err
		}
	}
	for i := range g.Or {
		if err := normalizeLeaf(&g.Or[i], caster); err != nil {
			return err
		}
	}
	return nil
}

func normalizeLeaf(l *FilterGroupOrLeaf, caster *ValueCaster) error {
	if l == nil {
		return nil
	}
	if l.Filter != nil {
		return normalizeFilterValue(l.Filter, caster)
	}
	if l.Group != nil {
		return NormalizeFilterGroupValues(l.Group, caster)
	}
	return nil
}

func normalizeFilterValue(f *Filter, caster *ValueCaster) error {
	if f == nil {
		return nil
	}
	op, _ := NormalizeOperator(f.Op)

	switch op {
	case "LIKE":
		f.Value = fmt.Sprintf("%v", f.Value)
		return nil
	case "IN":
		list, err := normalizeJSONList(f.Field, f.Value, caster)
		if err != nil {
			return err
		}
		f.Value = list
		return nil
	case "BETWEEN":
		pair, err := normalizeJSONBetweenPair(f.Field, f.Value, caster)
		if err != nil {
			return err
		}
		f.Value = pair
		return nil
	default:
		nv, err := caster.NormalizeJSONValue(f.Field, f.Value)
		if err != nil {
			return err
		}
		f.Value = nv
		return nil
	}
}

func normalizeJSONList(field string, v interface{}, caster *ValueCaster) ([]interface{}, error) {
	switch vv := v.(type) {
	case string:
		parts := splitCSV(vv)
		out := make([]interface{}, 0, len(parts))
		for _, p := range parts {
			cv, err := caster.CastFromString(field, p)
			if err != nil {
				return nil, err
			}
			out = append(out, cv)
		}
		return out, nil
	case []interface{}:
		out := make([]interface{}, 0, len(vv))
		for _, item := range vv {
			nv, err := caster.NormalizeJSONValue(field, item)
			if err != nil {
				return nil, err
			}
			out = append(out, nv)
		}
		return out, nil
	case []string:
		out := make([]interface{}, 0, len(vv))
		for _, item := range vv {
			cv, err := caster.CastFromString(field, item)
			if err != nil {
				return nil, err
			}
			out = append(out, cv)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("invalid IN value type for %s: %T", field, v)
	}
}

func normalizeJSONBetweenPair(field string, v interface{}, caster *ValueCaster) ([]interface{}, error) {
	switch vv := v.(type) {
	case string:
		parts := splitCSV(vv)
		if len(parts) != 2 {
			return nil, fmt.Errorf("BETWEEN value must have 2 items for %s", field)
		}
		lo, err := caster.CastFromString(field, parts[0])
		if err != nil {
			return nil, err
		}
		hi, err := caster.CastFromString(field, parts[1])
		if err != nil {
			return nil, err
		}
		return []interface{}{lo, hi}, nil
	case []interface{}:
		if len(vv) != 2 {
			return nil, fmt.Errorf("BETWEEN value must have 2 items for %s", field)
		}
		lo, err := caster.NormalizeJSONValue(field, vv[0])
		if err != nil {
			return nil, err
		}
		hi, err := caster.NormalizeJSONValue(field, vv[1])
		if err != nil {
			return nil, err
		}
		return []interface{}{lo, hi}, nil
	case []string:
		if len(vv) != 2 {
			return nil, fmt.Errorf("BETWEEN value must have 2 items for %s", field)
		}
		lo, err := caster.CastFromString(field, vv[0])
		if err != nil {
			return nil, err
		}
		hi, err := caster.CastFromString(field, vv[1])
		if err != nil {
			return nil, err
		}
		return []interface{}{lo, hi}, nil
	default:
		return nil, fmt.Errorf("invalid BETWEEN value type for %s: %T", field, v)
	}
}
