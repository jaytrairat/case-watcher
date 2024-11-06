package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/fsnotify/fsnotify"
)

// WatchDir starts watching the specified directory for changes
func WatchDir(dirPath string) error {
	// Initialize a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	// Define the regex pattern to match folders like F-YYYY-001
	folderPattern := regexp.MustCompile(`^F-\d{4}-\d{3}$`)

	// Start a goroutine to handle events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Check if the event is for folder creation
				if event.Op&fsnotify.Create == fsnotify.Create {
					// Check if the created item matches the folder pattern
					if folderPattern.MatchString(filepath.Base(event.Name)) {
						fmt.Printf("New folder matching pattern created: %s\n", event.Name)
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

	// Block the main goroutine to keep watching indefinitely
	select {}
}

func main() {
	// Example usage
	dirPath := "."

	// Create the directory if it doesn't exist
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0755)
		fmt.Println("Created directory:", dirPath)
	}

	err := WatchDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}
}
