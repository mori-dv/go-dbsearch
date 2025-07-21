package go_dbsearch

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdvancedSearchHandler(t *testing.T) {
	AllowedFields["id"] = true
	AllowedFields["name"] = true
	AllowedFields["email"] = true
	AllowedFields["age"] = true
	db := setupTestDB(t)
	router := gin.Default()
	router.POST("/test", AdvancedSearchHandler[TestModel](db, TestModel{}))

	payload := AdvancedSearchRequest{
		Filters: &FilterGroup{
			Or: []FilterGroupOrLeaf{
				{Filter: &Filter{Field: "email", Op: "like", Value: "@gmail"}},
			},
		},
		Pagination: Pagination{Limit: 5, Offset: 0},
	}

	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", w.Code)
	}
}
