package go_dbsearch

var AllowedFields = map[string]bool{
	"id":    true,
	"name":  true,
	"email": true,
	"age":   true,
}

func IsFieldAllowed(field string) bool {
	return AllowedFields[field]
}
