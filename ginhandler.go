package go_dbsearch

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SearchHandler[T any](db *gorm.DB, model T) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := ParseQuery(c.Request.URL.Query())
		var results []T
		tx := Apply(db.Model(&model), query)
		if err := tx.Find(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, results)

	}

}

type AdvancedSearchRequest struct {
	Filters    *FilterGroup `json:"filters"`
	Sorts      []SortOption `json:"sort"`
	Pagination Pagination   `json:"pagination"`
}

func AdvancedSearchHandler[T any](db *gorm.DB, model T) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AdvancedSearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx := db.Model(&model)
		if req.Filters != nil {
			tx = req.Filters.Apply(tx)
		}

		for _, sort := range req.Sorts {
			tx = tx.Order(sort.Field + " " + sort.Direction)
		}

		if req.Pagination.Limit > 0 {
			tx = tx.Limit(req.Pagination.Limit)
		}
		if req.Pagination.Offset > 0 {
			tx = tx.Offset(req.Pagination.Offset)
		}

		var results []T
		if err := tx.Find(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, results)
	}
}
