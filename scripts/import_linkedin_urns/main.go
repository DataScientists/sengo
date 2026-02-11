package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sheng-go-backend/ent"
	"sheng-go-backend/ent/profileentry"
	"sheng-go-backend/ent/schema/ulid"
	"sheng-go-backend/pkg/const/globalid"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type ImportStats struct {
	Total     int
	Inserted  int
	Skipped   int
	Failed    int
	StartTime time.Time
}

func main() {
	// Parse command line flags
	csvFile := flag.String(
		"file",
		"DataScientistsLinkedInUserNames20260209.csv",
		"Path to CSV file containing LinkedIn usernames",
	)
	dbURL := flag.String(
		"db",
		"postgresql://root:root@localhost:5433/sheng_db?sslmode=disable",
		"Database connection string",
	)
	batchSize := flag.Int("batch", 1000, "Batch size for inserts")
	dryRun := flag.Bool("dry-run", false, "Preview without inserting data")
	flag.Parse()

	log.Printf("Starting LinkedIn URN import from: %s", *csvFile)
	log.Printf("Dry run mode: %v", *dryRun)

	// Connect to database
	client, err := ent.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("failed connecting to database: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Import profiles
	stats := &ImportStats{
		StartTime: time.Now(),
	}

	if err := importLinkedInURNs(ctx, client, *csvFile, *batchSize, *dryRun, stats); err != nil {
		log.Fatalf("failed importing URNs: %v", err)
	}

	// Print statistics
	duration := time.Since(stats.StartTime)
	log.Printf("\n=== Import Complete ===")
	log.Printf("Total processed: %d", stats.Total)
	log.Printf("Successfully inserted: %d", stats.Inserted)
	log.Printf("Skipped (duplicates): %d", stats.Skipped)
	log.Printf("Failed: %d", stats.Failed)
	log.Printf("Duration: %v", duration)
	if stats.Total > 0 {
		log.Printf("Average: %.2f records/sec", float64(stats.Total)/duration.Seconds())
	}
}

func importLinkedInURNs(
	ctx context.Context,
	client *ent.Client,
	filename string,
	batchSize int,
	dryRun bool,
	stats *ImportStats,
) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	batch := make([]*ent.ProfileEntryCreate, 0, batchSize)
	urnsToInsert := make([]string, 0, batchSize)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		username := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if username == "" {
			continue
		}

		stats.Total++

		if dryRun {
			if stats.Total <= 10 {
				log.Printf("[DRY RUN] Would insert: %s", username)
			}
			stats.Inserted++
			continue
		}

		// Check if URN already exists
		exists, err := client.ProfileEntry.Query().
			Where(profileentry.LinkedinUrnEQ(username)).
			Exist(ctx)
		if err != nil {
			log.Printf("Error checking URN existence at line %d (%s): %v", lineNum, username, err)
			stats.Failed++
			continue
		}

		if exists {
			stats.Skipped++
			if stats.Total%100 == 0 {
				log.Printf("Progress: %d processed (%d inserted, %d skipped, %d failed)",
					stats.Total, stats.Inserted, stats.Skipped, stats.Failed)
			}
			continue
		}

		// Generate new ULID
		newID := ulid.MustNew(globalid.New().ProfileEntry.Prefix)

		// Add to batch
		profileEntryCreate := client.ProfileEntry.Create().
			SetID(newID).
			SetLinkedinUrn(username).
			SetStatus(profileentry.StatusPending)

		batch = append(batch, profileEntryCreate)
		urnsToInsert = append(urnsToInsert, username)

		// Execute batch when full
		if len(batch) >= batchSize {
			if err := executeBatch(ctx, batch, urnsToInsert, stats); err != nil {
				log.Printf("Failed to execute batch: %v", err)
			}
			log.Printf("Progress: %d processed (%d inserted, %d skipped, %d failed)",
				stats.Total, stats.Inserted, stats.Skipped, stats.Failed)
			batch = batch[:0]
			urnsToInsert = urnsToInsert[:0]
		}
	}

	// Execute remaining batch
	if !dryRun && len(batch) > 0 {
		if err := executeBatch(ctx, batch, urnsToInsert, stats); err != nil {
			log.Printf("Failed to execute final batch: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	return nil
}

func executeBatch(
	ctx context.Context,
	batch []*ent.ProfileEntryCreate,
	urns []string,
	stats *ImportStats,
) error {
	for i, create := range batch {
		_, err := create.Save(ctx)
		if err != nil {
			// Check if it's a duplicate error
			if strings.Contains(err.Error(), "duplicate") ||
				strings.Contains(err.Error(), "unique constraint") {
				stats.Skipped++
			} else {
				log.Printf("Failed to create profile entry (%s): %v", urns[i], err)
				stats.Failed++
			}
		} else {
			stats.Inserted++
		}
	}
	return nil
}
