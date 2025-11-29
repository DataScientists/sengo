package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sheng-go-backend/config"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/user"
	"sheng-go-backend/pkg/infrastructure/datastore"
	"sheng-go-backend/pkg/util/auth"
)

type SeedUser struct {
	Email    string
	Name     string
	Age      int
	Password string
}

var seedUsers = []SeedUser{
	{
		Email:    "admin@example.com",
		Name:     "Admin User",
		Age:      30,
		Password: "admin12345",
	},
}

func main() {
	// Parse command line flags
	env := flag.String("env", "", "Environment (development, test, e2e, staging, production)")
	truncate := flag.Bool("truncate", false, "Truncate data (delete all users)")
	flag.Parse()

	// Set environment if provided via flag, otherwise rely on APP_ENV or default
	if *env != "" {
		os.Setenv("APP_ENV", *env)
	}

	// Initialize config
	config.ReadConfig(config.ReadConfigOption{})
	log.Printf("Starting seed tool for environment: %s", config.C.AppEnv)

	// Initialize database client
	client, err := datastore.NewClient()
	if err != nil {
		log.Fatalf("Failed to create database client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Handle truncation
	if *truncate {
		if err := truncateData(ctx, client); err != nil {
			log.Fatalf("Failed to truncate data: %v", err)
		}
		log.Println("Truncation completed successfully!")
		return
	}

	// Run seeds
	if err := seedUsersData(ctx, client); err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}

	log.Println("Seeding completed successfully!")
}

func truncateData(ctx context.Context, client *ent.Client) error {
	log.Println("Truncating users table...")
	// Delete all users
	_, err := client.User.Delete().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete users: %w", err)
	}
	return nil
}

func seedUsersData(ctx context.Context, client *ent.Client) error {
	log.Println("Seeding users...")

	for _, u := range seedUsers {
		// Check if user already exists
		exists, err := client.User.Query().
			Where(user.EmailEQ(u.Email)).
			Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed to check if user exists: %w", err)
		}

		if exists {
			log.Printf("User %s already exists, skipping", u.Email)
			continue
		}

		// Hash password using auth package (Argon2)
		hashedPassword, err := auth.HashPassword(u.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password for user %s: %w", u.Email, err)
		}

		// Create user
		_, err = client.User.Create().
			SetEmail(u.Email).
			SetName(u.Name).
			SetAge(u.Age).
			SetPassword(hashedPassword).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", u.Email, err)
		}

		log.Printf("Created user: %s", u.Email)
	}

	return nil
}
