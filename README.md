# go-dbsearch

🔍 **Advanced, secure, and dynamic database search package in Go.**

> Powerful and extensible search functionality for GORM-based REST APIs using Gin framework.

---

## ✨ Features

- ✅ Dynamic filters (`=`, `LIKE`, `IN`, `>`, `<`, `BETWEEN`)
- ✅ Secure field validation (protection against SQL injection)
- ✅ Full pagination support (`limit`, `offset`)
- ✅ Multi-field sorting (`sort=name,-created_at`)
- ✅ Search via query string or JSON body
- ✅ Deep filter grouping (AND/OR nesting)
- ✅ Plug-and-play handler for [Gin](https://github.com/gin-gonic/gin)
- ✅ Works with any GORM model
- ✅ Built-in generic handler with Go generics

---

## 📦 Installation

```bash
go get github.com/mori-dv/go-dbsearch
````

---

## ⚙️ Setup Example

### 1. Define your model

```go
type User struct {
    ID        uint
    Name      string
    Email     string
    Age       int
    CreatedAt time.Time
}
```

---

### 2. Configure Gin + GORM + go-dbsearch

```go
import (
    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "github.com/mori-dv/go-dbsearch"
)

func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    db.AutoMigrate(&User{})

    dbsearch.AllowedFields = map[string]bool{
        "name":       true,
        "email":      true,
        "age":        true,
        "created_at": true,
    }

    r := gin.Default()

    // GET search using query string
    r.GET("/users", dbsearch.SearchHandler[User](db, User{}))

    // POST search using JSON
    r.POST("/users/search", dbsearch.AdvancedSearchHandler[User](db, User{}))

    r.Run()
}
```

---

## 🔍 Search via Query String (GET)

### 🧪 Sample Request

```http
GET /users?filter[name:like]=john&filter[age:>]=18&filter[created_at:between]=2023-01-01,2023-12-31&sort=-age&limit=10&offset=0
```

### 🔗 Supported Filter Operators

| Operator  | Meaning            | Example                                            |
| --------- | ------------------ | -------------------------------------------------- |
| `=`       | Equals             | `filter[age:=]=30`                                 |
| `LIKE`    | Partial match      | `filter[name:like]=john`                           |
| `>` / `<` | Greater/Less than  | `filter[age:>]=25`                                 |
| `IN`      | In list            | `filter[status:in]=active,inactive`                |
| `BETWEEN` | Between two values | `filter[created_at:between]=2023-01-01,2023-12-31` |

### 🔁 Sorting

```http
sort=-age,name
```

* `-` prefix for descending.
* Multi-field supported.

---

## 🧾 Search via JSON (POST)

### 📥 Endpoint

```http
POST /users/search
Content-Type: application/json
```

### 📤 Request Body Example

```json
{
  "filters": {
    "or": [
      {
        "filter": {
          "field": "name",
          "op": "like",
          "value": "john"
        }
      },
      {
        "group": {
          "and": [
            {
              "filter": {
                "field": "email",
                "op": "like",
                "value": "@gmail"
              }
            },
            {
              "filter": {
                "field": "age",
                "op": ">",
                "value": 25
              }
            }
          ]
        }
      }
    ]
  },
  "sort": [
    { "field": "created_at", "direction": "desc" }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0
  }
}
```

### 📥 JSON Schema

```json
{
  "filters": {
    "or": [
      { "filter": { "field": "...", "op": "...", "value": ... } },
      {
        "group": {
          "and": [ ... ],
          "or": [ ... ]
        }
      }
    ]
  },
  "sort": [
    { "field": "fieldname", "direction": "asc|desc" }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0
  }
}
```

---

## 🔐 Security

* ✅ Only whitelisted fields can be queried via `dbsearch.AllowedFields`
* ✅ All values are parameterized (protected from SQL injection)
* ✅ Unsupported fields or operators are ignored

---

## 🧩 Integrations

| Framework | Support                |
| --------- | ---------------------- |
| GORM      | ✅ Fully supported      |
| Gin       | ✅ Plug & play handlers |
| Echo      | ⏳ (Coming soon)        |
| Fiber     | ⏳ (Coming soon)        |

---

## 📌 Roadmap

* [x] Dynamic query string parsing
* [x] Between support
* [x] Nested filters with AND/OR logic
* [x] JSON body support for POST
* [ ] Caching layer for repeated queries
* [ ] Full-text search integration (PostgreSQL, SQLite FTS)
* [ ] Query export for GraphQL compatibility

---

## 🤝 Contributing

Contributions and feature requests are welcome. Fork the repo and submit a pull request 🙌

---

## 📄 License

MIT License — use freely and responsibly.

---

## 🧠 Inspiration

Built for scalable, safe, and expressive filtering in modern backend APIs.

### How It Works

- The core handlers (`SearchHandler`, `AdvancedSearchHandler`) are **generic** (`[T any]`), so you can use them for any GORM model: `User`, `Product`, `Order`, etc.
- You only need to:
  1. Register the allowed fields for each model.
  2. Register the handler for each model’s endpoint.

---

## Example: Use for Multiple Models

Suppose you have these models:

```go
type User struct {
    ID    uint
    Name  string
    Email string
    Age   int
}

type Product struct {
    ID    uint
    Name  string
    Price float64
    Stock int
}
```

### 1. Register Allowed Fields

```go
dbsearch.AllowedFields = map[string]bool{
    // User fields
    "id": true, "name": true, "email": true, "age": true,
    // Product fields
    "price": true, "stock": true,
}
```

### 2. Register Handlers

```go
r.GET("/users", dbsearch.SearchHandler[User](db, User{}))
r.POST("/users/search", dbsearch.AdvancedSearchHandler[User](db, User{}))

r.GET("/products", dbsearch.SearchHandler[Product](db, Product{}))
r.POST("/products/search", dbsearch.AdvancedSearchHandler[Product](db, Product{}))
```

### 3. That’s it!

- You now have full-featured, dynamic search endpoints for every model.
- You can add as many models as you want—just add their fields to `AllowedFields` and register the handlers.

---

## Notes

- **Field Whitelisting:** All searchable/sortable fields for all models must be in `AllowedFields`.
- **Security:** Only fields in `AllowedFields` can be queried, so you’re protected from SQL injection and accidental data leaks.
- **No Code Duplication:** You don’t need to write custom search logic for each model.

---

## Advanced: Per-Model Allowed Fields

If you want different allowed fields per model, you can:
- Use a map of model name → allowed fields, and set `dbsearch.AllowedFields` before each handler runs (using middleware or handler wrapper).
- Or, fork/extend the package to support per-model field whitelists.

---

**Summary:**  
You can use this package for every model in your database, with minimal setup.  
If you want a code template for multiple models, or want to see how to do per-model field whitelisting, just ask!
