package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/jaytrairat/case-watcher/cfuncs"
	_ "github.com/mattn/go-sqlite3"
)

// APIUrl is the URL of the API to call when a new folder is detected
const APIUrl = "http://policeadmin.com:8092/broadcast"

// APIKey is the API key for authorization
const APIKey = "LDabxoSBFmiedZI2w7o0dVIXbfQnzKV9Bgwy7YNWyfIlB7TWFXPAXS1A1oCN4hNQej7lKxPezvFLYQCtG6f38mAGUw2gKmix71zvw4i5KAJUlHpsPheLF9Q5pgTaUPBi"

// Global variable for the current working directory
var cwd string

// WatchDir starts watching the specified directory for changes
func WatchDir(ctx context.Context, db *sql.DB, dirPath string) error {
	// Initialize a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	// Define the regex pattern to match folders like F-YYYY-001 or 001
	folderPattern := regexp.MustCompile(`^(F-\d{4}-\d{3}|\w{1}\d{3})$`)

	// Channel to debounce folder creation events
	lastEvent := time.Now()

	// Start a goroutine to handle events
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Stopping watcher...")
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Check if the event is for folder creation
				if event.Op&fsnotify.Create == fsnotify.Create {
					// Check if the created item matches the folder pattern
					if folderPattern.MatchString(filepath.Base(event.Name)) {
						// Check the time difference since the last event
						if time.Since(lastEvent) < time.Second {
							// Skip this event if it's too soon after the last one
							continue
						}

						// Update last event time
						lastEvent = time.Now()

						logEntry := fmt.Sprintf("New folder created: %s at %s\n", event.Name, time.Now().Format(time.RFC3339))
						fmt.Print(logEntry)

						// Insert the log entry into the database
						_, err := db.Exec("INSERT INTO folder_logs (folder_name, created_at) VALUES (?, ?)", event.Name, time.Now())
						if err != nil {
							log.Println("ERROR writing to database:", err)
						}

						// Call the API to send a message
						message := fmt.Sprintf("มีโฟลเดอร์ Case ใหม่ชื่อ %s\nสร้างเมื่อ %s เวลา %s น.", filepath.Base(event.Name), time.Now().AddDate(543, 0, 0).Format("02 มกราคม 2006"), time.Now().Format("03.04"))
						if err := cfuncs.SendAPIRequest(message); err != nil {
							log.Println("ERROR sending API request:", err)
						}
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("ERROR:", err)
			}
		}
	}()

	// Walk through the directory and add all subfolders to the watcher
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add path to watcher: %w", err)
	}

	fmt.Println("Current working directory:", cwd) // Print the current working directory

	// Block until the context is canceled
	<-ctx.Done()
	return nil
}

func main() {
	// Example usage
	dirPath := `.`

	// Get the current working directory and assign to the global cwd variable
	var err error
	cwd, err = os.Getwd()
	if err != nil {
		log.Fatal("ERROR getting current working directory:", err)
	}

	// Set up context with cancel functionality
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for interrupt signals to stop the watcher
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	db := cfuncs.InitDB()

	// Run the watcher in a goroutine
	go func() {
		if err := WatchDir(ctx, db, dirPath); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for a stop signal
	<-stop
	fmt.Println("Received interrupt signal, shutting down...")
	cancel()                    // Cancel the watcher context to stop it gracefully
	time.Sleep(1 * time.Second) // Give it a moment to clean up
}
