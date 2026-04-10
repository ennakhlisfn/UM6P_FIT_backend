package jobs

import (
	"log"
	"time"
	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
)

// StartCronJobs boots a lightweight background ticker series 
// that sweeps for logical thresholds and triggers scheduled actions.
func StartCronJobs() {
	log.Println("Initializing Background CRON Scheduler...")

	// Sweep the system every 12 hours
	ticker := time.NewTicker(12 * time.Hour)

	go func() {
		// Run an immediate sweep on boot
		CheckInactiveWeightLogs()

		for range ticker.C {
			CheckInactiveWeightLogs()
		}
	}()
}

// CheckInactiveWeightLogs handles the core Phase 5 14-day inactivity logic
func CheckInactiveWeightLogs() {
	var users []models.User
	
	if result := database.DB.Find(&users); result.Error != nil {
		log.Println("Job Error: failed to fetch users:", result.Error)
		return
	}

	fourteenDaysAgo := time.Now().AddDate(0, 0, -14)

	for _, user := range users {
		var lastLog models.WeightLog
		
		err := database.DB.Where("user_id = ?", user.ID).Order("logged_at DESC").First(&lastLog).Error
		
		var lastUpdated time.Time
		if err != nil {
			lastUpdated = user.CreatedAt // Fallback
		} else {
			lastUpdated = lastLog.LoggedAt
		}

		if lastUpdated.Before(fourteenDaysAgo) {
            // Buffer: Don't spam them; ensure they haven't received a warning in the last 7 days!
            var recentNotif models.Notification
            sevenDaysAgo := time.Now().AddDate(0, 0, -7)
            errNotif := database.DB.Where("user_id = ? AND message LIKE ? AND created_at >= ?", 
                     user.ID, "%update%", sevenDaysAgo).First(&recentNotif).Error
            
            if errNotif != nil { 
                notification := models.Notification{
                    UserID:    user.ID,
                    Message:   "Reminder: You haven't updated your weight in 2 weeks! Log your progress to stay mathematically on-track.",
                    IsRead:    false,
                    CreatedAt: time.Now(),
                }
                database.DB.Create(&notification)
                log.Printf("Inactivity notification sent securely to User #%d\n", user.ID)
            }
		}
	}
}
