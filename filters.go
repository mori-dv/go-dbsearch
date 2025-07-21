package go_dbsearch

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Filter struct {
	Field string
	Op    string
	Value interface{}
}

func (f Filter) Apply(db *gorm.DB) *gorm.DB {
	switch strings.ToUpper(f.Op) {
	case "=":
		return db.Where(fmt.Sprintf("%s = ?", f.Field), f.Value)
	case "LIKE":
		return db.Where(fmt.Sprintf("%s LIKE ?", f.Field), fmt.Sprintf("%%%v%%", f.Value))
	case ">":
		return db.Where(fmt.Sprintf("%s > ?", f.Field), f.Value)
	case "<":
		return db.Where(fmt.Sprintf("%s < ?", f.Field), f.Value)
	case "IN":
		return db.Where(fmt.Sprintf("%s IN ?", f.Field), f.Value)
	case "BETWEEN":
		if valStr, ok := f.Value.(string); ok {
			parts := strings.Split(valStr, ",")
			if len(parts) == 2 {
				return db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", f.Field), parts[0], parts[1])
			}
		}
	default:
		return db
	}
	return db
}
