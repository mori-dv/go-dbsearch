package go_dbsearch

import (
	"strings"

	"gorm.io/gorm"
)

type FilterGroup struct {
	And []FilterGroupOrLeaf `json:"and,omitempty"`
	Or  []FilterGroupOrLeaf `json:"or,omitempty"`
}

type FilterGroupOrLeaf struct {
	Group  *FilterGroup `json:"group,omitempty"`
	Filter *Filter      `json:"filter,omitempty"`
}

func (g FilterGroup) Apply(db *gorm.DB) *gorm.DB {
	if len(g.And) > 0 {
		for _, cond := range g.And {
			db = cond.Apply(db)
		}
	}

	if len(g.Or) > 0 {
		var orQueries []string
		var orArgs []interface{}
		for _, cond := range g.Or {
			if cond.Filter != nil {
				switch strings.ToUpper(cond.Filter.Op) {
				case "=":
					orQueries = append(orQueries, cond.Filter.Field+" = ?")
					orArgs = append(orArgs, cond.Filter.Value)
				case "LIKE":
					orQueries = append(orQueries, cond.Filter.Field+" LIKE ?")
					orArgs = append(orArgs, "%"+cond.Filter.Value.(string)+"%")
				case ">":
					orQueries = append(orQueries, cond.Filter.Field+" > ?")
					orArgs = append(orArgs, cond.Filter.Value)
				case "<":
					orQueries = append(orQueries, cond.Filter.Field+" < ?")
					orArgs = append(orArgs, cond.Filter.Value)
				case "IN":
					orQueries = append(orQueries, cond.Filter.Field+" IN (?)")
					orArgs = append(orArgs, cond.Filter.Value)
				case "BETWEEN":
					if valStr, ok := cond.Filter.Value.(string); ok {
						parts := strings.Split(valStr, ",")
						if len(parts) == 2 {
							orQueries = append(orQueries, cond.Filter.Field+" BETWEEN ? AND ?")
							orArgs = append(orArgs, parts[0], parts[1])
						}
					}
				}
			} else if cond.Group != nil {
				sub := cond.Group
				dbSub := db.Session(&gorm.Session{NewDB: true})
				dbSub = sub.Apply(dbSub)
				stmt := dbSub.Statement
				if stmt.SQL.String() != "" {
					orQueries = append(orQueries, "("+stmt.SQL.String()+")")
					orArgs = append(orArgs, stmt.Vars...)
				}
			}
		}
		if len(orQueries) > 0 {
			orClause := strings.Join(orQueries, " OR ")
			db = db.Where(orClause, orArgs...)
		}
	}

	return db
}

func (f FilterGroupOrLeaf) Apply(db *gorm.DB) *gorm.DB {
	if f.Filter != nil {
		return f.Filter.Apply(db)
	} else if f.Group != nil {
		return f.Group.Apply(db)
	}
	return db

}
