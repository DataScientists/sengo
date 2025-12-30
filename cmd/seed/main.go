package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sheng-go-backend/config"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/profileentry"
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
	profileFile := flag.String(
		"profiles-file",
		"test-data/profiles.sql",
		"Path to profiles SQL dump",
	)
	profileLimit := flag.Int(
		"profiles-limit",
		1000,
		"Number of profile entries to seed from the dump",
	)
	profileGender := flag.String(
		"profiles-gender",
		"Male",
		"Gender to set for seeded profile entries",
	)
	skipProfiles := flag.Bool("skip-profiles", false, "Skip seeding profile entries")
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
	}

	// Run seeds
	if err := seedUsersData(ctx, client); err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}

	if !*skipProfiles {
		if err := seedProfileEntries(ctx, client, *profileFile, *profileLimit, *profileGender); err != nil {
			log.Fatalf("Failed to seed profile entries: %v", err)
		}
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

	log.Println("Truncating profiles table...")
	// Delete all users
	if _, err := client.Profile.Delete().Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete users: %w", err)
	}

	log.Println("Truncating job execution table...")
	// Delete all job execution histories
	if _, err := client.JobExecutionHistory.Delete().Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete job execution history: %w", err)
	}

	log.Println("Truncating profile entries table...")
	if _, err := client.ProfileEntry.Delete().Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete profile entries: %w", err)
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

func seedProfileEntries(
	ctx context.Context,
	client *ent.Client,
	sqlPath string,
	limit int,
	gender string,
) error {
	log.Printf("Seeding profile entries from %s (limit=%d)...", sqlPath, limit)

	urns, err := loadURNs(sqlPath, limit)
	if err != nil {
		return fmt.Errorf("failed to load URNs: %w", err)
	}

	seeded := 0
	for _, urn := range urns {
		exists, err := client.ProfileEntry.Query().Where(profileentry.LinkedinUrnEQ(urn)).Exist(ctx)
		if err != nil {
			return fmt.Errorf("failed checking urn %s: %w", urn, err)
		}
		if exists {
			log.Printf("Profile entry for URN %s already exists, skipping", urn)
			continue
		}

		if _, err := client.ProfileEntry.
			Create().
			SetLinkedinUrn(urn).
			SetGender(gender).
			Save(ctx); err != nil {
			return fmt.Errorf("failed inserting urn %s: %w", urn, err)
		}

		seeded++
	}

	log.Printf("Seeded %d profile entries (gender=%s)", seeded, gender)
	return nil
}

func loadURNs(path string, limit int) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`'(ACo[^']+)'`)
	matches := re.FindAllStringSubmatch(string(data), -1)

	seen := make(map[string]struct{})
	var urns []string

	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		urn := m[1]
		if _, ok := seen[urn]; ok {
			continue
		}
		seen[urn] = struct{}{}
		urns = append(urns, urn)
		if limit > 0 && len(urns) >= limit {
			break
		}
	}

	if limit > 0 && len(urns) > limit {
		urns = urns[:limit]
	}

	if len(urns) == 0 {
		abs := path
		if a, err := filepath.Abs(path); err == nil {
			abs = a
		}
		return nil, fmt.Errorf("no URNs found in %s", abs)
	}

	return urns, nil
}
