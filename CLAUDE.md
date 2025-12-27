# Expense Tracker Go API

A Go REST API backend for the Expense Tracker v2 application, replacing a Firebase-based system. Provides expense management with categories, subcategories, user management, and analytics.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22.2 |
| HTTP | Standard library `net/http` (no framework) |
| Database | MySQL 8.0 via `go-sql-driver/mysql` |
| Auth | JWT (`golang-jwt/jwt/v5`) + bcrypt (`golang.org/x/crypto`) |
| Config | `godotenv` for .env loading |
| Container | Multi-stage Docker build with distroless base |

## Project Structure

```
expense-tracker-api/
├── main.go          # HTTP server, routing, auth middleware
├── handlers.go      # All endpoint handlers (categories, subcategories, expenses, users)
├── database.go      # DB init, user queries, JWT operations, password verification
├── migrations.go    # Auto-migration runner (embeds SQL files)
├── currency.go      # EUR/BGN conversion helpers
├── logger.go        # Custom logging utility
├── go.mod           # Dependencies
├── Dockerfile       # Multi-stage build (PORT 8082)
├── migrations/      # SQL migration files (embedded into binary)
├── api_endpoints.md # Endpoint documentation
└── llm/             # LLM-specific docs (auth, schema, expenses)
```

## Database Schema

**users** - id, uid, email (unique), display_name, created_at, password (bcrypt), role (MEMBER/ADMIN), default_currency (EUR/BGN, default BGN)

**categories** - id, name (unique), created_at

**subcategories** - id, category_id (FK), name, created_at | Unique on (category_id, name)

**expenses** - id, amount (decimal), subcategory_id (FK), user_id (FK), note, created_at, is_euro (bool, default false)

## API Endpoints

### Public
- `POST /api/v1/login` - Authenticate user, returns JWT
- `GET /api/v1/health` - Health check

### Protected (require `Authorization: Bearer <token>`)

**Categories**
- `GET /api/v1/categories` - List all with nested subcategories
- `POST /api/v1/categories` - Create (name min 3 chars)
- `GET /api/v1/categories/{id}` - Get single with subcategories
- `PUT /api/v1/categories/{id}` - Update name
- `DELETE /api/v1/categories/{id}` - Delete (blocked if has subcategories)

**Subcategories**
- `GET /api/v1/subcategories` - List all with category info
- `POST /api/v1/subcategories` - Create (requires category_id)
- `GET /api/v1/subcategories/{id}` - Get single with category
- `PUT /api/v1/subcategories/{id}` - Update name
- `DELETE /api/v1/subcategories/{id}` - Delete (blocked if has expenses)

**Expenses**
- `GET /api/v1/expenses` - List with filtering/grouping (see below)
- `POST /api/v1/expenses` - Create (amount required)
- `DELETE /api/v1/expenses/{id}` - Delete

**Analytics**
- `GET /api/v1/grouped-expenses-by-subcategory` - Monthly breakdown (requires `month` param)
- `GET /api/v1/subcategories-by-expense-count` - Debug: expense counts per subcategory
- `GET /api/v1/member-users` - List users with role='MEMBER'

**User Preferences**
- `GET /api/v1/user/preferences` - Get current user's preferences (default_currency)
- `PUT /api/v1/user/preferences` - Update preferences (`{"default_currency": "EUR"}` or `"BGN"`)

### Expense Query Parameters

| Parameter | Description |
|-----------|-------------|
| `user_id` | Filter by user (0 for NULL users) |
| `category_id` | Single or comma-separated list |
| `subcategory_id` | Single or comma-separated list |
| `date_from`, `date_to` | Date range (YYYY-MM-DD) |
| `order_by` | `amount` or `date` |
| `order_dir` | `asc` or `desc` |
| `group_by` | `category`, `subcategory`, or `user` |
| `aggregates_only` | `true` to return only totals (SUM, COUNT) |

**Examples:**
```
GET /api/v1/expenses?user_id=1&category_id=2,3&order_by=amount&order_dir=desc
GET /api/v1/expenses?group_by=category&aggregates_only=true
GET /api/v1/expenses?date_from=2025-01-01&date_to=2025-01-31
```

## Environment Variables

**Required:**
- `MYSQL_HOST` - Database host
- `MYSQL_USER` - Database user
- `MYSQL_PASSWORD` - Database password
- `MYSQL_DATABASE` - Database name

**Optional:**
- `MYSQL_PORT` - Database port
- `PORT` - Server port (default: 8082)
- `HOST` - Server host
- `JWT_SECRET` - JWT signing key (change in production!)

## Development

```bash
# Run locally
go run .

# Build
go build -o server .

# Docker
docker build -t expense-tracker-api .
docker run -p 8082:8082 --env-file .env expense-tracker-api
```

## Deployment

Deployed on a Hetzner instance. **No CI/CD automation** - deployment is manual.

**Process:**
1. Develop locally on your machine
2. `git push` to GitHub (`github.com/j0sephh123/expense-tracker-api`)
3. SSH into Hetzner server
4. `git pull` to get latest changes
5. `docker build -t expense-tracker-api .`
6. `docker run -p 8082:8082 --env-file .env expense-tracker-api`

GitHub serves as the middleman to transfer code from dev machine to server.

## Database Migrations

Migrations run automatically on app startup via `migrations.go`.

- SQL files in `migrations/` folder are embedded into the binary at compile time
- `schema_migrations` table tracks which migrations have been applied
- New migrations: add numbered SQL files (e.g., `002_add_feature.sql`)
- Safe to run multiple times - only pending migrations are applied

## Security

- Bcrypt password hashing
- JWT tokens with 7-day expiration (HS256)
- Parameterized SQL queries (injection protection)
- Generic auth error messages (no email enumeration)
- Password excluded from JSON responses
- Distroless Docker image

## Currency

1 EUR = 1.95583 BGN (Bulgarian Lev fixed exchange rate)

## Key Implementation Notes

- Manual routing via path prefix matching (no router library)
- Constraint enforcement: cannot delete categories with subcategories, subcategories with expenses
- Nullable fields handled with `sql.Null*` types and pointers
- All errors logged with timestamps to stdout/stderr
