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

func WatchDir(ctx context.Context, db *sql.DB, dirPath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	folderPattern := regexp.MustCompile(`^(F-\d{4}-\d{3}|\w{1}\d{3})$`)

	lastEvent := time.Now()

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

				if event.Op&fsnotify.Create == fsnotify.Create {
					if folderPattern.MatchString(filepath.Base(event.Name)) {
						if time.Since(lastEvent) < time.Second {
							continue
						}

						lastEvent = time.Now()
						logEntry := fmt.Sprintf("New folder created: %s at %s\n", event.Name, time.Now().Format(time.RFC3339))
						fmt.Print(logEntry)

						_, err := db.Exec("INSERT INTO folder_logs (folder_name, created_at) VALUES (?, ?)", event.Name, time.Now())
						if err != nil {
							log.Println("ERROR writing to database:", err)
						}

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

	fmt.Println("Current working directory:", cwd)

	<-ctx.Done()
	return nil
}

func main() {
	dirPath := `.`

	var err error
	cwd, err = os.Getwd()
	if err != nil {
		log.Fatal("ERROR getting current working directory:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	db := cfuncs.InitDB()

	go func() {
		if err := WatchDir(ctx, db, dirPath); err != nil {
			log.Fatal(err)
		}
	}()

	<-stop
	fmt.Println("Received interrupt signal, shutting down...")
	cancel()
	time.Sleep(1 * time.Second)
}
