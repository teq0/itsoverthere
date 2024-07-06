// Service to redirect HTTP requests to addresses retrieved from a PostgreSQL database based on the Host header

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func main() {
	DB = connectDB()
	defer DB.Close()

	// Start the HTTP server
	http.HandleFunc("/", redirect)
	http.ListenAndServe(":8080", nil)
	http.ListenAndServe(":443", nil)
	// Wait forever
	select {}
}

// connectDB connects to the database
func connectDB() *sql.DB {
	// Get the database connection parameters
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Connect to the database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	return db
}

// redirect redirects the HTTP request to the address retrieved from the database
func redirect(w http.ResponseWriter, r *http.Request) {
	// Get the address from the database
	var host = r.Host
	var address string
	err := DB.QueryRow("SELECT address FROM redirects WHERE host = $1", host).Scan(&address)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Failed to get the address from the database: %v", err)
		return
	}

	// Redirect the request
	http.Redirect(w, r, address, http.StatusTemporaryRedirect)
	log.Printf("Redirected %s to %s", r.Host, address)
}
