package go_dbsearch

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

// InferFieldTypesFromModel infers FieldTypes (used for casting) from the provided GORM model.
//
// It inspects the GORM schema parsed from model and maps common Go kinds to FieldType.
// Only fields present in opts.AllowedFields are inferred; others are ignored.
//
// Notes:
//   - This function is best-effort. If a field cannot be resolved, it is not added.
//   - For timestamps, this looks for time.Time type.
//   - For dates-only vs timestamps, it defaults to FieldTypeTime (you can override manually).
func InferFieldTypesFromModel(db *gorm.DB, model any, opts *Options) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if opts == nil || opts.AllowedFields == nil || len(opts.AllowedFields) == 0 {
		return fmt.Errorf("opts.AllowedFields is required to infer FieldTypes")
	}

	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return fmt.Errorf("failed to parse gorm model: %w", err)
	}
	if stmt.Schema == nil {
		return fmt.Errorf("gorm schema is nil after parse")
	}

	if opts.FieldTypes == nil {
		opts.FieldTypes = map[string]FieldType{}
	}

	// GORM schema fields contain both Name and DBName.
	// We match AllowedFields keys to DBName (recommended) and also allow match to Name.
	for _, f := range stmt.Schema.Fields {
		dbName := f.DBName
		goName := f.Name

		// Which key is whitelisted?
		_, okDB := opts.AllowedFields[dbName]
		_, okGo := opts.AllowedFields[goName]
		if !okDB && !okGo {
			continue
		}

		ft, ok := inferFieldTypeFromReflect(f.FieldType)
		if !ok {
			continue
		}

		// Prefer DBName as the canonical key if it's in allowlist; else use goName.
		if okDB {
			opts.FieldTypes[dbName] = ft
		} else {
			opts.FieldTypes[goName] = ft
		}
	}

	return nil
}

func inferFieldTypeFromReflect(t reflect.Type) (FieldType, bool) {
	if t == nil {
		return "", false
	}

	// Deref pointers
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
		if t == nil {
			return "", false
		}
	}

	// time.Time detection by package + name
	if t.PkgPath() == "time" && t.Name() == "Time" {
		return FieldTypeTime, true
	}

	switch t.Kind() {
	case reflect.String:
		return FieldTypeString, true
	case reflect.Bool:
		return FieldTypeBool, true
	case reflect.Int:
		return FieldTypeInt, true
	case reflect.Int64:
		return FieldTypeInt64, true
	case reflect.Float32, reflect.Float64:
		return FieldTypeFloat64, true
	default:
		return "", false
	}
}
