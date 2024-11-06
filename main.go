package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/fsnotify/fsnotify"
)

const LogFile = "created_folders.log"

func WatchDir(dirPath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	folderPattern := regexp.MustCompile(`^F-\d{4}-\d{3}$`)

	logFile, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer logFile.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if folderPattern.MatchString(filepath.Base(event.Name)) {
						logEntry := fmt.Sprintf("New folder created: %s at %s\n", event.Name, time.Now().Format(time.RFC3339))
						fmt.Print(logEntry)

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

	select {}
}

func main() {
	dirPath := "."

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.Mkdir(dirPath, 0755)
		fmt.Println("Created directory:", dirPath)
	}

	err := WatchDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}
}
