package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Global variables
var DB *sql.DB
var MC *memcache.Client

func Start() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	DB, err = sql.Open("postgres", os.Getenv("POSTGRES_URL"))

	DB.SetMaxOpenConns(1000)

	if err != nil {
		log.Println(err)
	}

	err = DB.Ping()
	if err != nil {
		log.Println(err)
	}

	// Start memcache
	MC = memcache.New("127.0.0.1:11211")

	err = MC.Ping()
	if err != nil {
		log.Println("Error connecting to memcache", err)
	} else {
		log.Println("Memcache successfully connected!")
	}

	if os.Getenv("CLEARMEMCACHED") == "true" {
		MC.DeleteAll()
	}

	log.Println("Databases Successfully connected!")
}
