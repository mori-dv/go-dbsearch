````md
# go-dbsearch

Secure, dynamic search layer for GORM-based REST APIs (Gin-friendly), with:

- Query-string search (GET)
- JSON-body advanced search (POST)
- Safe allowlist validation (field whitelist + identifier regex)
- Nested AND/OR groups
- Casting for `IN` / `BETWEEN` + basic scalar types
- Optional automatic `FieldType` inference via GORM schema

> Module: `github.com/mori-dv/go-dbsearch`  
> Package: `go_dbsearch`

---

## Table of contents

- [Installation](#installation)
- [Quick start (production setup)](#quick-start-production-setup)
- [Concepts](#concepts)
  - [Allowlist](#allowlist)
  - [Operators](#operators)
  - [Casting](#casting)
  - [Nested groups](#nested-groups)
- [GET: Query-string search](#get-query-string-search)
  - [Filters](#filters)
  - [Sorting](#sorting)
  - [Pagination](#pagination)
  - [Examples (GET)](#examples-get)
- [POST: JSON advanced search](#post-json-advanced-search)
  - [Request body schema](#request-body-schema)
  - [Examples (POST)](#examples-post)
- [Type inference from GORM model](#type-inference-from-gorm-model)
- [Security](#security)
- [Performance notes](#performance-notes)
- [Compatibility](#compatibility)
- [Recommended production checklist](#recommended-production-checklist)
- [License](#license)

---

## Installation

```bash
go get github.com/mori-dv/go-dbsearch
````

---

## Quick start (production setup)

This is the recommended “production-safe” setup:

* Use per-handler `Options` with a **non-empty allowlist**.
* Keep `StrictJSON=true` (recommended).
* Cap pagination with `MaxLimit`.
* Optionally infer field types once at startup.

```go
package main

import (
  "github.com/gin-gonic/gin"
  "gorm.io/gorm"

  go_dbsearch "github.com/mori-dv/go-dbsearch"
)

type User struct {
  ID        uint
  Name      string
  Email     string
  Age       int
  Status    string
  CreatedAt time.Time
}

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
  // 1) Allow only fields you are willing to expose for search/sort
  opts := go_dbsearch.NewOptions([]string{
    "name", "email", "age", "status", "created_at",
  }).
    WithStrictJSON(true).
    WithMaxLimit(100)

  // 2) Optional: infer types once at startup (recommended)
  _ = go_dbsearch.InferFieldTypesFromModel(db, &User{}, opts)

  // 3) Register handlers
  r.GET("/users", go_dbsearch.SearchHandlerWithOptions[User](db, User{}, opts))
  r.POST("/users/search", go_dbsearch.AdvancedSearchHandlerWithOptions[User](db, User{}, opts))
}
```

---

## Concepts

### Allowlist

To prevent SQL injection through identifiers, this library requires a **field allowlist**:

* `Options.AllowedFields` must contain the only fields that can be used in `filter.field` and `sort.field`.

Examples:

* ✅ Allowed: `"name"`, `"users.email"`, `"created_at"`
* ❌ Not allowed: `"name; DROP TABLE users"`, `"CASE WHEN ..."`, `"LOWER(email)"`

In addition to allowlist, the library also rejects unsafe identifier characters using:

* `^[a-zA-Z0-9_.]+$`

---

### Operators

Canonical operators:

* `=`, `>`, `<`, `>=`, `<=`
* `LIKE`
* `IN`
* `BETWEEN`

Aliases (case-insensitive):

* `eq`, `gt`, `lt`, `gte`, `lte`
* `like`, `in`, `between`

So these are equivalent:

* `op: "eq"` and `op: "="`
* `op: "like"` and `op: "LIKE"`

---

### Casting

Casting is driven by `Options.FieldTypes`:

```go
opts.WithFieldTypes(map[string]go_dbsearch.FieldType{
  "age":        go_dbsearch.FieldTypeInt,
  "created_at": go_dbsearch.FieldTypeDate, // "2006-01-02"
})
```

If `FieldTypes` is empty, you can infer it from the GORM model schema:

```go
_ = go_dbsearch.InferFieldTypesFromModel(db, &User{}, opts)
```

Supported types:

* `string`, `int`, `int64`, `float64`, `bool`, `date`, `time`

Formats:

* `date`: `"2006-01-02"` (UTC midnight)
* `time`: RFC3339 (`"2023-01-02T15:04:05Z"`) or `"2006-01-02 15:04:05"`

`IN` / `BETWEEN` behavior:

* GET `IN`: `a,b,c` → `[]interface{}{...}`
* GET `BETWEEN`: `a,b` → `[]interface{}{lo,hi}`
* JSON `IN`: `"a,b"` or array `["a","b"]`
* JSON `BETWEEN`: `"a,b"` or array `[lo,hi]`

---

### Nested groups

JSON advanced search supports nested AND/OR groups at arbitrary depth.

Example structure:

* `(age >= 18) AND (status IN (...) OR email LIKE '%gmail%')`

---

## GET: Query-string search

### Filters

Format:

```
filter[field:op]=value
```

Examples:

* Equals:
  `filter[name:eq]=Alice`
* Greater-than:
  `filter[age:gt]=18`
* LIKE:
  `filter[email:like]=@gmail.com`
* IN:
  `filter[status:in]=active,inactive,pending`
* BETWEEN:
  `filter[created_at:between]=2024-01-01,2024-12-31`

Notes:

* Invalid fields (not in allowlist) are ignored in GET mode (permissive parsing).
* Invalid casts cause the specific filter to be ignored.

---

### Sorting

Format:

```
sort=name,-created_at
```

* `name` → `ASC`
* `-created_at` → `DESC`

Only allowlisted fields can be used for sorting.

---

### Pagination

Format:

```
limit=10&offset=0
```

If `Options.MaxLimit > 0`, the requested `limit` is capped to that value.

---

### Examples (GET)

#### 1) Basic filters + sort + pagination

```
GET /users?filter[name:like]=ali&filter[age:gte]=18&sort=-created_at&limit=20&offset=0
```

#### 2) IN filter

```
GET /users?filter[status:in]=active,pending
```

#### 3) BETWEEN filter (date)

```
GET /users?filter[created_at:between]=2024-01-01,2024-12-31
```

---

## POST: JSON advanced search

### Request body schema

```json
{
  "filters": {
    "and": [
      { "filter": { "field": "age", "op": ">=", "value": 18 } },
      {
        "or": [
          { "filter": { "field": "status", "op": "in", "value": ["active","pending"] } },
          { "filter": { "field": "email", "op": "like", "value": "@gmail.com" } }
        ]
      }
    ]
  },
  "sort": [
    { "field": "created_at", "direction": "desc" }
  ],
  "pagination": { "limit": 20, "offset": 0 }
}
```

Validation behavior:

* With `StrictJSON=true` (recommended), invalid input returns **HTTP 400**:

  * Unknown field (not allowlisted)
  * Unsupported operator
  * Invalid `sort.direction`
  * Type casting failure (if FieldTypes is configured)

---

### Examples (POST)

#### 1) Nested groups

```json
{
  "filters": {
    "and": [
      { "filter": { "field": "age", "op": ">=", "value": 18 } },
      {
        "or": [
          { "filter": { "field": "status", "op": "in", "value": "active,pending" } },
          { "filter": { "field": "email", "op": "like", "value": "@gmail.com" } }
        ]
      }
    ]
  },
  "sort": [{ "field": "created_at", "direction": "desc" }],
  "pagination": { "limit": 50, "offset": 0 }
}
```

#### 2) BETWEEN with array

```json
{
  "filters": {
    "and": [
      { "filter": { "field": "created_at", "op": "between", "value": ["2024-01-01","2024-12-31"] } }
    ]
  }
}
```

---

## Type inference from GORM model

To avoid manually maintaining `FieldTypes`, you can infer them from a GORM model:

```go
opts := go_dbsearch.NewOptions([]string{"name","age","created_at"}).
  WithStrictJSON(true).
  WithMaxLimit(100)

if err := go_dbsearch.InferFieldTypesFromModel(db, &User{}, opts); err != nil {
  // handle error
}
```

Notes:

* Inference is best-effort.
* `time.Time` fields map to `FieldTypeTime`.
* If you need date-only behavior, override manually:
  `opts.FieldTypes["created_at"] = go_dbsearch.FieldTypeDate`

---

## Security

This library prevents SQL injection by:

1. **Field allowlist**: only allowlisted identifiers can be used.
2. **Identifier regex check**: only `[a-zA-Z0-9_.]` characters are allowed in identifiers.
3. **Parameterized values**: values are always passed as `?` parameters to GORM.

Production recommendations:

* Keep allowlist small and explicit.
* Set `StrictJSON=true`.
* Cap `limit` using `MaxLimit`.
* Add DB indexes for the most-used filter/sort fields.

Non-goals:

* Function-based filters like `LOWER(email)` are intentionally not supported (unsafe).
* Authorization rules are not handled; your allowlist must match your auth policy.

---

## Performance notes

Dynamic filtering can be expensive without indexes.
For production:

* Add indexes for commonly filtered/sorted columns.
* Set a reasonable `MaxLimit`.
* Consider restricting operators exposed to public endpoints.

---

## Compatibility

* Go: 1.22+
* GORM: v2
* Gin: supported via provided handlers (the core logic can be used without Gin).

---

## Recommended production checklist

* [ ] Provide a non-empty allowlist (`Options.AllowedFields`)
* [ ] `StrictJSON=true`
* [ ] Set `MaxLimit` (e.g. 100)
* [ ] Add DB indexes for filter/sort columns
* [ ] Run: `go test ./...`
* [ ] Run: `go vet ./...`
* [ ] Run: `golangci-lint run`
* [ ] Add CI workflow (GitHub Actions)
* [ ] Tag releases with SemVer (e.g. `v2.0.0`)

---

## License

See `LICENSE`.

```
::contentReference[oaicite:0]{index=0}
```
