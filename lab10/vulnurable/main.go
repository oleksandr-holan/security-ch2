package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"` // Returning password to demonstrate dumping
}

var db *sql.DB

func connectDB() {
	var err error
	connStr := "postgres://postgres:password@db:5432/vulnerable_db?sslmode=disable"

	// Retry logic to wait for Postgres to start in Docker
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				fmt.Println("Successfully connected to the database!")
				return
			}
		}
		fmt.Println("Waiting for database...")
		time.Sleep(2 * time.Second)
	}
	log.Fatal("Could not connect to database:", err)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// !!! VULNERABLE CODE START !!!
	// NEVER DO THIS IN PRODUCTION. This allows SQL Injection.
	// We are directly formatting the string with user input.
	query := fmt.Sprintf("SELECT id, username, password FROM users WHERE username = '%s' AND password = '%s'", creds.Username, creds.Password)

	fmt.Println("Executing Query:", query) // Log query to console to see the injection
	// !!! VULNERABLE CODE END !!!

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []UserResponse
	for rows.Next() {
		var u UserResponse
		if err := rows.Scan(&u.ID, &u.Username, &u.Password); err != nil {
			continue
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	if len(users) > 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "success",
			"data":        users,
			"debug_query": query, // Returning the query to help visualize the lab
		})
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func main() {
	connectDB()

	// Serve static files (Frontend)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// API Endpoint
	http.HandleFunc("/login", loginHandler)

	fmt.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
