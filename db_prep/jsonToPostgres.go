// Script that can be used to convert https://github.com/RepDB/exercise-dataset to Postgres DB
package main

import (
	"fmt"
	"log"
	"os"

	"encoding/json"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DB connection details
const (
	host     = "localhost"
	port     = 5432
	user     = "workout_app_user"
	password = "super_secure_password"
	dbname   = "workout_app_db"
)

var schema = `
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

CREATE TABLE exercises_categories (
    id SERIAL PRIMARY KEY,
    exercise_type TEXT NOT NULL
);

CREATE TABLE exercises (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    category_id INT REFERENCES exercises_categories(id),
    equipment TEXT,
    primary_muscle TEXT,
    secondary_muscles TEXT[], 
    image_start_url TEXT,
    image_peak_url TEXT
);

CREATE TABLE workouts (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    workout_name TEXT NOT NULL
);

CREATE TABLE workout_junction (
    id SERIAL PRIMARY KEY,
    workout_id INT REFERENCES workouts(id) ON DELETE CASCADE,
    exercise_id INT REFERENCES exercises(id) ON DELETE CASCADE
);

CREATE TABLE logs (
    id SERIAL PRIMARY KEY,
    junction_id INT REFERENCES workout_junction(id) ON DELETE CASCADE,
    set_number INT NOT NULL,
    weight DOUBLE PRECISION,
    reps INT,
    time_recorded TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

// Required to access nested JSON data like exercises.name_en
type ExerciseWrapper struct {
	Exercises []Exercise `json:"exercises"`
}

type Exercise struct {
	Name             string   `json:"name_en"`
	Description      string   `json:"description_en"`
	Category         string   `json:"category"`
	Equipment        string   `json:"equipment"`
	PrimaryMuscles   []string `json:"primary_muscles"`
	SecondaryMuscles []string `json:"secondary_muscles"`
}

func main() {
	file, err := os.Open("exercises.json")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	db, err := connectToDb(host, port, user, password, dbname)

	if err != nil {
		log.Fatal(err)
	}

	prepareDatabase(db)
	defer db.Close()

	var data ExerciseWrapper

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)

	if err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)
	}

	// tx := db.MustBegin()

	if len(data.Exercises) > 0 {
		fmt.Printf("Successfully loaded %d exercises.\n", len(data.Exercises))
		fmt.Println("Uploading decoded data to Postgres")

		for i := 0; i < len(data.Exercises); i++ {
			// tx.MustExec("INSERT INTO exercises (name, description, category_id, equipment, primary_muscle, secondary_muscle) VALUES ($1, $2, $3, $4, $5, $6)", data.Exercises[i].Name, data.Exercises[i].Description, data.Exercises[i].Category, data.Exercises[i].Equipment, data.Exercises[i].PrimaryMuscles, data.Exercises[i].SecondaryMuscles)
		}
	}

	// tx.Commit()

}

func connectToDb(host string, port int, user string, password string, dbname string) (*sqlx.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// SQLX opens and pings the DB
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func prepareDatabase(db *sqlx.DB) {
	db.MustExec(schema)

	tx := db.MustBegin()
	tx.MustExec("INSERT INTO exercises_categories (exercise_type) VALUES ($1), ($2), ($3)", "strength", "cardio", "stretching")
	tx.Commit()
}
