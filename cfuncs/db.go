package cfuncs

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// DatabaseFile is the SQLite database file where logs will be stored
const DatabaseFile = "created_folders.db"

// InitDB initializes the SQLite database and creates the required table if it doesn't exist.
func InitDB() *sql.DB {
	db, err := sql.Open("sqlite", DatabaseFile)
	if err != nil {
		log.Fatal("ERROR opening SQLite database:", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS folder_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		folder_name TEXT,
		created_at TIMESTAMP
	)`)
	if err != nil {
		log.Fatal("ERROR creating table in SQLite database:", err)
	}

	return db
}
