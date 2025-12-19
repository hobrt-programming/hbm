package hbm

import (
	"errors"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Result int

const (
	ResultApplied Result = iota
	ResultNothingToDo
)

type Migration struct {
	ID        int       `db:"id"`
	FileName  string    `db:"file_name"`
	Batch     int       `db:"batch"`
	AppliedAt time.Time `db:"applied_at"`
	Checksum  string    `db:"checksum"`
}

func RunMigrations(db *sqlx.DB, direction string) (Result, error) {

	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		return ResultApplied, errors.New("The folder migrations deos not exist")
	}

	exists := false

	err := db.Get(&exists, `SELECT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'schema_migrations'
    )`)

	if err != nil {
		return ResultApplied, err
	}

	if !exists {
		_, err = db.Exec(`
			CREATE TABLE schema_migrations (
				id SERIAL PRIMARY KEY,
				file_name VARCHAR(255) NOT NULL,
				batch INT NOT NULL,
				applied_at TIMESTAMP NOT NULL DEFAULT NOW(),
				checksum VARCHAR(64) NOT NULL
			);
		`)

		if err != nil {
			return ResultApplied, err
		}
	}

	// Get all migrations
	appliedMigrations := []Migration{}
	err = db.Select(&appliedMigrations, "SELECT * FROM schema_migrations ORDER BY batch DESC")

	if err != nil {
		return ResultApplied, err
	}

	if direction == "up" {

		// Get all migration files from the migrations directory
		files, err := os.ReadDir("migrations")
		if err != nil {
			return ResultApplied, err
		}

		// Sort files by name to ensure they are applied in order
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})

		applied := make(map[string]struct{})

		lastBatch := 0

		for _, m := range appliedMigrations {
			if m.Batch > lastBatch {
				lastBatch = m.Batch
			}
			applied[m.FileName] = struct{}{}
		}

		currentBatch := lastBatch + 1

		// Unapplied migrations
		var notAppliedMigrations []os.DirEntry
		for _, file := range files {
			if _, ok := applied[file.Name()]; ok {
				continue
			}
			notAppliedMigrations = append(notAppliedMigrations, file)
		}

		if len(notAppliedMigrations) == 0 {
			return ResultNothingToDo, nil
		}

		for _, file := range notAppliedMigrations {

			tx, err := db.Beginx()
			if err != nil {
				return ResultApplied, err
			}

			err = RunUpMigrations(tx, file.Name())
			if err != nil {
				tx.Rollback()
				return ResultApplied, err
			}

			_, err = tx.Exec("INSERT INTO schema_migrations(file_name, batch, applied_at, checksum) VALUES($1, $2, $3, $4)", file.Name(), currentBatch, time.Now(), "any")
			if err != nil {
				tx.Rollback()
				return ResultApplied, err
			}

			err = tx.Commit()
			if err != nil {
				return ResultApplied, err
			}
		}

	}

	if direction == "down" {

		if len(appliedMigrations) == 0 {
			return ResultNothingToDo, nil
		}

		lastBatch := appliedMigrations[0].Batch

		for _, m := range appliedMigrations {

			if m.Batch != lastBatch {
				break
			}

			tx, err := db.Beginx()
			if err != nil {
				return ResultApplied, err
			}

			err = RunDownMigrations(tx, m.FileName)
			if err != nil {
				tx.Rollback()
				return ResultApplied, err
			}

			_, err = tx.Exec("DELETE FROM schema_migrations WHERE id = $1", m.ID)
			if err != nil {
				tx.Rollback()
				return ResultApplied, err
			}

			err = tx.Commit()
			if err != nil {
				return ResultApplied, err
			}
		}

	}

	return ResultApplied, nil
}

func RunUpMigrations(db *sqlx.Tx, file string) error {

	// Parse file
	fileContent, err := os.ReadFile("migrations/" + file)

	if err != nil {
		return err
	}

	parts := strings.Split(string(fileContent), "-- +hbm Down")

	if len(parts) != 2 {
		return errors.New("invalid migration file: missing '-- +hbm Down'")
	}

	up := strings.TrimSpace(strings.Replace(parts[0], "-- +hbm Up", "", 1))

	_, err = db.Exec(up)
	if err != nil {
		return err
	}

	return nil
}

func RunDownMigrations(db *sqlx.Tx, file string) error {

	fileContent, err := os.ReadFile("migrations/" + file)

	if err != nil {
		return err
	}

	text := strings.Split(string(fileContent), "-- +hbm Down")

	if len(text) < 2 {
		return errors.New("the has errors")
	}

	_, err = db.Exec(text[1])
	if err != nil {
		return err
	}

	return nil
}

func CreateMigrationFile(name string) error {

	filename := time.Now().Format("20060102150405") + "_" + name + ".hbm.sql"

	file, err := os.Create("migrations/" + filename)

	if err != nil {
		return err
	}
	defer file.Close()

	// Write template content to the file
	template := `-- +hbm Up
-- SQL in section 'Up' is executed when this migration is applied

-- +hbm Down
-- SQL in section 'Down' is executed when this migration is rolled back
`
	_, err = file.WriteString(template)

	if err != nil {
		return err
	}

	return nil
}
