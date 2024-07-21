// Service to redirect HTTP requests to addresses retrieved from a PostgreSQL database based on the Host header

package main

import (
	"database/sql"
	"log"
	"net/http"
	"unicode"

	_ "github.com/lib/pq"
	sites "github.com/teq0/itsoverthere/pkg/sites"
)

var DB *sql.DB

func main() {

	// Start the HTTP server
	http.HandleFunc("/", redirect)
	http.ListenAndServe(":8080", nil)
	http.ListenAndServe(":443", nil)
	// Wait forever
	select {}
}

// redirect redirects the HTTP request to the address retrieved from the database
func redirect(w http.ResponseWriter, r *http.Request) {
	// Get the address from the database
	host := r.Host

	// if host is numeric then return 200, it's probably a health check
	if unicode.IsDigit(rune(host[0])) {
		log.Printf("Ignoring %s", host)
		w.WriteHeader(http.StatusOK)
		return
	}

	redirectURL, err := sites.GetRedirect(host)
	if err != nil {
		// Return BadRequest if the address is not found
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("No redirect found for " + host))
		log.Printf("Failed to get redirect: %v", err)
		return
	}

	// Redirect the request
	log.Printf("Redirecting %s to %s", r.Host, redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}
