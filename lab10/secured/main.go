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
	Password string `json:"password"`
}

var db *sql.DB

func connectDB() {
	var err error
	connStr := "postgres://postgres:password@db:5432/vulnerable_db?sslmode=disable"

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

	query := "SELECT id, username, password FROM users WHERE username = $1 AND password = $2"

	fmt.Println("Executing Secure Query:", query)
	fmt.Println("Params:", creds.Username, creds.Password)
	rows, err := db.Query(query, creds.Username, creds.Password)
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
			"debug_query": "Parameterized: " + query,
		})
	} else {
		// Return 401 so frontend knows login failed
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func main() {
	connectDB()

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/login", loginHandler)

	fmt.Println("Secured Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
