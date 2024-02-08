package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "your-password"
	dbname   = "wallet"
)

type SqlDataBase struct {
	DB *sql.DB
}

func dsn(dbName string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
}

func DbConnection() (*SqlDataBase, error) {
	db, err := sql.Open("postgres", dsn("postgres"))
	if err != nil {
		log.Printf("Error %s when opening DB\n", err)
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT datname FROM pg_catalog.pg_database WHERE datname='wallet'")
	if err != nil {
		log.Printf("Error %s when creating DB\n", err)
	}
	defer rows.Close()

	if !rows.Next() {
		_, err = db.Exec("CREATE DATABASE wallet")
		if err != nil {
			log.Printf("Error %s when creating DB\n", err)
		}
		log.Println("Database 'wallet' created successfully")
	}

	db, err = sql.Open("postgres", dsn(dbname))
	if err != nil {
		log.Printf("Error %s when opening DB", err)
		return nil, err
	}

	scriptPath := "script.sql" // Provide the path to your SQL script file
	err = executeSQLScript(db, scriptPath)
	if err != nil {
		log.Printf("Error when executing sql script")
		return nil, err
	}

	fmt.Println("SQL script executed successfully!")

	sqlDataBase := SqlDataBase{DB: db}
	return &sqlDataBase, nil
}

func executeSQLScript(db *sql.DB, scriptPath string) error {
	// Read SQL script from file
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		log.Println("failed to read SQL script")
		return fmt.Errorf("failed to read SQL script: %v", err)
	}

	// Execute SQL statements from script
	queries := string(script)
	_, err = db.Exec(queries)
	if err != nil {
		log.Println("failed to execute SQL script")
		return fmt.Errorf("failed to execute SQL script: %v", err)
	}

	return nil
}
