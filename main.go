package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

// LogFile is the name of the log file where created folders will be recorded
const LogFile = "created_folders.log"

// WatchDir starts watching the specified directory for changes
func WatchDir(ctx context.Context, dirPath string) error {
	// Initialize a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	// Define the regex pattern to match folders like F-YYYY-001
	folderPattern := regexp.MustCompile(`^F-\d{4}-\d{3}$`)

	// Open or create the log file
	logFile, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

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
						logEntry := fmt.Sprintf("New folder created: %s at %s\n", event.Name, time.Now().Format(time.RFC3339))
						fmt.Print(logEntry)

						// Write the log entry to the log file
						if _, err := logFile.WriteString(logEntry); err != nil {
							log.Println("ERROR writing to log file:", err)
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

	fmt.Println("Watching directory:", dirPath)

	// Block until the context is canceled
	<-ctx.Done()
	return nil
}

func main() {
	// Example usage
	dirPath := "."

	// Create the directory if it doesn't exist
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0755)
		fmt.Println("Created directory:", dirPath)
	}

	// Set up context with cancel functionality
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for interrupt signals to stop the watcher
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run the watcher in a goroutine
	go func() {
		if err := WatchDir(ctx, dirPath); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for a stop signal
	<-stop
	fmt.Println("Received interrupt signal, shutting down...")
	cancel()                    // Cancel the watcher context to stop it gracefully
	time.Sleep(1 * time.Second) // Give it a moment to clean up
}
