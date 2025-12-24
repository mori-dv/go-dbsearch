package go_dbsearch

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type inferModel struct {
	ID        int64
	Name      string
	Age       int
	Active    bool
	Score     float64
	CreatedAt time.Time
}

func TestInferFieldTypesFromModel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	opts := NewOptions([]string{"Name", "Age", "Active", "Score", "CreatedAt"})
	if err := InferFieldTypesFromModel(db, &inferModel{}, opts); err != nil {
		t.Fatalf("infer: %v", err)
	}

	if opts.FieldTypes["Name"] != FieldTypeString {
		t.Fatalf("Name expected string, got %v", opts.FieldTypes["Name"])
	}
	if opts.FieldTypes["Age"] != FieldTypeInt {
		t.Fatalf("Age expected int, got %v", opts.FieldTypes["Age"])
	}
	if opts.FieldTypes["Active"] != FieldTypeBool {
		t.Fatalf("Active expected bool, got %v", opts.FieldTypes["Active"])
	}
	if opts.FieldTypes["Score"] != FieldTypeFloat64 {
		t.Fatalf("Score expected float64, got %v", opts.FieldTypes["Score"])
	}
	if opts.FieldTypes["CreatedAt"] != FieldTypeTime {
		t.Fatalf("CreatedAt expected time, got %v", opts.FieldTypes["CreatedAt"])
	}
}
