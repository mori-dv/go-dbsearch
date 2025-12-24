package go_dbsearch

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SearchHandlerWithOptions performs GET search using query-string parameters.
//
// Options is required (AllowedFields must be set).
func SearchHandlerWithOptions[T any](db *gorm.DB, model T, opts *Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		query, err := ParseQueryWithOptions(c.Request.URL.Query(), opts)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var results []T
		tx := ApplyWithOptions(db.Model(&model), query, opts)

		if err := tx.Find(&results).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, results)
	}
}

// AdvancedSearchRequest is the JSON payload for AdvancedSearchHandlerWithOptions.
type AdvancedSearchRequest struct {
	Filters    *FilterGroup `json:"filters"`
	Sort       []SortOption `json:"sort"`
	Pagination Pagination   `json:"pagination"`
}

// AdvancedSearchHandlerWithOptions performs POST search using JSON body.
//
// Phase-4: Options is required (AllowedFields must be set).
// If opts.FieldTypes is empty, you may call InferFieldTypesFromModel(db, model, opts) once at startup.
func AdvancedSearchHandlerWithOptions[T any](db *gorm.DB, model T, opts *Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AdvancedSearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		v, err := NewValidatorFromOptions(opts)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		caster := NewValueCaster(opts)

		if err := v.ValidateFilterGroup(req.Filters); err != nil {
			if opts.StrictJSON {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			req.Filters = nil
		}

		if err := NormalizeFilterGroupValues(req.Filters, caster); err != nil {
			if opts.StrictJSON {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			req.Filters = nil
		}

		for i := range req.Sort {
			norm, err := v.ValidateSortOption(req.Sort[i])
			if err != nil {
				if opts.StrictJSON {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				req.Sort[i] = SortOption{}
				continue
			}
			req.Sort[i] = norm
		}

		tx := db.Model(&model)

		if req.Filters != nil {
			tx = req.Filters.Apply(tx)
		}

		for _, s := range req.Sort {
			if s.Field == "" {
				continue
			}
			tx = tx.Order(s.Field + " " + s.Direction)
		}

		limit := req.Pagination.Limit
		if opts.MaxLimit > 0 && limit > opts.MaxLimit {
			limit = opts.MaxLimit
		}
		if limit > 0 {
			tx = tx.Limit(limit)
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
