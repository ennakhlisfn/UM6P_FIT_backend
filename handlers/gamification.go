package handlers

import (
	"time"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"

	"gorm.io/gorm"
)

func awardPoints(userID uint, points int, reason string) {
	log := models.UserPointsLog{
		UserID:   userID,
		Points:   points,
		Reason:   reason,
		EarnedAt: time.Now(),
	}
	database.DB.Create(&log)
}

func processWorkoutGamification(workout models.Workout) {
	// +10 pts for logging
	awardPoints(workout.UserID, 10, "Workout Logged")

	// +15 pts for exceeding previous session volume
	checkSessionVolume(workout)

	// +30 pts for PR
	checkPersonalRecords(workout)

	// +50 pts for weekly streak
	checkWeeklyStreak(workout.UserID)
}

func checkSessionVolume(currentWorkout models.Workout) {
	var prevWorkout models.Workout
	if err := database.DB.Preload("Exercises").Preload("Exercises.Sets").
		Where("user_id = ? AND date < ?", currentWorkout.UserID, currentWorkout.Date).
		Order("date DESC").First(&prevWorkout).Error; err == nil {

		prevVol := 0.0
		for _, ex := range prevWorkout.Exercises {
			prevVol += calculateVolume(ex.Sets)
		}
		currVol := 0.0
		for _, ex := range currentWorkout.Exercises {
			currVol += calculateVolume(ex.Sets)
		}
		if currVol > prevVol && prevVol > 0 {
			awardPoints(currentWorkout.UserID, 15, "Session Volume Faster/Heavier Than Previous")
		}
	}
}

func checkPersonalRecords(workout models.Workout) {
	for _, ex := range workout.Exercises {
		vol := calculateVolume(ex.Sets)
		maxWeight := 0.0
		for _, s := range ex.Sets {
			if s.Weight > maxWeight {
				maxWeight = s.Weight
			}
		}

		var pr models.PersonalRecord
		err := database.DB.Where("user_id = ? AND exercise_id = ?", workout.UserID, ex.ExerciseID).First(&pr).Error

		if err == gorm.ErrRecordNotFound {
			// New record baseline
			pr = models.PersonalRecord{
				UserID:     workout.UserID,
				ExerciseID: ex.ExerciseID,
				Weight:     maxWeight,
				Volume:     vol,
				AchievedAt: workout.Date,
			}
			database.DB.Create(&pr)
			awardPoints(workout.UserID, 30, "New Personal Record Baseline ("+ex.ExerciseID+")")
		} else {
			// Beaten?
			beaten := false
			if maxWeight > pr.Weight {
				pr.Weight = maxWeight
				beaten = true
			}
			if vol > pr.Volume {
				pr.Volume = vol
				beaten = true
			}
			if beaten {
				pr.AchievedAt = workout.Date
				database.DB.Save(&pr)
				awardPoints(workout.UserID, 30, "Personal Record Shattered! ("+ex.ExerciseID+")")
			}
		}
	}
}

func checkWeeklyStreak(userID uint) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	database.DB.Model(&models.UserPointsLog{}).
		Where("user_id = ? AND reason = 'Weekly Streak Bonus' AND earned_at >= ?", userID, today).
		Count(&count)
	if count > 0 {
		return
	}

	sevenDaysAgo := today.AddDate(0, 0, -6)
	type Result struct {
		LogDate string
	}
	var dates []Result
	database.DB.Table("workouts").
		Select("DISTINCT CAST(date AS DATE) as log_date").
		Where("user_id = ? AND date >= ?", userID, sevenDaysAgo).
		Scan(&dates)

	if len(dates) == 7 {
		awardPoints(userID, 50, "Weekly Streak Bonus")
	}
}
