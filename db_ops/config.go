package db_ops

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func InitiateDatabase() error {
	conn, err := Connect()

	if err != nil {
		fmt.Printf("cannot connect to db: %s\n", err)
		return err
	}

	defer conn.Close()
	fmt.Println("connected to the database...")

	statements, err := convertSQLFileToStatements("db_ops/structure.sql")

	if err != nil {
		log.Fatal(err)
	}

	tx, err := conn.Begin()

	if err != nil {
		fmt.Printf("cannot begin transaction in db: %s\n", err)
		return err
	}

	defer tx.Rollback()

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("executing statement (%s): %w", stmt, err)
		}
	}

	if err = tx.Commit(); err != nil {
		fmt.Printf("cannot commit transaction: %s\n", err)
		return err
	}
	fmt.Println("completed table initiation...")
	return nil
}

func convertSQLFileToStatements(filename string) ([]string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot get working dir: %w", err)
	}

	fileContents, err := os.ReadFile(fmt.Sprintf("%s/%s", wd, filename))
	if err != nil {
		return nil, fmt.Errorf("cannot read sql file: %w", err)
	}

	content := string(fileContents)
	var filteredStatements []string

	// First, extract the trigger definition
	triggerStart := strings.Index(content, "DELIMITER //")
	triggerEnd := strings.Index(content, "DELIMITER ;")

	if triggerStart != -1 && triggerEnd != -1 {
		// Get the trigger statement without DELIMITER commands
		triggerContent := strings.TrimSpace(content[triggerStart+len("DELIMITER //") : triggerEnd])
		// Remove the END;// and add END;
		triggerContent = strings.TrimSuffix(triggerContent, "END;//")
		triggerContent = triggerContent + "END;"

		// Split the content before the trigger
		beforeTrigger := content[:triggerStart]
		statements := strings.Split(beforeTrigger, ";")

		// Add non-empty statements
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				filteredStatements = append(filteredStatements, stmt)
			}
		}

		// Add the trigger statement
		filteredStatements = append(filteredStatements, triggerContent)

	} else {
		// If no trigger found, just split normally
		statements := strings.Split(content, ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				filteredStatements = append(filteredStatements, stmt)
			}
		}
	}

	return filteredStatements, nil
}

func testDbConn(db *sql.DB) error {
	pingErr := db.Ping()

	if pingErr != nil {
		return pingErr
	}

	log.Print("Connected to the database.")
	return nil
}

// func getConnectionString() string {
// 	user := os.Getenv("DB_UNAME")
// 	pass := os.Getenv("DB_PASS")
// 	host := os.Getenv("DB_HOST")
// 	port := os.Getenv("DB_PORT")
// 	dbName := os.Getenv("DB_NAME")

// 	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, pass, host, port, dbName)
// }

func Connect() (*sql.DB, error) {
	var err error
	dbConn, err := sql.Open("mysql", "root@tcp(localhost:3310)/potholes")

	if err != nil {
		return nil, err
	}

	err = testDbConn(dbConn)

	if err != nil {
		return nil, err
	}

	return dbConn, nil
}
