// Service to redirect HTTP requests to addresses retrieved from a PostgreSQL database based on the Host header

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"unicode"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func main() {
	initDB()
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
	dbName := os.Getenv("DB_NAME")

	// Connect to the database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)

	log.Printf("Connection string: %s\n", connStr)

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

	// if host is numeric then return 200, it's probably a health check

	if unicode.IsDigit(rune(host[0])) {
		log.Printf("Ignoring %s", host)
		w.WriteHeader(http.StatusOK)
		return
	}

	var address string
	err := DB.QueryRow("SELECT address FROM redirects WHERE host = $1", host).Scan(&address)
	if err != nil {
		// Return BadRequest if the address is not found
		w.Write([]byte("No redirect found for " + host))
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Failed to get the address for %s from the database: %v", host, err)
		return
	}

	// Redirect the request
	log.Printf("Redirecting %s to %s", r.Host, address)
	http.Redirect(w, r, address, http.StatusTemporaryRedirect)
}

func initDB() {
	// Get the database connection parameters
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Connect to the default database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable", host, port, user, password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Check if the database exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '" + dbName + "');").Scan(&exists)

	if err != nil {
		log.Fatalf("Failed to check if the database exists: %v", err)
	}

	if !exists {
		// Create the database
		createDbSQL := `
		CREATE DATABASE ` + dbName + `
		WITH
		OWNER = postgres
		ENCODING = 'UTF8'
		LC_COLLATE = 'en_US.utf8'
		LC_CTYPE = 'en_US.utf8'
		LOCALE_PROVIDER = 'libc'
		TABLESPACE = pg_default
		CONNECTION LIMIT = -1
		IS_TEMPLATE = False;
	`
		_, err = db.Exec(createDbSQL)
		if err != nil {
			log.Fatalf("Failed to create the database: %v", err)
		}
	}

	connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
	db.Close()
	DB = connectDB()

	// Create the table
	createTableSQL := `
CREATE SEQUENCE IF NOT EXISTS public.redirects_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 2147483647
    CACHE 1;
    
-- Table: public.redirects

-- DROP TABLE IF EXISTS public.redirects;

CREATE TABLE IF NOT EXISTS public.redirects
(
    id integer NOT NULL DEFAULT nextval('redirects_id_seq'::regclass),
    host character varying COLLATE pg_catalog."default" NOT NULL,
    address character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT redirects_pkey PRIMARY KEY (id),
    CONSTRAINT redirects_host_key UNIQUE (host)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.redirects
    OWNER to postgres;

ALTER SEQUENCE public.redirects_id_seq
    OWNED BY public.redirects.id;

ALTER SEQUENCE public.redirects_id_seq
    OWNER TO postgres;
`
	_, err = DB.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create the table: %v", err)
	}

	// Insert some rows
	// check if they're there first

	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM redirects").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to count the number of rows in the table: %v", err)
	}

	if count == 0 {
		redirects := map[string]string{
			"opa.engdemo.me":          "https://www.openpolicyagent.org/docs/latest/",
			"sbom.engdemo.me":         "https://cyclonedx.org/capabilities/sbom/",
			"foo.itsoverthere.lol":    "https://www.youtube.com/watch?v=IdkCEioCp24",
			"foo.itsoverthere.info":   "https://www.youtube.com/watch?v=IdkCEioCp24",
			"chain.yeahsure.cloud":    "https://slsa.dev/spec/v1.0/provenance",
			"fish.someotherthing.xyz": "https://lumifish.info/english",
		}

		for host, address := range redirects {
			log.Printf("Inserting redirect: %s -> %s", host, address)

			insertSQL := `
			INSERT INTO public.redirects(
			host, address)
			VALUES ($1, $2);
		`
			_, err = DB.Exec(insertSQL, host, address)
			if err != nil {
				log.Fatalf("Failed to insert the default redirect: %v", err)
			}
		}
	}
}
