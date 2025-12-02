package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alextreichler/crochetbyjuliette/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	addUserCmd := flag.NewFlagSet("add-user", flag.ExitOnError)
	username := addUserCmd.String("username", "", "Username for the new user")
	password := addUserCmd.String("password", "", "Password for the new user")

	if len(os.Args) < 2 {
		fmt.Println("expected 'add-user' subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add-user":
		addUserCmd.Parse(os.Args[2:])
		if *username == "" || *password == "" {
			fmt.Println("username and password are required")
			addUserCmd.PrintDefaults()
			os.Exit(1)
		}
		createUser(*username, *password)
	default:
		fmt.Println("expected 'add-user' subcommand")
		os.Exit(1)
	}
}

func createUser(username, password string) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./crochet.db"
	}

	db, err := store.NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	// Ensure table exists if running cli before server
	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to init schema: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	err = db.CreateUser(username, string(hashedPassword))
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("User '%s' created successfully.\n", username)
}
