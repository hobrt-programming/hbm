package main

import (
	"fmt"
	"os"

	"github.com/hobrt-programming/hbm"
	"github.com/jmoiron/sqlx"
)

func openDB() (*sqlx.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	return sqlx.Connect("postgres", dsn)
}

func main() {
	// This is a placeholder main package.
	// The actual application entry point is in cmd/api/main.go

	if len(os.Args) < 2 {
		fmt.Println("Please provide an operation: migrate or create")
		os.Exit(1)
	}

	operation := os.Args[1]

	switch operation {
	case "migrate":
		fmt.Println("Running database migrations...")

		if len(os.Args) < 3 {
			fmt.Println("Please provide a name for the migration file")
			os.Exit(1)
		}

		direction := os.Args[2]

		if direction != "up" && direction != "down" {
			fmt.Printf("Please use up or down for migrations")
		}

		db, err := openDB()
		if err != nil {
			fmt.Println("Error connecting to the database:", err)
			os.Exit(1)
		}
		defer db.Close()

		// Here you would call the actual migration logic
		res, err := hbm.RunMigrations(db, direction)

		if err != nil {
			fmt.Println("Error in RunMigrations :", err)
			os.Exit(1)
		}

		switch res {
		case hbm.ResultApplied:
			fmt.Println("Migrations applied.")
		case hbm.ResultNothingToDo:
			fmt.Println("Nothing to migrate/rollback")
		}

		fmt.Println("Done ...")

	case "create":

		if len(os.Args) < 3 {
			fmt.Println("Please provide a name for the migration file")
			os.Exit(1)
		}

		name := os.Args[2]
		fmt.Println("Creating new migration file with name:", name)
		err := hbm.CreateMigrationFile(name)
		if err != nil {
			fmt.Println("Error creating migration file:", err)
			os.Exit(1)
		} else {
			fmt.Println("Migration file created successfully")
		}

	default:
		fmt.Println("Unknown operation:", operation)
	}

}
