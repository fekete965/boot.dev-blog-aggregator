# Boot.dev Blog Aggregator

A Go-based blog aggregator CLI application.

## Database Management

This project uses **Goose** for database migrations and **SQLC** for type-safe SQL code generation.

### Goose Migrations

**Goose** is used to manage database schema migrations. All migration files are located in the `sql/schema/` directory.

#### Running Migrations

To run migrations, use the goose CLI tool:

```bash
# Run all pending migrations
goose -dir sql/schema postgres "your-database-url" up

# Rollback the last migration
goose -dir sql/schema postgres "your-database-url" down

# Check migration status
goose -dir sql/schema postgres "your-database-url" status
```

#### Creating New Migrations

Migration files follow the naming pattern `XXX_description.sql` (e.g., `001_users.sql`). Each migration file should include:

- `-- +goose up` comment followed by the migration SQL
- `-- +goose down` comment followed by the rollback SQL

Example:
```sql
-- +goose up
CREATE TABLE users (
  id UUID PRIMARY KEY NOT NULL,
  name VARCHAR(255) NOT NULL
);

-- +goose down
DROP TABLE users;
```

### SQLC Code Generation

**SQLC** is used to generate type-safe Go code from SQL queries. This ensures compile-time safety and eliminates the need for manual struct mapping.

#### Generating Code

After writing SQL queries, run SQLC to generate the corresponding Go code:

```bash
sqlc generate
```

#### SQLC Configuration

SQLC reads from a configuration file (typically `sqlc.yaml`) that defines:
- Database type (PostgreSQL, MySQL, etc.)
- SQL query locations
- Generated code output paths
- Code generation settings

Make sure to configure SQLC properly before generating code.

## Project Structure

```
.
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
├── internal/
│   └── config/            # Configuration management
└── sql/
    └── schema/            # Goose migration files
```

## Configuration

The application stores configuration in `~/.gatorconfig.json` with the following structure:

```json
{
  "db_url": "postgres://user:password@localhost/dbname",
  "current_user_name": "username"
}
```

## Usage

```bash
# Login as a user
gator login <username>
```

## Development

### Prerequisites

- Go 1.23.2 or later
- PostgreSQL database
- Goose CLI tool
- SQLC CLI tool

### Setup

1. Clone the repository
2. Install dependencies: `go mod download`
3. Configure your database URL in `~/.gatorconfig.json`
4. Run migrations: `goose -dir sql/schema postgres "$DB_URL" up`
5. Generate SQLC code: `sqlc generate`
6. Build and run: `go build && ./boot.dev-blog-aggregator`

