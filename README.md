# RESTful API Endpoints & Testing Guide

Make sure your Go server is running (`go run .`) before testing these in a separate terminal window!

## 1. Fetch the Exercise Dictionary
**Description:** Retrieves a paginated list (up to 50) of the exercises stored in your database.

```bash
curl http://localhost:8080/api/exercises
```

**Expected Result:** A large JSON array containing exercises like "ab_roller" or "barbell_bench_press" with their muscles, instructions, and equipment.

## 2. Create a New User
**Description:** Registers a user with their physical metrics so workouts can be attached to them.

```bash
curl -X POST http://localhost:8080/api/users \
-H "Content-Type: application/json" \
-d '{  "name": "Senna",  "email": "senna.test@example.com",  "age": 25,  "height": 180.5,  "weight": 80.2}'
```

**Expected Result:** A `201 Created` response returning your user data, importantly including their new database ID (e.g., "id": 1).

## 3. Log a New Workout
**Description:** Saves a workout and automatically links the specific sets, reps, and weights to both the User and the Exercises.

```bash
curl -X POST http://localhost:8080/api/workouts \
-H "Content-Type: application/json" \
-d '{
  "userId": 1,
  "name": "Morning Core Routine",
  "exercises": [
    {
      "exerciseId": "ab_roller",
      "sets": 3,
      "reps": 15,
      "weight": 0
    }
  ]
}'
```

**Expected Result:** A JSON response confirming the workout was saved, assigning a `workoutId` to your specific sets and reps.

## 4. Fetch a User's Workout History
**Description:** Pulls every workout a specific user has done. Thanks to GORM's Preload, it deeply nests all the sets, reps, and full exercise descriptions inside the response.

```bash
curl http://localhost:8080/api/users/1/workouts
```

**Expected Result:** A comprehensive JSON array. You will see the **"Morning Core Routine"** workout, and inside of it, the exact details of the **"ab_roller"** exercise they performed.
