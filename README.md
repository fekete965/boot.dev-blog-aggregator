# 🐊 Gator - Blog Aggregator CLI

> A modern, type-safe RSS feed aggregator built with Go. Collect, organize, and browse blog posts from your favorite RSS feeds all from the command line.

[![Go Version](https://img.shields.io/badge/go-1.23.2-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## ✨ Features

- 🔐 **User Management** - Register, login, and manage multiple users
- 📰 **RSS Feed Aggregation** - Add and follow RSS feeds from any source
- 🤖 **Automatic Scraping** - Periodically fetch and store new posts
- 📖 **Post Browsing** - Browse posts from your followed feeds in a beautiful CLI interface
- 🗄️ **Type-Safe Database** - Built with SQLC for compile-time query safety
- 🔄 **Database Migrations** - Managed with Goose for version-controlled schema changes
- ⚡ **Statically Compiled** - Single binary, no runtime dependencies

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

- **Go** 1.23.2 or later ([Download](https://golang.org/dl/))
- **PostgreSQL** 13+ ([Download](https://www.postgresql.org/download/))
  - Required for `gen_random_uuid()` support

### Verify Installation

```bash
# Check Go version
go version

# Check PostgreSQL version
psql --version
```

## 🚀 Installation

### Install the Gator CLI

Install `gator` globally using Go's `install` command:

```bash
go install github.com/fekete965/boot.dev-blog-aggregator@latest
```

This will compile and install the `gator` binary to your `$GOPATH/bin` or `$HOME/go/bin` directory. Make sure this directory is in your `PATH`.

**Note:** Go programs are statically compiled binaries. After installation, you can run `gator` from anywhere without needing the Go toolchain.

### Verify Installation

```bash
gator --help
# or simply
gator
```

## ⚙️ Configuration

Gator stores its configuration in `~/.gatorconfig.json` in your home directory.

### Initial Setup

1. **Create the configuration file:**

```bash
cat > ~/.gatorconfig.json << EOF
{
  "db_url": "postgres://username:password@localhost/dbname?sslmode=disable"
}
EOF
```

2. **Set up your PostgreSQL database:**

```bash
# Create a database (if you haven't already)
createdb blog_aggregator

# Or using psql
psql -U postgres -c "CREATE DATABASE blog_aggregator;"
```

3. **Run database migrations:**

You'll need to install [Goose](https://github.com/pressly/goose) for migrations:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Then run migrations:

```bash
DB_URL=$(jq -r .db_url ~/.gatorconfig.json)
goose -dir sql/schema postgres "$DB_URL" up
```

### Configuration File Format

```json
{
  "db_url": "postgres://user:password@localhost:5432/dbname?sslmode=disable",
  "current_user_name": "alice"
}
```

- `db_url`: PostgreSQL connection string (required)
- `current_user_name`: Currently logged-in user (set automatically)

## 💻 Usage

### Quick Start

```bash
# 1. Register a new user (automatically logs you in)
gator register alice

# 2. Add an RSS feed
gator addfeed "Boot.dev Blog" https://blog.boot.dev/index.xml

# 3. Start aggregating feeds (runs continuously)
gator agg 30s

# 4. Browse your posts
gator browse 10
```

### Available Commands

#### User Management

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

#### Feed Management

```bash
# Add a new RSS feed (requires login)
gator addfeed <feed_name> <feed_url>

# List all feeds in the system
gator feeds

# Follow an existing feed (requires login)
gator follow <feed_url>

# List feeds you're following (requires login)
gator following

# Unfollow a feed (requires login)
gator unfollow <feed_url>
```

#### Post Aggregation & Browsing

```bash
# Aggregate feeds continuously (requires feeds to exist)
# Time format: "30s", "5m", "1h", etc.
gator agg <time_between_requests>

# Browse posts from your followed feeds (requires login)
# Default limit is 2 posts
gator browse [limit]
```

### Command Examples

```bash
# Register and login
gator register alice
# Output: New user registered: alice

# Add a feed
gator addfeed "Go Blog" https://go.dev/blog/feed.atom
# Output: Feed successfully added
#         - Id: 550e8400-e29b-41d4-a716-446655440000
#         - User ID: ...
#         - Name: Go Blog
#         - URL: https://go.dev/blog/feed.atom
#         Successfully followed the feed: Go Blog

# List your feeds
gator following
# Output: Current user is following 1 feed
#         - Go Blog

# Aggregate feeds every 30 seconds
gator agg 30s
# Output: Collecting feeds every 30s
#         Post successfully created: Go 1.23 Release Notes
#         Post successfully created: Working with Go Modules
#         ...

# Browse latest 5 posts
gator browse 5
# Output: ┌─────────────────────────────────────────────────────────────┐
#         │ Title: Go 1.23 Release Notes                                │
#         │ Published At: 15 August 2024 10:30                          │
#         ├─────────────────────────────────────────────────────────────┤
#         │ Description:                                                │
#         │ Go 1.23 brings new features and improvements...             │
#         ├─────────────────────────────────────────────────────────────┤
#         │ URL: https://go.dev/blog/go1.23                             │
#         └─────────────────────────────────────────────────────────────┘
```

## 🛠️ Development

### Project Structure

```
.
├── main.go                    # Application entry point & CLI handlers
├── go.mod                     # Go module dependencies
├── sqlc.yaml                  # SQLC configuration
├── internal/
│   ├── config/               # Configuration management
│   │   └── config.go
│   └── database/             # SQLC generated code
│       ├── db.go
│       ├── models.go
│       ├── users.sql.go
│       ├── feeds.sql.go
│       ├── feed_follows.sql.go
│       └── posts.sql.go
└── sql/
    ├── schema/               # Goose migration files
    │   ├── 001_users.sql
    │   ├── 002_feeds.sql
    │   ├── 003_feed_follows.sql
    │   ├── 004_feeds.sql
    │   └── 005_posts.sql
    └── queries/              # SQLC query files
        ├── users.sql
        ├── feeds.sql
        ├── feed_follows.sql
        └── posts.sql
```

### Development Setup

1. **Clone the repository:**

```bash
git clone https://github.com/your-username/boot.dev-blog-aggregator.git
cd boot.dev-blog-aggregator
```

2. **Install dependencies:**

```bash
go mod download
```

3. **Install development tools:**

```bash
# Install Goose for migrations
go install github.com/pressly/goose/v3/cmd/goose@latest

# Install SQLC for code generation
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

4. **Set up configuration:**

Create `~/.gatorconfig.json` with your database connection string.

5. **Run migrations:**

```bash
DB_URL=$(jq -r .db_url ~/.gatorconfig.json)
goose -dir sql/schema postgres "$DB_URL" up
```

6. **Generate SQLC code:**

```bash
sqlc generate
```

7. **Run during development:**

```bash
# Use go run for development
go run . register testuser

# Or build and run
go build -o gator
./gator register testuser
```

### Development Workflow

#### Adding a New Migration

```bash
DB_URL=$(jq -r .db_url ~/.gatorconfig.json)
goose -dir sql/schema postgres "$DB_URL" create add_new_table sql
```

Edit the generated migration file in `sql/schema/`, then run:

```bash
goose -dir sql/schema postgres "$DB_URL" up
```

#### Adding a New Query

1. Add SQL query to `sql/queries/<table>.sql`:

```sql
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;
```

2. Generate Go code:

```bash
sqlc generate
```

3. Use the generated function in your code:

```go
user, err := dbQueries.GetUserByID(ctx, userID)
```

### Dependencies

- `github.com/lib/pq` - PostgreSQL driver
- `github.com/google/uuid` - UUID generation
- SQLC generated code in `internal/database/`

## 📚 Architecture

### Database Schema

- **users** - User accounts
- **feeds** - RSS feed definitions
- **feed_follows** - User-feed relationships
- **posts** - Aggregated blog posts

### Key Design Decisions

- **Type-Safe Queries**: SQLC generates Go code from SQL, providing compile-time safety
- **Migration Management**: Goose handles schema versioning
- **CLI-First**: Designed for automation and scripting
- **Stateless Aggregation**: Feed scraping runs independently of user sessions

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built as part of the [Boot.dev](https://boot.dev) curriculum
- Uses [SQLC](https://sqlc.dev/) for type-safe SQL
- Uses [Goose](https://github.com/pressly/goose) for migrations

---

**Note:** `go run .` is for development only. For production use, install `gator` using `go install` and run it as a statically compiled binary.
