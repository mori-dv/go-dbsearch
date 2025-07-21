package go_dbsearch

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestModel struct {
	ID    uint
	Name  string
	Age   int
	Email string
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	db.Create(&TestModel{Name: "Alice", Age: 30, Email: "alice@test.com"})
	db.Create(&TestModel{Name: "Bob", Age: 25, Email: "bob@gmail.com"})
	return db
}

func TestFilter_Apply_Equals(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupTestDB(t)
	f := Filter{
		Field: "name",
		Op:    "=",
		Value: "Alice",
	}
	var result []TestModel
	tx := f.Apply(db.Model(&TestModel{})).Find(&result)
	if tx.Error != nil || len(result) != 1 || result[0].Name != "Alice" {
		t.Fatalf("Filter failed, got: %+v", result)
	}
}

func TestFilter_Apply_Between(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupTestDB(t)
	f := Filter{
		Field: "age",
		Op:    "between",
		Value: "20,28",
	}
	var result []TestModel
	tx := f.Apply(db.Model(&TestModel{})).Find(&result)
	if tx.Error != nil || len(result) != 1 || result[0].Name != "Bob" {
		t.Fatalf("Between filter failed, got: %+v", result)
	}
}

func TestFilter_Apply_Like(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupTestDB(t)
	f := Filter{
		Field: "email",
		Op:    "like",
		Value: "gmail",
	}
	var result []TestModel
	tx := f.Apply(db.Model(&TestModel{})).Find(&result)
	if tx.Error != nil || len(result) != 1 || result[0].Email != "bob@gmail.com" {
		t.Fatalf("LIKE filter failed, got: %+v", result)
	}
}

func TestFilter_Apply_GreaterThan(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupTestDB(t)
	f := Filter{
		Field: "age",
		Op:    ">",
		Value: 26,
	}
	var result []TestModel
	tx := f.Apply(db.Model(&TestModel{})).Find(&result)
	if tx.Error != nil || len(result) != 1 || result[0].Name != "Alice" {
		t.Fatalf("> filter failed, got: %+v", result)
	}
}

func TestFilter_Apply_LessThan(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupTestDB(t)
	f := Filter{
		Field: "age",
		Op:    "<",
		Value: 28,
	}
	var result []TestModel
	tx := f.Apply(db.Model(&TestModel{})).Find(&result)
	if tx.Error != nil || len(result) != 1 || result[0].Name != "Bob" {
		t.Fatalf("< filter failed, got: %+v", result)
	}
}

func TestFilter_Apply_In(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupTestDB(t)
	f := Filter{
		Field: "name",
		Op:    "in",
		Value: []string{"Alice", "Bob"},
	}
	var result []TestModel
	tx := f.Apply(db.Model(&TestModel{})).Find(&result)
	if tx.Error != nil || len(result) != 2 {
		t.Fatalf("IN filter failed, got: %+v", result)
	}
}

func TestFilter_Apply_InvalidOperator(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupTestDB(t)
	f := Filter{
		Field: "name",
		Op:    "invalid",
		Value: "Alice",
	}
	var result []TestModel
	tx := f.Apply(db.Model(&TestModel{})).Find(&result)
	if tx.Error != nil {
		t.Fatalf("Invalid operator should not error, got: %+v", tx.Error)
	}
}

func TestFilter_Apply_BetweenMalformed(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupTestDB(t)
	f := Filter{
		Field: "age",
		Op:    "between",
		Value: "20",
	}
	var result []TestModel
	tx := f.Apply(db.Model(&TestModel{})).Find(&result)
	if tx.Error != nil {
		t.Fatalf("Malformed BETWEEN should not error, got: %+v", tx.Error)
	}
}
