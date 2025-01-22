package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	dsn := flag.String("dsn", "", "Database connection string (DSN)")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("DSN is required. Use the --dsn flag to provide it.")
	}

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		log.Fatalf("Can not connect to db: %v", err)
	}
	defer db.Close()

	insertRoles(db)
}

func insertRoles(db *sql.DB) {
	roles := []string{"admin", "customer"}

	for _, role := range roles {
		query := `INSERT INTO role (name)
				VALUES ($1)
				ON CONFLICT (name) DO NOTHING`
		_, err := db.Exec(query, role)

		if err != nil {
			log.Printf("Err while seed role '%s': %v", role, err)
			panic(fmt.Sprintf("Err while seed role '%s': %v", role, err))
		} else {
			log.Printf("Role '%s' successfully added", role)
		}
	}
}
