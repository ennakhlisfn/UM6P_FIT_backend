package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


var db *gorm.DB

func InitDB() {
	dsn := "host=localhost user=sennakhl password=postgres dbname=um6p_fit port=5432 sslmode=disable TimeZone=UTC"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	fmt.Println("Connected to PostgreSQL successfully!")

	err = db.AutoMigrate(&User{}, &Exercise{}, &Workout{}, &WorkoutExercise{}, &WorkoutTemplate{}, &TemplateExercise{})
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
    http.HandleFunc("GET /api/users/{id}/exercises/{exId}/progress", GetExerciseProgress)
	http.HandleFunc("GET /api/workout-templates", GetWorkoutTemplates)
	http.HandleFunc("POST /api/workout-templates", CreateWorkoutTemplate)
	http.HandleFunc("PUT /api/workout-templates/{id}", UpdateWorkoutTemplate)
	http.HandleFunc("DELETE /api/workout-templates/{id}", DeleteWorkoutTemplate)

	port := ":8080"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
