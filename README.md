# Boot.dev Blog Aggregator

A modern Go-based blog aggregator CLI application built with type-safe database queries.

## 🚀 Features

- **Type-safe database queries** using SQLC
- **Database migrations** managed with Goose
- **User management** with registration, login, and listing
- **CLI-first** design for easy automation

## 📋 Prerequisites

- **Go** 1.23.2 or later
- **PostgreSQL** 13+ (for `gen_random_uuid()` support)
- **Goose** CLI tool for migrations
- **SQLC** CLI tool for code generation

### Installing Tools

```bash
# Install Goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Install SQLC
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## 🗄️ Database Management

This project uses **Goose** for database migrations and **SQLC** for type-safe SQL code generation.

### Goose Migrations

**Goose** manages database schema migrations. All migration files are located in `sql/schema/`.

#### Running Migrations

```bash
# Run all pending migrations
goose -dir sql/schema postgres "$DB_URL" up

# Rollback the last migration
goose -dir sql/schema postgres "$DB_URL" down

# Check migration status
goose -dir sql/schema postgres "$DB_URL" status

# Create a new migration
goose -dir sql/schema postgres "$DB_URL" create migration_name sql
```

#### Migration File Structure

Migration files follow the pattern `XXX_description.sql` (e.g., `001_users.sql`). Each file includes:

- `-- +goose up` - Migration SQL
- `-- +goose down` - Rollback SQL

**Example:**
```sql
-- +goose up
CREATE TABLE users (
  id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose down
DROP TABLE users;
```

### SQLC Code Generation

**SQLC** generates type-safe Go code from SQL queries, providing compile-time safety and eliminating manual struct mapping.

#### Workflow

1. Write SQL queries in `sql/queries/` directory
2. Run SQLC to generate Go code:
   ```bash
   sqlc generate
   ```
3. Generated code appears in `internal/database/`

#### SQLC Configuration

The project uses `sqlc.yaml` to configure:
- Database engine (PostgreSQL)
- Schema location (`sql/schema`)
- Query location (`sql/queries`)
- Output directory (`internal/database`)

**Example Query:**
```sql
-- name: CreateUser :one
INSERT INTO users (id, name, created_at, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING *;
```

After running `sqlc generate`, this creates type-safe Go functions in `internal/database/users.sql.go`.

## 📁 Project Structure

```
.
├── main.go                    # Application entry point
├── go.mod                     # Go module dependencies
├── sqlc.yaml                  # SQLC configuration
├── internal/
│   ├── config/               # Configuration management
│   │   └── config.go
│   └── database/             # SQLC generated code
│       ├── db.go
│       ├── models.go
│       └── users.sql.go
└── sql/
    ├── schema/               # Goose migration files
    │   └── 001_users.sql
    └── queries/              # SQLC query files
        └── users.sql
```

## ⚙️ Configuration

The application stores configuration in `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://user:password@localhost/dbname",
  "current_user_name": "username"
}
```

The `db_url` should point to your PostgreSQL database connection string.

## 💻 Usage

### Commands

```bash
# Register a new user
gator register <username>

# Login as an existing user
gator login <username>

# List all users (current user is marked)
gator users

# Reset database (delete all users)
gator reset
```

### Examples

```bash
# Register and automatically login
gator register alice

# Login as existing user
gator login bob

# List all users
gator users
# Output:
# * alice (current)
# * bob
```

## 🛠️ Development

### Initial Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd boot.dev-blog-aggregator
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure database**
   Create `~/.gatorconfig.json` with your PostgreSQL connection string:
   ```json
   {
     "db_url": "postgres://user:password@localhost/dbname"
   }
   ```

4. **Run migrations**
   ```bash
   goose -dir sql/schema postgres "$(jq -r .db_url ~/.gatorconfig.json)" up
   ```

5. **Generate SQLC code**
   ```bash
   sqlc generate
   ```

6. **Build and run**
   ```bash
   go build -o gator
   ./gator register testuser
   ```

### Development Workflow

1. **Add a new migration:**
   ```bash
   goose -dir sql/schema postgres "$DB_URL" create add_posts_table sql
   ```

2. **Write SQL queries** in `sql/queries/`

3. **Generate code:**
   ```bash
   sqlc generate
   ```

4. **Use generated code** in your Go application

### Dependencies

- `github.com/lib/pq` - PostgreSQL driver
- `github.com/google/uuid` - UUID generation
- SQLC generated code in `internal/database/`

## 📝 Notes

- UUIDs are auto-generated using PostgreSQL's `gen_random_uuid()` function
- User names must be unique (enforced by database constraint)
- Timestamps are automatically managed with `CURRENT_TIMESTAMP` defaults
- The current user is stored in the config file, not the database
