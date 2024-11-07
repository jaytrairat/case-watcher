package cfuncs

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

const DatabaseFile = "created_folders.db"

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

func GetLastTimestamp(db *sql.DB) (time.Time, error) {
	var lastTimestamp time.Time
	row := db.QueryRow("SELECT created_at FROM folder_logs ORDER BY created_at DESC LIMIT 1")
	err := row.Scan(&lastTimestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return lastTimestamp, nil
}

func ShouldSendAPIRequest(db *sql.DB) bool {
	lastTimestamp, err := GetLastTimestamp(db)
	if err != nil || lastTimestamp.IsZero() || time.Since(lastTimestamp) >= 1*time.Second {
		return true
	}
	fmt.Printf("%s\n", lastTimestamp)
	return false
}
