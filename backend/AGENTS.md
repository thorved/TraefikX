# Backend (Go)

Go 1.25 API using Gin framework, GORM ORM, and SQLite.

## Structure

```
cmd/api/main.go           # Entry point
internal/
├── auth/                 # JWT & OIDC
├── config/               # Environment config
├── database/             # DB init, migrations, password hash
├── handlers/             # HTTP handlers
├── middleware/           # Auth middleware
├── models/               # GORM models + request/response structs
└── routes/               # Route setup
```

## Patterns

### Handlers
- Struct-based with DB dependency injection
- Return JSON with `c.JSON(status, gin.H{"error": "msg"})` or structs
- Use `c.ShouldBindJSON(&req)` for request parsing
- Validate with struct tags: `binding:"required,email"`

### Models
- GORM tags: `gorm:"primaryKey"`, `gorm:"uniqueIndex;not null"`
- JSON tags: `json:"field_name"`
- Use `json:"-"` to exclude sensitive fields (password)
- Response structs with `ToResponse()` method
- Time fields use `*time.Time` for nullable

### Error Responses
```go
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
```

## Commands

```bash
go run cmd/api/main.go          # Dev server
go test ./...                   # Run tests
go fmt ./...                    # Format
go vet ./...                    # Static analysis
go build -o api cmd/api/main.go # Build binary
```

## Environment Variables

Copy `.env.example` to `.env`:
- `PORT=8080`
- `JWT_SECRET` (min 32 chars)
- `DATABASE_PATH=./data/traefikx.db`
- `OIDC_ENABLED`, `OIDC_*` for OIDC config

## Dependencies

Key packages:
- `github.com/gin-gonic/gin` - HTTP framework
- `gorm.io/gorm` + `github.com/glebarez/sqlite` - ORM
- `github.com/golang-jwt/jwt/v5` - JWT
- `golang.org/x/crypto/bcrypt` - Password hashing
- `golang.org/x/oauth2` - OIDC
