package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"um6p_fit_backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "host=localhost user=sennakhl password=postgres dbname=um6p_fit port=5432 sslmode=disable TimeZone=UTC"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	fmt.Println("Connected to PostgreSQL successfully!")

	err = DB.AutoMigrate(
		&models.User{},
		&models.Exercise{},
		&models.Workout{},
		&models.Set{},
		&models.WorkoutExercise{},
		&models.WorkoutTemplate{},
		&models.TemplateExercise{},
		&models.WorkoutProgram{},
		&models.ProgramDay{},
		&models.UserProgramProgress{},
		&models.UserPointsLog{},
		&models.PersonalRecord{},
		&models.Notification{},
		&models.WeightLog{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database schema: %v", err)
	}
}

func SeedDBIfEmpty(filepath string) {
	var count int64
	DB.Model(&models.Exercise{}).Count(&count)

	if count == 0 {
		fmt.Println("Database is empty. Seeding from JSON file...")

		jsonFile, err := os.Open(filepath)
		if err != nil {
			log.Fatalf("Failed to open seed file: %v", err)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		var exercises []models.Exercise
		err = json.Unmarshal(byteValue, &exercises)
		if err != nil {
			log.Fatalf("Failed to parse seed JSON: %v", err)
		}

		result := DB.CreateInBatches(exercises, 100)
		if result.Error != nil {
			log.Fatalf("Failed to seed database: %v", result.Error)
		}

		fmt.Printf("Successfully seeded %d exercises into PostgreSQL!\n", result.RowsAffected)
	} else {
		fmt.Printf("Database already contains %d exercises. Skipping seed step.\n", count)
	}
}
