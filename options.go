package go_dbsearch

type SortOption struct {
	Field     string
	Direction string
}

type Pagination struct {
	Limit  int
	Offset int
}
