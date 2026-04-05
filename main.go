package main

import (
	"log"
	"net/http"

	"um6p_fit_backend/database"
	"um6p_fit_backend/handlers"
	"um6p_fit_backend/middleware"
)

func main() {
	// Initialize Database
	database.InitDB()
	seedFile := "exercises_seed.json"
	database.SeedDBIfEmpty(seedFile)

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

	// Static Web Frontend Serving
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	log.Println("Starting server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
