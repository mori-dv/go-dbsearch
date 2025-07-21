package go_dbsearch

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GroupTestModel struct {
	ID    uint
	Name  string
	Age   int
	Email string
}

func setupGroupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	if err := db.AutoMigrate(&GroupTestModel{}); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
	db.Create(&GroupTestModel{Name: "Alice", Age: 30, Email: "alice@test.com"})
	db.Create(&GroupTestModel{Name: "Bob", Age: 25, Email: "bob@gmail.com"})
	return db
}

func TestFilterGroup_And(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupGroupTestDB(t)
	group := FilterGroup{
		And: []FilterGroupOrLeaf{
			{Filter: &Filter{Field: "name", Op: "=", Value: "Alice"}},
			{Filter: &Filter{Field: "age", Op: ">", Value: 20}},
		},
	}
	var result []GroupTestModel
	db = group.Apply(db.Model(&GroupTestModel{})).Find(&result).Statement.DB
	if len(result) != 1 || result[0].Name != "Alice" {
		t.Fatalf("AND group failed, got: %+v", result)
	}
}

func TestFilterGroup_Or(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupGroupTestDB(t)
	group := FilterGroup{
		Or: []FilterGroupOrLeaf{
			{Filter: &Filter{Field: "name", Op: "=", Value: "Alice"}},
			{Filter: &Filter{Field: "name", Op: "=", Value: "Bob"}},
		},
	}
	var result []GroupTestModel
	db = group.Apply(db.Model(&GroupTestModel{})).Find(&result).Statement.DB
	if len(result) != 2 {
		t.Fatalf("OR group failed, got: %+v", result)
	}
}

func TestFilterGroup_Nested(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupGroupTestDB(t)
	group := FilterGroup{
		And: []FilterGroupOrLeaf{
			{Group: &FilterGroup{
				Or: []FilterGroupOrLeaf{
					{Filter: &Filter{Field: "name", Op: "=", Value: "Alice"}},
					{Filter: &Filter{Field: "name", Op: "=", Value: "Bob"}},
				},
			}},
			{Filter: &Filter{Field: "age", Op: ">", Value: 20}},
		},
	}
	var result []GroupTestModel
	db = group.Apply(db.Model(&GroupTestModel{})).Find(&result).Statement.DB
	if len(result) != 2 {
		t.Fatalf("Nested group failed, got: %+v", result)
	}
}

func TestFilterGroup_Empty(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupGroupTestDB(t)
	group := FilterGroup{}
	var result []GroupTestModel
	db = group.Apply(db.Model(&GroupTestModel{})).Find(&result).Statement.DB
	if len(result) != 2 {
		t.Fatalf("Empty group should return all, got: %+v", result)
	}
}
