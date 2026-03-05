package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Exercise matches the structure of the yuhonas/free-exercise-db and maps directly to PostgreSQL
type Exercise struct {
	ID               string         `gorm:"primaryKey" json:"id"`
	Name             string         `json:"name"`
	Force            string         `json:"force"`
	Level            string         `json:"level"`
	Mechanic         string         `json:"mechanic"`
	Equipment        string         `json:"equipment"`
	PrimaryMuscles   pq.StringArray `gorm:"type:text[]" json:"primaryMuscles"`
	SecondaryMuscles pq.StringArray `gorm:"type:text[]" json:"secondaryMuscles"`
	Instructions     pq.StringArray `gorm:"type:text[]" json:"instructions"`
	Category         string         `json:"category"`
	Images           pq.StringArray `gorm:"type:text[]" json:"images"`
}

var db *gorm.DB

func InitDB() {
	dsn := "host=localhost user=sennakhl password=postgres dbname=um6p_fit port=5432 sslmode=disable TimeZone=UTC"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	fmt.Println("Connected to PostgreSQL successfully!")

	// Auto-migrate the schema
	err = db.AutoMigrate(&Exercise{})
	if err != nil {
		log.Fatalf("Failed to auto-migrate database schema: %v", err)
	}
}

func SeedDBIfEmpty(filepath string) {
	var count int64
	db.Model(&Exercise{}).Count(&count)

	if count == 0 {
		fmt.Println("Database is empty. Seeding from JSON file...")

		jsonFile, err := os.Open(filepath)
		if err != nil {
			log.Fatalf("Failed to open seed file: %v", err)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		var exercises []Exercise
		err = json.Unmarshal(byteValue, &exercises)
		if err != nil {
			log.Fatalf("Failed to parse seed JSON: %v", err)
		}

		// Insert the exercises into the database in batches to avoid overwhelming PostgreSQL query limits
		result := db.CreateInBatches(exercises, 100)
		if result.Error != nil {
			log.Fatalf("Failed to seed database: %v", result.Error)
		}

		fmt.Printf("Successfully seeded %d exercises into PostgreSQL!\n", result.RowsAffected)
	} else {
		fmt.Printf("Database already contains %d exercises. Skipping seed step.\n", count)
	}
}

func main() {
	// 1. Initialize DB and Seed Data
	InitDB()
	SeedDBIfEmpty("exercises.json")

	// 2. Set up HTTP handlers
	// Serve the exercises data dynamically from PostgreSQL
	http.HandleFunc("/api/exercises", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var exercises []Exercise
		// Query the database, limiting to 50 for performance
		result := db.Limit(50).Find(&exercises)
		if result.Error != nil {
			http.Error(w, "Failed to fetch exercises from database", http.StatusInternalServerError)
			return
		}

		err := json.NewEncoder(w).Encode(exercises)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Serve the static frontend files from the "public" directory
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	// 3. Start the server
	port := ":8080"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
