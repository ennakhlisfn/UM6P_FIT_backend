package handlers

import (
	"fmt"
	"time"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
	"um6p_fit_backend/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB provisions a totally independent in-memory SQLite database 
// and overrides the global DB pointer so testing handlers works seamlessly
// without altering the production PostgreSQL instance.
func SetupTestDB() {
	dbName := fmt.Sprintf("file:memdb_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	err = db.AutoMigrate(
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
		panic("failed to auto-migrate test db")
	}

	database.DB = db
	utils.GlobalCache.Clear() // Evict stale cache items so asserts are pristine
}
