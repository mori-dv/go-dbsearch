package go_dbsearch

import "gorm.io/gorm"

// FilterGroup represents a nested boolean expression for filters.
type FilterGroup struct {
	And []FilterGroupOrLeaf `json:"and,omitempty"`
	Or  []FilterGroupOrLeaf `json:"or,omitempty"`
}

// FilterGroupOrLeaf is a union type used inside FilterGroup.
type FilterGroupOrLeaf struct {
	Group  *FilterGroup `json:"group,omitempty"`
	Filter *Filter      `json:"filter,omitempty"`
}

// Apply applies the filter group to a GORM query.
//
// Implementation notes:
//   - Uses nested scopes (Where(subDB)/Or(subDB)) so deep nesting is reliable.
//   - Does not rely on extracting raw SQL from GORM statements.
//
// Security:
//   - Validate fields/operators with Validator before calling Apply (recommended).
func (g *FilterGroup) Apply(db *gorm.DB) *gorm.DB {
	if g == nil {
		return db
	}

	for _, item := range g.And {
		db = applyLeafAsAnd(db, item)
	}

	if len(g.Or) > 0 {
		orBlock := newScopeDB(db)
		first := true

		for _, item := range g.Or {
			branch := newScopeDB(db)
			branch = item.applyToScope(branch)

			if first {
				orBlock = orBlock.Where(branch)
				first = false
			} else {
				orBlock = orBlock.Or(branch)
			}
		}

		if !first {
			db = db.Where(orBlock)
		}
	}

	return db
}

func applyLeafAsAnd(db *gorm.DB, item FilterGroupOrLeaf) *gorm.DB {
	if item.Filter != nil {
		return item.Filter.Apply(db)
	}
	if item.Group != nil {
		sub := newScopeDB(db)
		sub = item.Group.Apply(sub)
		return db.Where(sub)
	}
	return db
}

func (l FilterGroupOrLeaf) applyToScope(scope *gorm.DB) *gorm.DB {
	if l.Filter != nil {
		return l.Filter.Apply(scope)
	}
	if l.Group != nil {
		return l.Group.Apply(scope)
	}
	return scope
}

func newScopeDB(parent *gorm.DB) *gorm.DB {
	scope := parent.Session(&gorm.Session{NewDB: true})

	if parent != nil && parent.Statement != nil {
		if parent.Statement.Table != "" {
			scope = scope.Table(parent.Statement.Table)
		} else if parent.Statement.Model != nil {
			scope = scope.Model(parent.Statement.Model)
		}
	}

	return scope
}
