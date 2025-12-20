# hbm

**hbm** is a minimal SQL migration library and CLI for Go, designed to be simple, explicit, and easy to reuse across projects.

It focuses on:
- File-based SQL migrations
- PostgreSQL
- Batch-based rollbacks
- Zero magic

This tool is intentionally minimal and used in real projects.

---

## Features

- SQL file migrations
- Ordered execution by filename
- Batch-based `up` / `down` migrations
- PostgreSQL support
- CLI + reusable Go library
- No framework or DSL

---

## Installation (CLI)

Install the CLI globally using Go:

```bash
go install github.com/hobrt-programming/hbm/cmd/hbm@latest
```

Make sure your Go bin directory is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## Database Configuration

`hbm` **does not hardcode database credentials**.

You must provide the database connection string using the environment variable:

```bash
export DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
```

This is required for all migration commands.

If `DATABASE_URL` is not set, the CLI will exit with an error.

---

## Usage

### Create a migration

```bash
hbm create name_of_the_migration
```

This creates a new file in the `migrations/` directory:

```
migrations/
└── 20250101123000_name_of_the_migration.hbm.sql
```

Generated file template:

```sql
-- +hbm Up
-- SQL in section 'Up' is executed when this migration is applied

-- +hbm Down
-- SQL in section 'Down' is executed when this migration is rolled back
```

---

### Run migrations (up)

Apply all unapplied migrations:

```bash
hbm migrate up
```

What happens:
- Migrations are executed in filename order
- A new batch is created
- Each migration runs inside a transaction
- Applied migrations are recorded in `schema_migrations`

---

### Rollback last batch (down)

Rollback the **last applied batch only**:

```bash
hbm migrate down
```

What happens:
- Only the most recent batch is rolled back
- Each rollback runs inside a transaction
- Migration records are removed accordingly

If there is nothing to rollback, you will be informed.

---

## Migration Table

`hbm` automatically creates and manages this table if it does not exist:

```sql
schema_migrations
```

Columns:
- `id`
- `file_name`
- `batch`
- `applied_at`
- `checksum`

You normally do **not** need to interact with this table directly.

---

## Library Usage (optional)

You can also use `hbm` as a Go library:

```bash
go get github.com/hobrt-programming/hbm@latest
```

Example:

```go
db, err := sqlx.Connect("postgres", dsn)
if err != nil {
	log.Fatal(err)
}
defer db.Close()

res, err := hbm.RunMigrations(db, "up")
if err != nil {
	log.Fatal(err)
}

if res == hbm.ResultNothingToDo {
	log.Println("Nothing to migrate")
}
```

---

## Philosophy

- No global state
- No hidden configuration
- No automatic execution
- You control when migrations run

`hbm` is designed to be predictable and easy to understand.

---

## License

MIT