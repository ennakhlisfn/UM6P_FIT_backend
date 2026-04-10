package main

import (
	"log"
	"net/http"

	"um6p_fit_backend/database"
	"um6p_fit_backend/handlers"
	"um6p_fit_backend/jobs"
	"um6p_fit_backend/middleware"
)

func main() {
	// Initialize Database
	database.InitDB()
	seedFile := "exercises_seed.json"
	database.SeedDBIfEmpty(seedFile)

	// Boot background scheduling jobs
	jobs.StartCronJobs()

	// User Routes
	http.HandleFunc("POST /api/users", handlers.CreateUser)
	http.HandleFunc("POST /api/login", handlers.LoginUser)

	// Exercise Routes
	http.HandleFunc("GET /api/exercises", handlers.GetExercises)

	// Workout Routes
	http.HandleFunc("POST /api/workouts", middleware.AuthMiddleware(handlers.CreateWorkout))
	http.HandleFunc("GET /api/users/{id}/workouts", middleware.AuthMiddleware(handlers.GetUserWorkouts))
	http.HandleFunc("PUT /api/workouts/{id}", middleware.AuthMiddleware(handlers.UpdateWorkout))
	http.HandleFunc("DELETE /api/workouts/{id}", middleware.AuthMiddleware(handlers.DeleteWorkout))
	http.HandleFunc("GET /api/users/{id}/exercises/{exId}/progress", middleware.AuthMiddleware(handlers.GetExerciseProgress))

	// Template Routes
	http.HandleFunc("POST /api/workout-templates", middleware.AuthMiddleware(handlers.CreateWorkoutTemplate))
	http.HandleFunc("GET /api/workout-templates", handlers.GetWorkoutTemplates)
	http.HandleFunc("PUT /api/workout-templates/{id}", middleware.AuthMiddleware(handlers.UpdateWorkoutTemplate))
	http.HandleFunc("DELETE /api/workout-templates/{id}", middleware.AuthMiddleware(handlers.DeleteWorkoutTemplate))

	// Leaderboard Routes
	http.HandleFunc("GET /api/leaderboard", handlers.GetLeaderboard)

	// Program Routes
	http.HandleFunc("POST /api/programs", middleware.AuthMiddleware(handlers.CreateWorkoutProgram))
	http.HandleFunc("GET /api/programs", handlers.GetWorkoutPrograms)
	http.HandleFunc("PUT /api/programs/{id}", middleware.AuthMiddleware(handlers.UpdateWorkoutProgram))
	http.HandleFunc("DELETE /api/programs/{id}", middleware.AuthMiddleware(handlers.DeleteWorkoutProgram))
	http.HandleFunc("POST /api/programs/{id}/start", middleware.AuthMiddleware(handlers.StartProgram))
	http.HandleFunc("POST /api/programs/advance", middleware.AuthMiddleware(handlers.AdvanceProgramDay))
	http.HandleFunc("GET /api/users/{id}/programs-history", handlers.GetProgramHistory)

	// Admin / Dashboard Routes
	http.HandleFunc("GET /api/admin/stats/total-users", middleware.AdminMiddleware(handlers.AdminGetTotalUsers))
	http.HandleFunc("GET /api/admin/stats/active-today", middleware.AdminMiddleware(handlers.AdminGetActiveToday))
	http.HandleFunc("GET /api/admin/stats/new-signups", middleware.AdminMiddleware(handlers.AdminGetNewSignups))
	http.HandleFunc("GET /api/admin/stats/total-workouts", middleware.AdminMiddleware(handlers.AdminGetTotalWorkouts))
	http.HandleFunc("GET /api/admin/charts/user-growth", middleware.AdminMiddleware(handlers.AdminGetUserGrowth))
	http.HandleFunc("GET /api/admin/charts/active-users-hourly", middleware.AdminMiddleware(handlers.AdminGetActiveUsersHourly))
	http.HandleFunc("GET /api/admin/stats/popular-exercises", middleware.AdminMiddleware(handlers.AdminGetPopularExercises))
	http.HandleFunc("GET /api/admin/stats/avg-workouts", middleware.AdminMiddleware(handlers.AdminGetAvgWorkouts))
	http.HandleFunc("GET /api/admin/stats/community-rank", middleware.AdminMiddleware(handlers.AdminGetCommunityRank))
	http.HandleFunc("GET /api/admin/live-classes/active", middleware.AdminMiddleware(handlers.AdminGetLiveClasses))

	// User Profile & Notification Routes
	http.HandleFunc("PUT /api/users/{id}/weight", middleware.AuthMiddleware(handlers.UpdateUserWeight))
	http.HandleFunc("GET /api/users/{id}/notifications", middleware.AuthMiddleware(handlers.GetUserNotifications))
	http.HandleFunc("PUT /api/notifications/{id}/read", middleware.AuthMiddleware(handlers.MarkNotificationRead))

	// Static Web Frontend Serving
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	log.Println("Starting server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
