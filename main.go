package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
    "time"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Workout struct {
	ID        uint              `gorm:"primaryKey" json:"id"`
    UserID    uint              `json:"userId"`
	Name      string            `json:"name"`
	Date      time.Time         `json:"date"`
	Exercises []WorkoutExercise `json:"exercises"`
}

type WorkoutExercise struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	WorkoutID  uint       `json:"workoutId"`
	ExerciseID string     `json:"exerciseId"`
	Sets       int        `json:"sets"`
	Reps       []int      `gorm:"serializer:json" json:"reps"`
	Weight     []float64  `gorm:"serializer:json" json:"weight"`     // In kg
	Exercise   Exercise   `gorm:"foreignKey:ExerciseID" json:"exercise"`
}

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


type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Email     string    `gorm:"unique" json:"email"`
	Age       int       `json:"age"`
	Height    float64   `json:"height"`
	Weight    float64   `json:"weight"`
	CreatedAt time.Time `json:"createdAt"`
	Workouts  []Workout `json:"workouts,omitempty"`
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

	err = db.AutoMigrate(&User{}, &Exercise{}, &Workout{}, &WorkoutExercise{})
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
	InitDB()
	SeedDBIfEmpty("exercises.json")

	http.HandleFunc("/api/exercises", GetExercises)
	http.HandleFunc("/api/workouts", CreateWorkout)
	http.HandleFunc("/api/users", CreateUser)
	http.HandleFunc("/api/users/{id}/workouts", GetUserWorkouts)
	http.HandleFunc("DELETE /api/workouts/{id}", DeleteWorkout)
	http.HandleFunc("PUT /api/workouts/{id}", UpdateWorkout)

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	port := ":8080"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
