package go_dbsearch

var AllowedFields = map[string]bool{}

func IsFieldAllowed(field string) bool {
	return AllowedFields[field]
}
