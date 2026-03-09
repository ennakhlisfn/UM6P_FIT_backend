package main

import (
    "time"
	"github.com/lib/pq"
)

type Workout struct {
	ID        uint              `gorm:"primaryKey" json:"id"`
    UserID    uint              `json:"userId"`
	Name      string            `json:"name"`
	Date      time.Time         `json:"date"`
	Exercises []WorkoutExercise `json:"exercises"`
}

type Set struct {
	Reps   int     `json:"reps"`
	Weight float64 `json:"weight"`
}

type WorkoutExercise struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	WorkoutID  uint       `json:"workoutId"`
	ExerciseID string     `json:"exerciseId"`
    Sets       []Set      `gorm:"serializer:json" json:"sets"`
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
    Password  string    `json:"-"`
	Age       int       `json:"age"`
	Height    float64   `json:"height"`
	Weight    float64   `json:"weight"`
	CreatedAt time.Time `json:"createdAt"`
	Workouts  []Workout `json:"workouts,omitempty"`
}

type WorkoutTemplate struct {
	ID        uint               `gorm:"primaryKey" json:"id"`
	Name      string             `json:"name"`
	Type      string             `json:"type"` 
	CreatedBy uint               `json:"createdBy"` // 0 = Admin/System, 1+ = Specific User
	Exercises []TemplateExercise `gorm:"foreignKey:TemplateID" json:"exercises"`
}

type TemplateExercise struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	TemplateID uint   `json:"templateId"`
	ExerciseID string `json:"exerciseId"`
	TargetSets int    `json:"targetSets"` 
}
