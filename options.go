package go_dbsearch

// Options controls how dbsearch parses, validates, casts, and applies search requests.
//
// Phase-4 changes:
//   - Global AllowedFields is removed. Options.AllowedFields is REQUIRED.
//   - FieldTypes can be inferred automatically from a GORM model via InferFieldTypesFromModel.
type Options struct {
	// AllowedFields is the whitelist of safe fields/columns that can be filtered/sorted on.
	// REQUIRED in Phase-4.
	AllowedFields map[string]struct{}

	// FieldTypes provides optional per-field type information used to cast query-string values
	// and normalize JSON values.
	//
	// If not provided, you can infer it via InferFieldTypesFromModel.
	FieldTypes map[string]FieldType

	// StrictJSON controls JSON-body behavior:
	//   - true: invalid filters/sorts/types return HTTP 400 (recommended)
	//   - false: invalid parts may be ignored (not recommended for public APIs)
	StrictJSON bool

	// MaxLimit, if > 0, caps pagination limit for both GET and POST handlers.
	MaxLimit int
}

// NewOptions constructs Options with an allowlist.
// allowedFields must be non-empty for production use.
func NewOptions(allowedFields []string) *Options {
	m := make(map[string]struct{}, len(allowedFields))
	for _, f := range allowedFields {
		if f == "" {
			continue
		}
		m[f] = struct{}{}
	}
	return &Options{
		AllowedFields: m,
		FieldTypes:    map[string]FieldType{},
		StrictJSON:    true,
		MaxLimit:      0,
	}
}

// WithFieldTypes sets FieldTypes and returns opts for chaining.
func (o *Options) WithFieldTypes(types map[string]FieldType) *Options {
	if o == nil {
		return o
	}
	o.FieldTypes = types
	return o
}

// WithMaxLimit sets MaxLimit and returns opts for chaining.
func (o *Options) WithMaxLimit(max int) *Options {
	if o == nil {
		return o
	}
	o.MaxLimit = max
	return o
}

// WithStrictJSON sets StrictJSON and returns opts for chaining.
func (o *Options) WithStrictJSON(strict bool) *Options {
	if o == nil {
		return o
	}
	o.StrictJSON = strict
	return o
}

// SortOption represents a sort term for JSON-body search.
type SortOption struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// Pagination represents limit/offset pagination.
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
