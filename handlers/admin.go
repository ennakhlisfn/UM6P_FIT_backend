package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"um6p_fit_backend/database"
	"um6p_fit_backend/models"
	"um6p_fit_backend/utils"
)

// serveCachedJSON encapsulates the caching logic so the DB is protected from rapid dashboard refresh cycles
func serveCachedJSON(w http.ResponseWriter, cacheKey string, computeFunc func() (interface{}, error)) {
	if val, found := utils.GlobalCache.Get(cacheKey); found {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(val)
		return
	}

	data, err := computeFunc()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.GlobalCache.Set(cacheKey, data, 60) // 60 second dashboard flush TTL
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(data)
}

// ----------------- PHASE 1: SUMMARY STATS -----------------
func AdminGetTotalUsers(w http.ResponseWriter, r *http.Request) {
	serveCachedJSON(w, "total_users", func() (interface{}, error) {
		var count int64
		database.DB.Model(&models.User{}).Count(&count)

		var thisWeekCount int64
		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		database.DB.Model(&models.User{}).Where("created_at >= ?", sevenDaysAgo).Count(&thisWeekCount)

		growth := 0.0
		if count > 0 {
			growth = (float64(thisWeekCount) / float64(count)) * 100
		}

		return map[string]interface{}{
			"total":         count,
			"growthPercent": growth,
			"syncTime":      time.Now(),
		}, nil
	})
}

func AdminGetActiveToday(w http.ResponseWriter, r *http.Request) {
	serveCachedJSON(w, "active_today", func() (interface{}, error) {
		var count int64
		todayStart := time.Now().Truncate(24 * time.Hour)
		database.DB.Table("workouts").Select("COUNT(DISTINCT user_id)").Where("date >= ?", todayStart).Scan(&count)
		return map[string]interface{}{
			"activeUsersCount": count,
			"syncTime":         time.Now(),
		}, nil
	})
}

func AdminGetNewSignups(w http.ResponseWriter, r *http.Request) {
	serveCachedJSON(w, "new_signups", func() (interface{}, error) {
		var count int64
		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		database.DB.Model(&models.User{}).Where("created_at >= ?", sevenDaysAgo).Count(&count)
		return map[string]interface{}{
			"count":    count,
			"goal":     100,
			"syncTime": time.Now(),
		}, nil
	})
}

func AdminGetTotalWorkouts(w http.ResponseWriter, r *http.Request) {
	serveCachedJSON(w, "total_workouts", func() (interface{}, error) {
		var count int64
		database.DB.Model(&models.Workout{}).Count(&count)
		return map[string]interface{}{
			"totalWorkouts": count,
			"syncTime":      time.Now(),
		}, nil
	})
}

// ----------------- PHASE 2: USER GROWTH -----------------
func AdminGetUserGrowth(w http.ResponseWriter, r *http.Request) {
	serveCachedJSON(w, "user_growth", func() (interface{}, error) {
		type Result struct {
			Month string `json:"month"`
			Count int    `json:"count"`
		}
		var results []Result
		database.DB.Table("users").
			Select("TO_CHAR(created_at, 'Mon') as month, COUNT(id) as count").
			Group("TO_CHAR(created_at, 'Mon')").
			Scan(&results)

		return results, nil
	})
}

// ----------------- PHASE 3: ACTIVE USERS HOURLY -----------------
func AdminGetActiveUsersHourly(w http.ResponseWriter, r *http.Request) {
	serveCachedJSON(w, "active_users_hourly", func() (interface{}, error) {
		type Result struct {
			Hour  float64 `json:"hour"`
			Count int     `json:"count"`
		}
		var results []Result
		database.DB.Table("workouts").
			Select("EXTRACT(HOUR FROM date) as hour, COUNT(id) as count").
			Group("EXTRACT(HOUR FROM date)").
			Order("hour ASC").
			Scan(&results)

		var final [24]int
		for _, res := range results {
			if res.Hour >= 0 && res.Hour < 24 {
				final[int(res.Hour)] = res.Count
			}
		}

		type Output struct {
			Time  int `json:"time"`
			Count int `json:"count"`
		}
		var out []Output
		for i := 0; i < 24; i++ {
			out = append(out, Output{Time: i, Count: final[i]})
		}
		return out, nil
	})
}

// ----------------- PHASE 4: POPULAR EXERCISES -----------------
func AdminGetPopularExercises(w http.ResponseWriter, r *http.Request) {
	serveCachedJSON(w, "popular_exercises", func() (interface{}, error) {
		type Result struct {
			ExerciseId string `json:"exerciseId"`
			TotalLogs  int    `json:"totalLogs"`
		}
		var results []Result
		database.DB.Table("workout_exercises").
			Select("exercise_id, COUNT(id) as total_logs").
			Group("exercise_id").
			Order("total_logs DESC").
			Limit(10).
			Scan(&results)

		return results, nil
	})
}

// ----------------- PHASE 5: AVG WORKOUTS -----------------
func AdminGetAvgWorkouts(w http.ResponseWriter, r *http.Request) {
	serveCachedJSON(w, "avg_workouts", func() (interface{}, error) {
		var totalUsers int64
		var totalWorkouts int64
		database.DB.Table("users").Count(&totalUsers)

		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		database.DB.Table("workouts").Where("date >= ?", sevenDaysAgo).Count(&totalWorkouts)

		avg := 0.0
		if totalUsers > 0 {
			avg = float64(totalWorkouts) / float64(totalUsers)
		}
		return map[string]interface{}{
			"averageWeekly": avg,
			"syncTime":      time.Now(),
		}, nil
	})
}

// ----------------- PHASE 6: COMMUNITY & LIVE CLASSES -----------------
func AdminGetCommunityRank(w http.ResponseWriter, r *http.Request) {
	serveCachedJSON(w, "community_rank", func() (interface{}, error) {
		return map[string]interface{}{
			"communityScore": 9845,
			"globalRank":     12,
			"status":         "Top 10%",
			"syncTime":       time.Now(),
		}, nil
	})
}

func AdminGetLiveClasses(w http.ResponseWriter, r *http.Request) {
	serveCachedJSON(w, "live_classes", func() (interface{}, error) {
		return map[string]interface{}{
			"activeNow":      12,
			"scheduledToday": 45,
			"syncTime":       time.Now(),
		}, nil
	})
}
